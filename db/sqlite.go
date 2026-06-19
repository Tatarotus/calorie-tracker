package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"calorie-tracker/models"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func NewDB() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(home, ".calorie_tracker.db")
	return NewTestDB(dbPath)
}

func NewTestDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if dbPath == ":memory:" {
		conn.SetMaxOpenConns(1)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS food_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			description TEXT,
			calories REAL,
			protein REAL,
			carbs REAL,
			fat REAL,
			original_query TEXT,
			normalized_query TEXT,
			canonical_key TEXT,
			resolution_trace TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS water_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			amount_ml REAL
		)`,
		`CREATE TABLE IF NOT EXISTS food_cache (
			description TEXT PRIMARY KEY,
			base_quantity REAL,
			unit TEXT,
			calories REAL,
			protein REAL,
			carbs REAL,
			fat REAL
		)`,
		`CREATE TABLE IF NOT EXISTS goals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reference_foods (
			name TEXT PRIMARY KEY,
			base_quantity REAL,
			unit TEXT,
			calories REAL,
			protein REAL,
			carbs REAL,
			fat REAL
		)`,
		`CREATE TABLE IF NOT EXISTS canonical_foods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			canonical_name TEXT UNIQUE,
			normalized_name TEXT,
			aliases_json TEXT,
			language TEXT,
			category TEXT,
			food_type TEXT,
			composition_hints TEXT,
			default_serving_amount REAL,
			default_serving_unit TEXT,
			density_multiplier REAL,
			grams_per_ml REAL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS nutrition_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			canonical_food_id INTEGER,
			serving_amount REAL,
			serving_unit TEXT,
			calories REAL,
			protein REAL,
			carbs REAL,
			fat REAL,
			fiber REAL,
			source_type TEXT,
			source_confidence REAL,
			source_reference TEXT,
			resolution_method TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY(canonical_food_id) REFERENCES canonical_foods(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_overrides (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			canonical_food_id INTEGER,
			serving_amount REAL,
			serving_unit TEXT,
			calories REAL,
			protein REAL,
			carbs REAL,
			fat REAL,
			fiber REAL,
			override_reason TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY(canonical_food_id) REFERENCES canonical_foods(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_canonical_foods_normalized_name ON canonical_foods(normalized_name)`,
		`CREATE INDEX IF NOT EXISTS idx_nutrition_cache_canonical_food_id ON nutrition_cache(canonical_food_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nutrition_cache_source_type ON nutrition_cache(source_type)`,
		`CREATE INDEX IF NOT EXISTS idx_nutrition_cache_updated_at ON nutrition_cache(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_user_overrides_canonical_food_id ON user_overrides(canonical_food_id)`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return err
		}
	}

	if err := db.migrateExistingTables(); err != nil {
		return err
	}

	if err := db.seedReferenceFoods(); err != nil {
		return err
	}

	return db.seedCanonicalFoods()
}

func (db *DB) migrateExistingTables() error {
	needsFoodEntriesMigration, err := db.tableNeedsMigration("food_entries", []string{"original_query", "normalized_query", "canonical_key", "resolution_trace"})
	if err != nil {
		return fmt.Errorf("checking food_entries migration: %w", err)
	}
	if needsFoodEntriesMigration {
		_, _ = db.conn.Exec(`ALTER TABLE food_entries ADD COLUMN original_query TEXT`)
		_, _ = db.conn.Exec(`ALTER TABLE food_entries ADD COLUMN normalized_query TEXT`)
		_, _ = db.conn.Exec(`ALTER TABLE food_entries ADD COLUMN canonical_key TEXT`)
		_, _ = db.conn.Exec(`ALTER TABLE food_entries ADD COLUMN resolution_trace TEXT`)
	}

	needsFoodCacheMigration, err := db.tableNeedsMigration("food_cache", []string{"base_quantity", "unit"})
	if err != nil {
		return fmt.Errorf("checking food_cache migration: %w", err)
	}

	if needsFoodCacheMigration {
		if _, err = db.conn.Exec(`
			CREATE TABLE IF NOT EXISTS food_cache_new (
				description TEXT PRIMARY KEY,
				base_quantity REAL,
				unit TEXT,
				calories REAL,
				protein REAL,
				carbs REAL,
				fat REAL
			)
		`); err != nil {
			return fmt.Errorf("creating food_cache_new: %w", err)
		}

		_, _ = db.conn.Exec(`
			INSERT OR IGNORE INTO food_cache_new (description, calories, protein, carbs, fat)
			SELECT description, calories, protein, carbs, fat FROM food_cache
		`)

		if _, err = db.conn.Exec(`DROP TABLE IF EXISTS food_cache`); err != nil {
			return fmt.Errorf("dropping old food_cache: %w", err)
		}
		if _, err = db.conn.Exec(`ALTER TABLE food_cache_new RENAME TO food_cache`); err != nil {
			return fmt.Errorf("renaming food_cache_new: %w", err)
		}
	}

	needsRefFoodsMigration, err := db.tableNeedsMigration("reference_foods", []string{"base_quantity", "unit"})
	if err != nil {
		return fmt.Errorf("checking reference_foods migration: %w", err)
	}

	if needsRefFoodsMigration {
		if _, err = db.conn.Exec(`
			CREATE TABLE IF NOT EXISTS reference_foods_new (
				name TEXT PRIMARY KEY,
				base_quantity REAL,
				unit TEXT,
				calories REAL,
				protein REAL,
				carbs REAL,
				fat REAL
			)
		`); err != nil {
			return fmt.Errorf("creating reference_foods_new: %w", err)
		}

		_, _ = db.conn.Exec(`
			INSERT OR IGNORE INTO reference_foods_new (name, calories, protein, carbs, fat)
			SELECT name, calories, protein, carbs, fat FROM reference_foods
		`)

		if _, err = db.conn.Exec(`DROP TABLE IF EXISTS reference_foods`); err != nil {
			return fmt.Errorf("dropping old reference_foods: %w", err)
		}
		if _, err = db.conn.Exec(`ALTER TABLE reference_foods_new RENAME TO reference_foods`); err != nil {
			return fmt.Errorf("renaming reference_foods_new: %w", err)
		}
	}

	needsCanonicalMigration, err := db.tableNeedsMigration("canonical_foods", []string{"food_type", "composition_hints"})
	if err != nil {
		return fmt.Errorf("checking canonical_foods migration: %w", err)
	}
	if needsCanonicalMigration {
		_, _ = db.conn.Exec(`ALTER TABLE canonical_foods ADD COLUMN food_type TEXT`)
		_, _ = db.conn.Exec(`ALTER TABLE canonical_foods ADD COLUMN composition_hints TEXT`)
	}

	// Trigger the custom migration of legacy food_cache entries to canonical_foods & nutrition_cache
	if err := db.migrateLegacyCache(); err != nil {
		return fmt.Errorf("migrating legacy food cache: %w", err)
	}

	return nil
}

func (db *DB) tableNeedsMigration(tableName string, requiredColumns []string) (bool, error) {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, nil
	}
	defer func() { _ = rows.Close() }()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		columns[name] = true
	}

	// If no columns found, table doesn't exist - no migration needed
	if len(columns) == 0 {
		return false, nil
	}

	for _, col := range requiredColumns {
		if !columns[col] {
			return true, nil
		}
	}

	return false, nil
}

func (db *DB) seedReferenceFoods() error {
	foods := []models.ReferenceFood{
		{Name: "arroz branco", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 130, Protein: 2.7, Carbs: 28, Fat: 0.3}},
		{Name: "white rice", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 130, Protein: 2.7, Carbs: 28, Fat: 0.3}},
		{Name: "frango grelhado", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 165, Protein: 31, Carbs: 0, Fat: 3.6}},
		{Name: "grilled chicken", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 165, Protein: 31, Carbs: 0, Fat: 3.6}},
		{Name: "chicken breast", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 165, Protein: 31, Carbs: 0, Fat: 3.6}},
		{Name: "ovo", BaseQuantity: 1, Unit: "unit", Macros: models.Macros{Calories: 70, Protein: 6, Carbs: 0.6, Fat: 5}},
		{Name: "egg", BaseQuantity: 1, Unit: "unit", Macros: models.Macros{Calories: 70, Protein: 6, Carbs: 0.6, Fat: 5}},
		{Name: "banana", BaseQuantity: 1, Unit: "unit", Macros: models.Macros{Calories: 89, Protein: 1.1, Carbs: 23, Fat: 0.3}},
		{Name: "olive oil", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 884, Protein: 0, Carbs: 0, Fat: 100}},
		{Name: "azeite", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 884, Protein: 0, Carbs: 0, Fat: 100}},
		{Name: "butter", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 717, Protein: 0.9, Carbs: 0.1, Fat: 81}},
		{Name: "manteiga", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 717, Protein: 0.9, Carbs: 0.1, Fat: 81}},
		{Name: "bread", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 265, Protein: 9, Carbs: 49, Fat: 3.2}},
		{Name: "pao", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 265, Protein: 9, Carbs: 49, Fat: 3.2}},
		{Name: "pao de forma", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 265, Protein: 9, Carbs: 49, Fat: 3.2}},
		{Name: "sandwich bread", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 265, Protein: 9, Carbs: 49, Fat: 3.2}},
		{Name: "pizza", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 266, Protein: 11, Carbs: 33, Fat: 10}},
		{Name: "pizza mussarela", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 266, Protein: 11, Carbs: 33, Fat: 10}},
		{Name: "pizza calabresa", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 292, Protein: 12, Carbs: 30, Fat: 14}},
		{Name: "abacaxi", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 50, Protein: 0.5, Carbs: 13, Fat: 0.1}},
		{Name: "pineapple", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 50, Protein: 0.5, Carbs: 13, Fat: 0.1}},
		{Name: "mamao", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 45, Protein: 0.5, Carbs: 11, Fat: 0.1}},
		{Name: "papaya", BaseQuantity: 100, Unit: "gram", Macros: models.Macros{Calories: 45, Protein: 0.5, Carbs: 11, Fat: 0.1}},
	}

	for _, f := range foods {
		_, err := db.conn.Exec(
			"INSERT OR IGNORE INTO reference_foods (name, base_quantity, unit, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?, ?)",
			f.Name, f.BaseQuantity, f.Unit, f.Macros.Calories, f.Macros.Protein, f.Macros.Carbs, f.Macros.Fat,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTimestamp(ts string) time.Time {
	formats := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, ts); err == nil {
			return t
		}
	}
	if len(ts) >= 10 {
		if t, err := time.Parse("2006-01-02", ts[:10]); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (db *DB) seedCanonicalFoods() error {
	now := time.Now()
	// Seed known canonical foods
	canonicalFoods := []struct {
		name      string
		normName  string
		aliases   string
		cat       string
		foodType  models.FoodType
		compHints string
		amount    float64
		unit      string
	}{
		{"cafe_com_leite", "café com leite", `["café com leite", "cafe com leite", "coffee with milk"]`, "beverage", models.FoodTypeBeverage, "milk", 200, "ml"},
		{"ovo", "ovo", `["ovo", "egg", "boiled egg", "egg fried"]`, "protein", models.FoodTypeProtein, "egg", 1, "unit"},
		{"banana", "banana", `["banana", "banana da terra", "banana prata"]`, "fruit", models.FoodTypeFruit, "fruit", 1, "unit"},
		{"arroz_branco", "arroz branco", `["arroz branco", "white rice", "arroz"]`, "grain", models.FoodTypeGrain, "carb", 100, "gram"},
		{"frango_grelhado", "frango grelhado", `["frango grelhado", "grilled chicken", "peito de frango"]`, "protein", models.FoodTypeProtein, "protein", 100, "gram"},
		{"azeite", "azeite de oliva", `["azeite", "azeite de oliva", "olive oil"]`, "fat", models.FoodTypeComposite, "fat", 13, "ml"},
		{"pao", "pão", `["pão", "pao", "bread", "pão francês"]`, "grain", models.FoodTypeGrain, "carb", 50, "gram"},
		{"pao_de_forma", "pão de forma", `["pão de forma", "pao de forma", "sandwich bread", "sliced bread"]`, "grain", models.FoodTypeGrain, "carb", 25, "gram"},
		{"cafe", "café", `["café", "cafe", "black coffee"]`, "beverage", models.FoodTypeBeverage, "coffee", 50, "ml"},
		{"pizza", "pizza", `["pizza", "pizza mussarela", "pizza calabresa", "pizza portuguesa"]`, "composite", models.FoodTypeComposite, "carb+protein+fat", 120, "gram"},
		{"abacaxi", "abacaxi", `["abacaxi", "pineapple", "abacaxi havaiano"]`, "fruit", models.FoodTypeFruit, "fruit", 900, "gram"},
		{"mamao", "mamão", `["mamão", "mamao", "mamao formosa", "papaya", "mamão formosa"]`, "fruit", models.FoodTypeFruit, "fruit", 1500, "gram"},
	}

	for _, cf := range canonicalFoods {
		var id int64
		err := db.conn.QueryRow("SELECT id FROM canonical_foods WHERE canonical_name = ?", cf.name).Scan(&id)
		if err == sql.ErrNoRows {
			_, err = db.conn.Exec(`
				INSERT INTO canonical_foods (
					canonical_name, normalized_name, aliases_json, language, category, food_type, composition_hints,
					default_serving_amount, default_serving_unit, density_multiplier, grams_per_ml,
					created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1.0, 1.0, ?, ?)`,
				cf.name, cf.normName, cf.aliases, "en", cf.cat, string(cf.foodType), cf.compHints,
				cf.amount, cf.unit, now, now,
			)
			if err != nil {
				return fmt.Errorf("seeding canonical food %s: %w", cf.name, err)
			}
		}
	}
	return nil
}
