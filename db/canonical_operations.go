package db

import (
	"database/sql"
	"strings"
	"time"

	"calorie-tracker/models"
)

// GetCanonicalFood gets a canonical food by ID
func (db *DB) GetCanonicalFood(id int64) (*models.CanonicalFood, error) {
	var f models.CanonicalFood
	var aliases string
	var createdAtStr, updatedAtStr string
	var foodType, compHints sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, canonical_name, normalized_name, aliases_json, language, category, food_type, composition_hints,
		       default_serving_amount, default_serving_unit, density_multiplier, grams_per_ml,
		       created_at, updated_at
		FROM canonical_foods WHERE id = ?`, id).Scan(
		&f.ID, &f.CanonicalName, &f.NormalizedName, &aliases, &f.Language, &f.Category, &foodType, &compHints,
		&f.DefaultServingAmount, &f.DefaultServingUnit, &f.DensityMultiplier, &f.GramsPerML,
		&createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.AliasesJSON = aliases
	if foodType.Valid {
		f.FoodType = foodType.String
	}
	if compHints.Valid {
		f.CompositionHints = compHints.String
	}
	f.CreatedAt = parseTimestamp(createdAtStr)
	f.UpdatedAt = parseTimestamp(updatedAtStr)
	return &f, nil
}

// GetCanonicalFoodByName gets a canonical food by its canonical_name
func (db *DB) GetCanonicalFoodByName(name string) (*models.CanonicalFood, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	var f models.CanonicalFood
	var aliases string
	var createdAtStr, updatedAtStr string
	var foodType, compHints sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, canonical_name, normalized_name, aliases_json, language, category, food_type, composition_hints,
		       default_serving_amount, default_serving_unit, density_multiplier, grams_per_ml,
		       created_at, updated_at
		FROM canonical_foods WHERE canonical_name = ?`, name).Scan(
		&f.ID, &f.CanonicalName, &f.NormalizedName, &aliases, &f.Language, &f.Category, &foodType, &compHints,
		&f.DefaultServingAmount, &f.DefaultServingUnit, &f.DensityMultiplier, &f.GramsPerML,
		&createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.AliasesJSON = aliases
	if foodType.Valid {
		f.FoodType = foodType.String
	}
	if compHints.Valid {
		f.CompositionHints = compHints.String
	}
	f.CreatedAt = parseTimestamp(createdAtStr)
	f.UpdatedAt = parseTimestamp(updatedAtStr)
	return &f, nil
}

