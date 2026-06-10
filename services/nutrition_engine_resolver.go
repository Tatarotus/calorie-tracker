package services

import (
	"calorie-tracker/db"
	"calorie-tracker/models"
	"fmt"
	"strings"
)

type HybridNutritionResolver struct {
	db                db.DBProvider
	llm               *LLMService
	providers         []NutritionProvider
	fuzzyResolver     *FuzzyResolver
	ttlPolicy         *TTLPolicy
	cacheResolver     CacheResolver
	validator         *SemanticValidator
	canonicalResolver *CanonicalResolverService
	ProgressCallback  func(stage string)
}

func (r *HybridNutritionResolver) Resolve(canonical *models.CanonicalFood, item ParsedFoodItem) (*models.ReferenceFood, *models.ResolutionTrace, error) {
	trace := &models.ResolutionTrace{
		ParserUsed:   "llm",
		CanonicalKey: canonical.CanonicalName,
	}

	if r.ProgressCallback != nil {
		r.ProgressCallback("cache")
	}

	// 1. Cache-first: user override, then active local cache.
	if ref, ok, err := r.fastBypassCheck(canonical, item, trace); ok {
		return ref, trace, err
	}

	cacheEntry, _ := r.cacheResolver.Get(canonical.ID, item.Unit)
	if accepted, err := r.acceptCacheCandidate(canonical, cacheEntry, trace); accepted != nil || err != nil {
		if err != nil {
			return nil, trace, err
		}
		r.updateTraceAndCache(canonical, accepted, trace)
		if accepted.IsStale {
			go r.refreshCacheInBackground(canonical, item)
		}
		return accepted.ReferenceFood, trace, nil
	}
	if accepted, err := r.resolveExactReference(canonical, item); accepted != nil || err != nil {
		if err != nil {
			return nil, trace, err
		}
		r.updateTraceAndCache(canonical, accepted, trace)
		return accepted.ReferenceFood, trace, nil
	}
	if accepted, err := r.resolveFuzzyReference(canonical, item); accepted != nil || err != nil {
		if err != nil {
			return nil, trace, err
		}
		r.updateTraceAndCache(canonical, accepted, trace)
		return accepted.ReferenceFood, trace, nil
	}

	// 2. Run configured primary providers + SerpAPI concurrently and choose the best candidate
	accepted, err := r.resolveActiveProvidersAndCompare(canonical, item, trace)
	if err != nil {
		return nil, trace, err
	}
	if accepted != nil {
		r.updateTraceAndCache(canonical, accepted, trace)
		return accepted.ReferenceFood, trace, nil
	}

	// 3. Last resort: LLM estimate
	llmCandidate, err := r.resolveLLM(item)
	if err != nil {
		if trace != nil {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "llm fallback failed: "+err.Error())
		}
	}
	if llmCandidate != nil {
		ok, err := r.evaluateCandidate(canonical, llmCandidate)
		if err != nil {
			return nil, trace, err
		}
		if ok {
			r.updateTraceAndCache(canonical, llmCandidate, trace)
			return llmCandidate.ReferenceFood, trace, nil
		}
	}

	return nil, trace, nil
}

