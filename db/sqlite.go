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
			fat REAL
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
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return err
		}
	}

	if err := db.migrateExistingTables(); err != nil {
		return err
	}

	return db.seedReferenceFoods()
}

func (db *DB) migrateExistingTables() error {
	needsFoodCacheMigration, err := db.tableNeedsMigration("food_cache", []string{"base_quantity", "unit"})
	if err != nil {
		return fmt.Errorf("checking food_cache migration: %w", err)
	}

	if needsFoodCacheMigration {
		if _, err := db.conn.Exec(`
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

		if _, err := db.conn.Exec(`DROP TABLE IF EXISTS food_cache`); err != nil {
			return fmt.Errorf("dropping old food_cache: %w", err)
		}
		if _, err := db.conn.Exec(`ALTER TABLE food_cache_new RENAME TO food_cache`); err != nil {
			return fmt.Errorf("renaming food_cache_new: %w", err)
		}
	}

	needsRefFoodsMigration, err := db.tableNeedsMigration("reference_foods", []string{"base_quantity", "unit"})
	if err != nil {
		return fmt.Errorf("checking reference_foods migration: %w", err)
	}

	if needsRefFoodsMigration {
		if _, err := db.conn.Exec(`
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

		if _, err := db.conn.Exec(`DROP TABLE IF EXISTS reference_foods`); err != nil {
			return fmt.Errorf("dropping old reference_foods: %w", err)
		}
		if _, err := db.conn.Exec(`ALTER TABLE reference_foods_new RENAME TO reference_foods`); err != nil {
			return fmt.Errorf("renaming reference_foods_new: %w", err)
		}
	}

	return nil
}

func (db *DB) tableNeedsMigration(tableName string, requiredColumns []string) (bool, error) {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, nil
	}
	defer rows.Close()

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