// SaveCanonicalFood inserts or updates a canonical food
func (db *DB) SaveCanonicalFood(food *models.CanonicalFood) error {
	now := time.Now()
	food.UpdatedAt = now
	if food.CreatedAt.IsZero() {
		food.CreatedAt = now
	}

	if food.DensityMultiplier == 0 {
		food.DensityMultiplier = 1.0
	}
	if food.GramsPerML == 0 {
		food.GramsPerML = 1.0
	}

	if food.ID != 0 {
		_, err := db.conn.Exec(`
			UPDATE canonical_foods SET
				canonical_name = ?, normalized_name = ?, aliases_json = ?, language = ?, category = ?, food_type = ?, composition_hints = ?,
				default_serving_amount = ?, default_serving_unit = ?, density_multiplier = ?, grams_per_ml = ?,
				updated_at = ?
			WHERE id = ?`,
			food.CanonicalName, food.NormalizedName, food.AliasesJSON, food.Language, food.Category, food.FoodType, food.CompositionHints,
			food.DefaultServingAmount, food.DefaultServingUnit, food.DensityMultiplier, food.GramsPerML,
			food.UpdatedAt, food.ID,
		)
		return err
	}

	res, err := db.conn.Exec(`
		INSERT INTO canonical_foods (
			canonical_name, normalized_name, aliases_json, language, category, food_type, composition_hints,
			default_serving_amount, default_serving_unit, density_multiplier, grams_per_ml,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		food.CanonicalName, food.NormalizedName, food.AliasesJSON, food.Language, food.Category, food.FoodType, food.CompositionHints,
		food.DefaultServingAmount, food.DefaultServingUnit, food.DensityMultiplier, food.GramsPerML,
		food.CreatedAt, food.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	food.ID = id
	return nil
}

// GetAllCanonicalFoods retrieves all canonical foods from the database
func (db *DB) GetAllCanonicalFoods() ([]models.CanonicalFood, error) {
	rows, err := db.conn.Query(`
		SELECT id, canonical_name, normalized_name, aliases_json, language, category, food_type, composition_hints,
		       default_serving_amount, default_serving_unit, density_multiplier, grams_per_ml,
		       created_at, updated_at
		FROM canonical_foods`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []models.CanonicalFood
	for rows.Next() {
		var f models.CanonicalFood
		var aliases string
		var createdAtStr, updatedAtStr string
		var foodType, compHints sql.NullString
		err := rows.Scan(
			&f.ID, &f.CanonicalName, &f.NormalizedName, &aliases, &f.Language, &f.Category, &foodType, &compHints,
			&f.DefaultServingAmount, &f.DefaultServingUnit, &f.DensityMultiplier, &f.GramsPerML,
			&createdAtStr, &updatedAtStr,
		)
		if err != nil {
			return nil, err
		}
		f.AliasesJSON = aliases
		if foodType.Valid {
			f.FoodType = foodType.String
		}
		if compHints.Valid {
			f.CompositionHints = compHints.String
		}
		f.CreatedAt = parseTimestamp(createdAtStr)
		f.UpdatedAt = parseTimestamp(updatedAtStr)
		list = append(list, f)
	}
	return list, nil
}

// GetNutritionCache gets a cache entry for a food by unit
func (db *DB) GetNutritionCache(canonicalFoodID int64, unit string) (*models.NutritionCacheEntry, error) {
	unit = strings.ToLower(strings.TrimSpace(unit))
	var c models.NutritionCacheEntry
	var createdAtStr, updatedAtStr string
	err := db.conn.QueryRow(`
		SELECT id, canonical_food_id, serving_amount, serving_unit,
		       calories, protein, carbs, fat, fiber,
		       source_type, source_confidence, source_reference, resolution_method,
		       created_at, updated_at
		FROM nutrition_cache
		WHERE canonical_food_id = ? AND LOWER(serving_unit) = ?
		LIMIT 1`, canonicalFoodID, unit).Scan(
		&c.ID, &c.CanonicalFoodID, &c.ServingAmount, &c.ServingUnit,
		&c.Calories, &c.Protein, &c.Carbs, &c.Fat, &c.Fiber,
		&c.SourceType, &c.SourceConfidence, &c.SourceReference, &c.ResolutionMethod,
		&createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		err = db.conn.QueryRow(`
			SELECT id, canonical_food_id, serving_amount, serving_unit,
			       calories, protein, carbs, fat, fiber,
			       source_type, source_confidence, source_reference, resolution_method,
			       created_at, updated_at
			FROM nutrition_cache
			WHERE canonical_food_id = ?
			LIMIT 1`, canonicalFoodID).Scan(
			&c.ID, &c.CanonicalFoodID, &c.ServingAmount, &c.ServingUnit,
			&c.Calories, &c.Protein, &c.Carbs, &c.Fat, &c.Fiber,
			&c.SourceType, &c.SourceConfidence, &c.SourceReference, &c.ResolutionMethod,
			&createdAtStr, &updatedAtStr,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	c.CreatedAt = parseTimestamp(createdAtStr)
	c.UpdatedAt = parseTimestamp(updatedAtStr)
	return &c, nil
}

// SaveNutritionCache inserts or updates a nutrition cache entry
func (db *DB) SaveNutritionCache(entry *models.NutritionCacheEntry) error {
	now := time.Now()
	entry.UpdatedAt = now
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}

	if entry.ID != 0 {
		_, err := db.conn.Exec(`
			UPDATE nutrition_cache SET
				canonical_food_id = ?, serving_amount = ?, serving_unit = ?,
				calories = ?, protein = ?, carbs = ?, fat = ?, fiber = ?,
				source_type = ?, source_confidence = ?, source_reference = ?, resolution_method = ?,
				updated_at = ?
			WHERE id = ?`,
			entry.CanonicalFoodID, entry.ServingAmount, entry.ServingUnit,
			entry.Calories, entry.Protein, entry.Carbs, entry.Fat, entry.Fiber,
			entry.SourceType, entry.SourceConfidence, entry.SourceReference, entry.ResolutionMethod,
			entry.UpdatedAt, entry.ID,
		)
		return err
	}

	res, err := db.conn.Exec(`
		INSERT INTO nutrition_cache (
			canonical_food_id, serving_amount, serving_unit,
			calories, protein, carbs, fat, fiber,
			source_type, source_confidence, source_reference, resolution_method,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.CanonicalFoodID, entry.ServingAmount, entry.ServingUnit,
		entry.Calories, entry.Protein, entry.Carbs, entry.Fat, entry.Fiber,
		entry.SourceType, entry.SourceConfidence, entry.SourceReference, entry.ResolutionMethod,
		entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = id
	return nil
}

// GetUserOverride gets an override entry for a food
func (db *DB) GetUserOverride(canonicalFoodID int64) (*models.UserOverrideEntry, error) {
	var o models.UserOverrideEntry
	var createdAtStr, updatedAtStr string
	err := db.conn.QueryRow(`
		SELECT id, canonical_food_id, serving_amount, serving_unit,
		       calories, protein, carbs, fat, fiber, override_reason,
		       created_at, updated_at
		FROM user_overrides
		WHERE canonical_food_id = ?
		LIMIT 1`, canonicalFoodID).Scan(
		&o.ID, &o.CanonicalFoodID, &o.ServingAmount, &o.ServingUnit,
		&o.Calories, &o.Protein, &o.Carbs, &o.Fat, &o.Fiber, &o.OverrideReason,
		&createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.CreatedAt = parseTimestamp(createdAtStr)
	o.UpdatedAt = parseTimestamp(updatedAtStr)
	return &o, nil
}

// SaveUserOverride inserts or updates a user override entry
func (db *DB) SaveUserOverride(override *models.UserOverrideEntry) error {
	now := time.Now()
	override.UpdatedAt = now
	if override.CreatedAt.IsZero() {
		override.CreatedAt = now
	}

	if override.ID != 0 {
		_, err := db.conn.Exec(`
			UPDATE user_overrides SET
				canonical_food_id = ?, serving_amount = ?, serving_unit = ?,
				calories = ?, protein = ?, carbs = ?, fat = ?, fiber = ?, override_reason = ?,
				updated_at = ?
			WHERE id = ?`,
			override.CanonicalFoodID, override.ServingAmount, override.ServingUnit,
			override.Calories, override.Protein, override.Carbs, override.Fat, override.Fiber, override.OverrideReason,
			override.UpdatedAt, override.ID,
		)
		return err
	}

	res, err := db.conn.Exec(`
		INSERT INTO user_overrides (
			canonical_food_id, serving_amount, serving_unit,
			calories, protein, carbs, fat, fiber, override_reason,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		override.CanonicalFoodID, override.ServingAmount, override.ServingUnit,
		override.Calories, override.Protein, override.Carbs, override.Fat, override.Fiber, override.OverrideReason,
		override.CreatedAt, override.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	override.ID = id
	return nil
}