func (r *HybridNutritionResolver) resolveActiveProvidersAndCompare(canonical *models.CanonicalFood, item ParsedFoodItem, trace *models.ResolutionTrace) (*ResolutionCandidate, error) {
	var activeProviders []NutritionProvider
	for _, provider := range r.providers {
		if provider != nil {
			activeProviders = append(activeProviders, provider)
		}
	}

	if len(activeProviders) == 0 {
		return nil, nil
	}

	type providerResult struct {
		provider NutritionProvider
		ref      *models.ReferenceFood
		err      error
	}
	resChan := make(chan providerResult, len(activeProviders))
	for _, provider := range activeProviders {
		go func(p NutritionProvider) {
			if r.ProgressCallback != nil {
				providerType := fmt.Sprintf("%T", p)
				if strings.Contains(strings.ToLower(providerType), "fatsecret") {
					r.ProgressCallback("fatsecret")
				} else if strings.Contains(strings.ToLower(providerType), "calorieninjas") {
					r.ProgressCallback("calorieninjas")
				} else if strings.Contains(strings.ToLower(providerType), "serp") {
					r.ProgressCallback("serp")
				}
			}
			ref, err := p.ResolveFood(parsedFoodFromItem(item))
			resChan <- providerResult{provider: p, ref: ref, err: err}
		}(provider)
	}

	var validCandidates []*ResolutionCandidate

	for i := 0; i < len(activeProviders); i++ {
		res := <-resChan
		if res.err != nil {
			if trace != nil {
				trace.ValidationWarnings = append(trace.ValidationWarnings, fmt.Sprintf("%T lookup failed: %v", res.provider, res.err))
			}
			continue
		}

		if res.ref != nil {
			sourceType := "fatsecret"
			resolutionMethod := "fatsecret_api"
			confidence := 1.0

			if _, ok := res.provider.(*FatSecretProvider); ok {
				sourceType = "fatsecret"
				resolutionMethod = "fatsecret_api"
				confidence = 1.0
			} else if _, ok := res.provider.(*CalorieNinjasProvider); ok {
				sourceType = "calorieninjas"
				resolutionMethod = "calorieninjas_api"
				confidence = 0.95
			} else if _, ok := res.provider.(*SerpAPIProvider); ok {
				sourceType = "serpapi_fallback"
				resolutionMethod = "serpapi_fallback"
				confidence = 0.85
			}

			candidate := &ResolutionCandidate{
				ReferenceFood:    res.ref,
				Confidence:       confidence,
				SourceType:       sourceType,
				ResolutionMethod: resolutionMethod,
			}

			ok, evalErr := r.evaluateCandidate(canonical, candidate)
			if evalErr != nil {
				return nil, evalErr
			}
			if ok {
				validCandidates = append(validCandidates, candidate)
			} else {
				if trace != nil {
					trace.ValidationWarnings = append(trace.ValidationWarnings, fmt.Sprintf("%s response failed validation", sourceType))
				}
			}
		}
	}

	if len(validCandidates) == 0 {
		return nil, nil
	}

	best := r.evaluateBestCandidate(validCandidates)
	return best, nil
}

func (r *HybridNutritionResolver) resolveExactReference(canonical *models.CanonicalFood, item ParsedFoodItem) (*ResolutionCandidate, error) {
	ref, err := r.db.GetReferenceFood(item.FoodName)
	if err != nil || ref == nil {
		return nil, err
	}
	candidate := &ResolutionCandidate{
		ReferenceFood:    ref,
		Confidence:       0.95,
		SourceType:       "reference_db",
		ResolutionMethod: "exact_reference",
	}
	ok, err := r.evaluateCandidate(canonical, candidate)
	if err != nil || !ok {
		return nil, err
	}
	return candidate, nil
}

func (r *HybridNutritionResolver) resolveFuzzyReference(canonical *models.CanonicalFood, item ParsedFoodItem) (*ResolutionCandidate, error) {
	ref, err := r.fuzzyResolver.Find(item.FoodName)
	if err != nil || ref == nil {
		return nil, err
	}
	if hasPreparationMismatch(item.FoodName, ref.Name) {
		return nil, nil
	}
	candidate := &ResolutionCandidate{
		ReferenceFood:    ref,
		Confidence:       0.90,
		SourceType:       "reference_db",
		ResolutionMethod: "fuzzy_fallback",
	}
	ok, err := r.evaluateCandidate(canonical, candidate)
	if err != nil || !ok {
		return nil, err
	}
	return candidate, nil
}

func (r *HybridNutritionResolver) acceptCacheCandidate(canonical *models.CanonicalFood, cacheEntry *models.NutritionCacheEntry, trace *models.ResolutionTrace) (*ResolutionCandidate, error) {
	if cacheEntry == nil {
		return nil, nil
	}

	confidence := cacheEntry.SourceConfidence
	method := "local_cache"
	isStale := false
	if r.ttlPolicy.IsExpired(cacheEntry) {
		confidence = 0.55
		method = "stale_cache"
		isStale = true
	}

	candidate := &ResolutionCandidate{
		ReferenceFood: &models.ReferenceFood{
			Name:         canonical.NormalizedName,
			BaseQuantity: cacheEntry.ServingAmount,
			Unit:         cacheEntry.ServingUnit,
			Macros: models.Macros{
				Calories: cacheEntry.Calories,
				Protein:  cacheEntry.Protein,
				Carbs:    cacheEntry.Carbs,
				Fat:      cacheEntry.Fat,
			},
		},
		Confidence:       confidence,
		SourceType:       cacheEntry.SourceType,
		ResolutionMethod: method,
		IsStale:          isStale,
	}

	ok, err := r.evaluateCandidate(canonical, candidate)
	if err != nil || !ok {
		if trace != nil {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "cache response failed validation")
		}
		return nil, err
	}
	return candidate, nil
}

func (r *HybridNutritionResolver) resolveLLMWithSerpFallback(canonical *models.CanonicalFood, item ParsedFoodItem, trace *models.ResolutionTrace) (*ResolutionCandidate, error) {
	var llmCandidate *ResolutionCandidate
	var serpCandidate *ResolutionCandidate
	var err error

	llmCandidate, err = r.resolveLLM(item)
	if err != nil {
		if trace != nil {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "llm fallback failed: "+err.Error())
		}
	}

	if r.ProgressCallback != nil {
		r.ProgressCallback("serp")
	}

	serpCandidate, err = r.resolveSerpAPI(item)
	if err != nil {
		if trace != nil {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "serpapi fallback failed: "+err.Error())
		}
	}

	switch {
	case llmCandidate != nil && serpCandidate != nil:
		if macrosConsistent(llmCandidate.ReferenceFood, serpCandidate.ReferenceFood) {
			llmCandidate.Confidence = 0.75
			ok, err := r.evaluateCandidate(canonical, llmCandidate)
			if err != nil || !ok {
				return nil, err
			}
			trace.SerpAPIFallback = true
			trace.ValidationWarnings = append(trace.ValidationWarnings, "llm estimate cross-checked by serpapi")
			return llmCandidate, nil
		}
		trace.ValidationWarnings = append(trace.ValidationWarnings, "llm and serpapi disagreed; rejecting uncorroborated estimates")
		return nil, nil
	case serpCandidate != nil:
		if !hasUsefulMacroProfile(serpCandidate.ReferenceFood) {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "serpapi response rejected because macro profile was incomplete")
			return nil, nil
		}
		ok, err := r.evaluateCandidate(canonical, serpCandidate)
		if err != nil || !ok {
			return nil, err
		}
		return serpCandidate, nil
	case llmCandidate != nil:
		if r.hasSerpAPIProvider() {
			trace.ValidationWarnings = append(trace.ValidationWarnings, "llm estimate rejected because serpapi did not corroborate it")
			return nil, nil
		}
		llmCandidate.Confidence = 0.60
		ok, err := r.evaluateCandidate(canonical, llmCandidate)
		if err != nil || !ok {
			return nil, err
		}
		trace.ValidationWarnings = append(trace.ValidationWarnings, "llm estimate used without provider corroboration")
		return llmCandidate, nil
	default:
		return nil, nil
	}
}

func (r *HybridNutritionResolver) resolveLLM(item ParsedFoodItem) (*ResolutionCandidate, error) {
	if r.llm == nil {
		return nil, nil
	}
	ref, err := r.llm.ParseFood(item.FoodName)
	if err != nil || ref == nil {
		return nil, err
	}
	return &ResolutionCandidate{
		ReferenceFood:    ref,
		Confidence:       0.60,
		SourceType:       "llm",
		ResolutionMethod: "llm_fallback",
	}, nil
}

func (r *HybridNutritionResolver) resolveSerpAPI(item ParsedFoodItem) (*ResolutionCandidate, error) {
	for _, provider := range r.providers {
		if !isSerpAPIProvider(provider) {
			continue
		}
		ref, err := provider.ResolveFood(parsedFoodFromItem(item))
		if err != nil {
			return nil, err
		}
		if ref != nil {
			return &ResolutionCandidate{
				ReferenceFood:    ref,
				Confidence:       0.50,
				SourceType:       "serpapi_fallback",
				ResolutionMethod: "serpapi_fallback",
			}, nil
		}
	}
	return nil, nil
}

func (r *HybridNutritionResolver) evaluateCandidate(canonical *models.CanonicalFood, candidate *ResolutionCandidate) (bool, error) {
	if err := checkUnrealisticData(candidate.ReferenceFood); err != nil {
		return false, err
	}

	// Deterministic Category Compatibility Check
	queryType := models.FoodType(canonical.FoodType)
	matchedType := models.FoodType("")
	if r.canonicalResolver != nil && candidate.ReferenceFood != nil {
		matchedCanonical, err := r.canonicalResolver.Resolve(candidate.ReferenceFood.Name)
		if err == nil && matchedCanonical != nil {
			matchedType = models.FoodType(matchedCanonical.FoodType)
		}
	}

	shouldVal := false
	if candidate.SourceType == "user_override" || (candidate.SourceType == "fatsecret" && candidate.Confidence >= 1.0) {
		shouldVal = false
	} else if candidate.Confidence < 0.90 || candidate.SourceType == "serpapi_fallback" || candidate.SourceType == "fuzzy_fallback" || candidate.SourceType == "reference_db" || candidate.SourceType == "llm" {
		shouldVal = true
	}

	if shouldVal && r.validator != nil {
		valid, _ := r.validator.Validate(queryType, matchedType, candidate.ReferenceFood)
		if !valid {
			return false, nil
		}
	}

	return true, nil
}

func (r *HybridNutritionResolver) updateTraceAndCache(canonical *models.CanonicalFood, accepted *ResolutionCandidate, trace *models.ResolutionTrace) {
	trace.ResolutionMethod = accepted.ResolutionMethod
	trace.SourceType = accepted.SourceType
	trace.SourceConfidence = accepted.Confidence

	if accepted.IsStale {
		trace.StaleCacheUsed = true
	}

	if accepted.ResolutionMethod == "local_cache" {
		trace.CacheHit = true
		trace.ValidationTriggered = false
	} else {
		trace.ValidationTriggered = true
		queryType := models.FoodType(canonical.FoodType)
		matchedType := models.FoodType("")
		if r.canonicalResolver != nil {
			matchedCanonical, err := r.canonicalResolver.Resolve(accepted.ReferenceFood.Name)
			if err == nil && matchedCanonical != nil {
				matchedType = models.FoodType(matchedCanonical.FoodType)
			}
		}
		if r.validator != nil {
			_, warnings := r.validator.Validate(queryType, matchedType, accepted.ReferenceFood)
			trace.ValidationWarnings = append(trace.ValidationWarnings, warnings...)
		}
		trace.ValidationResult = "VALID"
	}

	// Save to local cache database if from dynamic external source
	if accepted.SourceType == "fatsecret" && accepted.ResolutionMethod == "fatsecret_api" {
		trace.FatSecretQueried = true
		cacheEntry := &models.NutritionCacheEntry{
			CanonicalFoodID:  canonical.ID,
			ServingAmount:    accepted.ReferenceFood.BaseQuantity,
			ServingUnit:      accepted.ReferenceFood.Unit,
			Calories:         accepted.ReferenceFood.Macros.Calories,
			Protein:          accepted.ReferenceFood.Macros.Protein,
			Carbs:            accepted.ReferenceFood.Macros.Carbs,
			Fat:              accepted.ReferenceFood.Macros.Fat,
			Fiber:            0.0,
			SourceType:       "fatsecret",
			SourceConfidence: 1.0,
			SourceReference:  "FatSecret API: " + accepted.ReferenceFood.Name,
			ResolutionMethod: "fatsecret_api",
		}
		_ = r.db.SaveNutritionCache(cacheEntry)
		_ = r.db.CacheFood(*accepted.ReferenceFood)
	} else if accepted.SourceType == "calorieninjas" && accepted.ResolutionMethod == "calorieninjas_api" {
		cacheEntry := &models.NutritionCacheEntry{
			CanonicalFoodID:  canonical.ID,
			ServingAmount:    accepted.ReferenceFood.BaseQuantity,
			ServingUnit:      accepted.ReferenceFood.Unit,
			Calories:         accepted.ReferenceFood.Macros.Calories,
			Protein:          accepted.ReferenceFood.Macros.Protein,
			Carbs:            accepted.ReferenceFood.Macros.Carbs,
			Fat:              accepted.ReferenceFood.Macros.Fat,
			Fiber:            0.0,
			SourceType:       "calorieninjas",
			SourceConfidence: 0.95,
			SourceReference:  "CalorieNinjas: " + accepted.ReferenceFood.Name,
			ResolutionMethod: "calorieninjas_api",
		}
		_ = r.db.SaveNutritionCache(cacheEntry)
		_ = r.db.CacheFood(*accepted.ReferenceFood)
	} else if accepted.SourceType == "serpapi_fallback" {
		trace.SerpAPIFallback = true
		cacheEntry := &models.NutritionCacheEntry{
			CanonicalFoodID:  canonical.ID,
			ServingAmount:    accepted.ReferenceFood.BaseQuantity,
			ServingUnit:      accepted.ReferenceFood.Unit,
			Calories:         accepted.ReferenceFood.Macros.Calories,
			Protein:          accepted.ReferenceFood.Macros.Protein,
			Carbs:            accepted.ReferenceFood.Macros.Carbs,
			Fat:              accepted.ReferenceFood.Macros.Fat,
			Fiber:            0.0,
			SourceType:       "serpapi_fallback",
			SourceConfidence: 0.5,
			SourceReference:  "SerpAPI Fallback: " + accepted.ReferenceFood.Name,
			ResolutionMethod: "serpapi_fallback",
		}
		_ = r.db.SaveNutritionCache(cacheEntry)
		_ = r.db.CacheFood(*accepted.ReferenceFood)
	}
}

func getNormalizedMacros(ref *models.ReferenceFood) models.Macros {
	if ref == nil {
		return models.Macros{}
	}
	if ref.BaseQuantity <= 0 {
		return ref.Macros
	}
	factor := 100.0 / ref.BaseQuantity
	return models.Macros{
		Calories: ref.Macros.Calories * factor,
		Protein:  ref.Macros.Protein * factor,
		Carbs:    ref.Macros.Carbs * factor,
		Fat:      ref.Macros.Fat * factor,
	}
}

func (r *HybridNutritionResolver) evaluateBestCandidate(candidates []*ResolutionCandidate) *ResolutionCandidate {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	type candidateScore struct {
		cand  *ResolutionCandidate
		score float64
	}

	scores := make([]candidateScore, len(candidates))
	for i, c1 := range candidates {
		m1 := getNormalizedMacros(c1.ReferenceFood)
		var diffSum float64
		for j, c2 := range candidates {
			if i == j {
				continue
			}
			m2 := getNormalizedMacros(c2.ReferenceFood)
			
			calDiff := abs(m1.Calories - m2.Calories)
			proDiff := abs(m1.Protein - m2.Protein)
			carbDiff := abs(m1.Carbs - m2.Carbs)
			fatDiff := abs(m1.Fat - m2.Fat)
			
			diffSum += calDiff + 4.0*(proDiff+carbDiff+fatDiff)
		}
		scores[i] = candidateScore{
			cand:  c1,
			score: diffSum / float64(len(candidates)-1),
		}
	}

	bestIdx := 0
	minScore := scores[0].score

	for i := 1; i < len(scores); i++ {
		if scores[i].score < minScore-5.0 {
			minScore = scores[i].score
			bestIdx = i
		} else if abs(scores[i].score-minScore) <= 5.0 {
			if scores[i].cand.Confidence > scores[bestIdx].cand.Confidence {
				minScore = scores[i].score
				bestIdx = i
			}
		}
	}

	return scores[bestIdx].cand
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
