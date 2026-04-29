package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// Check if food_cache table needs migration
	needsFoodCacheMigration, err := db.tableNeedsMigration("food_cache", []string{"base_quantity", "unit"})
	if err != nil {
		return fmt.Errorf("checking food_cache migration: %w", err)
	}

	if needsFoodCacheMigration {
		// Create a new table with the correct schema
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

		// Copy data from old table if it exists
		_, _ = db.conn.Exec(`
		INSERT OR IGNORE INTO food_cache_new (description, calories, protein, carbs, fat)
		SELECT description, calories, protein, carbs, fat FROM food_cache
	`)

		// Drop old table and rename new one
		if _, err := db.conn.Exec(`DROP TABLE IF EXISTS food_cache`); err != nil {
			return fmt.Errorf("dropping old food_cache: %w", err)
		}
		if _, err := db.conn.Exec(`ALTER TABLE food_cache_new RENAME TO food_cache`); err != nil {
			return fmt.Errorf("renaming food_cache_new: %w", err)
		}
	}

	// Check if reference_foods table needs migration
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

// tableNeedsMigration checks if a table has all the required columns
func (db *DB) tableNeedsMigration(tableName string, requiredColumns []string) (bool, error) {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		// Table doesn't exist, no migration needed (it will be created)
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
	// Try a very permissive parse if it contains a date-like string
	if len(ts) >= 10 {
		if t, err := time.Parse("2006-01-02", ts[:10]); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (db *DB) AddFoodEntry(entry models.FoodEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO food_entries (timestamp, description, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?)",
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.Description, entry.Calories, entry.Protein, entry.Carbs, entry.Fat,
	)
	return err
}

func (db *DB) AddWaterEntry(entry models.WaterEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO water_entries (timestamp, amount_ml) VALUES (?, ?)",
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.AmountML,
	)
	return err
}

func (db *DB) GetDailyFoodEntries(t time.Time) ([]models.FoodEntry, error) {
	// Ensure t is evaluated in its local timezone context
	t = t.Local()
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	// Convert it to UTC to match stored 'Z' timestamps
	startUTC := startOfDay.UTC()
	endUTC := startUTC.Add(24 * time.Hour)

	rows, err := db.conn.Query(
		"SELECT id, timestamp, description, calories, protein, carbs, fat FROM food_entries WHERE timestamp >= ? AND timestamp < ?",
		startUTC.Format(time.RFC3339Nano),
		endUTC.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.FoodEntry
	for rows.Next() {
		var e models.FoodEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.Description, &e.Calories, &e.Protein, &e.Carbs, &e.Fat); err != nil {
			return nil, err
		}
		e.Timestamp = parseTimestamp(ts).Local()
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetDailyWaterEntries(t time.Time) ([]models.WaterEntry, error) {
	t = t.Local()
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	startUTC := startOfDay.UTC()
	endUTC := startUTC.Add(24 * time.Hour)

	rows, err := db.conn.Query(
		"SELECT id, timestamp, amount_ml FROM water_entries WHERE timestamp >= ? AND timestamp < ?",
		startUTC.Format(time.RFC3339Nano),
		endUTC.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.WaterEntry
	for rows.Next() {
		var e models.WaterEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.AmountML); err != nil {
			return nil, err
		}
		e.Timestamp = parseTimestamp(ts).Local()
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetStatsRange(days int) ([]models.DailyStats, error) {
	// We use the last N days relative to now local
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStartUTC := todayStart.AddDate(0, 0, -days).UTC()

	// For strftime to work reliably, we use RFC3339 which is ISO8601
	// We want to group by the LOCAL date, so we tell SQLite to convert to local time
	// if it can, but since our format has Z or timezone, SQLite handles it.
	query := `
		SELECT 
			strftime('%Y-%m-%d', datetime(timestamp, 'localtime')) as date,
			SUM(calories) as calories,
			SUM(protein) as protein,
			SUM(carbs) as carbs,
			SUM(fat) as fat,
			0 as water_ml
		FROM food_entries
		WHERE timestamp >= ?
		GROUP BY date
		UNION ALL
		SELECT 
			strftime('%Y-%m-%d', datetime(timestamp, 'localtime')) as date,
			0 as calories,
			0 as protein,
			0 as carbs,
			0 as fat,
			SUM(amount_ml) as water_ml
		FROM water_entries
		WHERE timestamp >= ?
		GROUP BY date
	`

	rows, err := db.conn.Query(query,
		rangeStartUTC.Format(time.RFC3339Nano),
		rangeStartUTC.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsMap := make(map[string]*models.DailyStats)
	for rows.Next() {
		var date sql.NullString
		var cal, pro, carb, fat, water float64
		if err := rows.Scan(&date, &cal, &pro, &carb, &fat, &water); err != nil {
			return nil, err
		}
		if !date.Valid || date.String == "" {
			continue
		}
		if _, ok := statsMap[date.String]; !ok {
			statsMap[date.String] = &models.DailyStats{Date: date.String}
		}
		statsMap[date.String].Calories += cal
		statsMap[date.String].Protein += pro
		statsMap[date.String].Carbs += carb
		statsMap[date.String].Fat += fat
		statsMap[date.String].WaterML += water
	}

	var result []models.DailyStats
	for _, s := range statsMap {
		result = append(result, *s)
	}
	return result, nil
}

func (db *DB) GetFoodEntriesRange(days int) ([]models.FoodEntry, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStartUTC := todayStart.AddDate(0, 0, -days).UTC()

	rows, err := db.conn.Query(
		"SELECT id, timestamp, description, calories, protein, carbs, fat FROM food_entries WHERE timestamp >= ? ORDER BY timestamp DESC",
		rangeStartUTC.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.FoodEntry
	for rows.Next() {
		var e models.FoodEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.Description, &e.Calories, &e.Protein, &e.Carbs, &e.Fat); err != nil {
			return nil, err
		}
		e.Timestamp = parseTimestamp(ts).Local()
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetWaterEntriesRange(days int) ([]models.WaterEntry, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStartUTC := todayStart.AddDate(0, 0, -days).UTC()

	rows, err := db.conn.Query(
		"SELECT id, timestamp, amount_ml FROM water_entries WHERE timestamp >= ? ORDER BY timestamp DESC",
		rangeStartUTC.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.WaterEntry
	for rows.Next() {
		var e models.WaterEntry
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.AmountML); err != nil {
			return nil, err
		}
		e.Timestamp = parseTimestamp(ts).Local()
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetCachedFood(description string) (*models.ReferenceFood, error) {
	description = strings.ToLower(strings.TrimSpace(description))
	var f models.ReferenceFood
	err := db.conn.QueryRow(
		"SELECT base_quantity, unit, calories, protein, carbs, fat FROM food_cache WHERE description = ?",
		description,
	).Scan(&f.BaseQuantity, &f.Unit, &f.Macros.Calories, &f.Macros.Protein, &f.Macros.Carbs, &f.Macros.Fat)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.Name = description
	return &f, nil
}

func (db *DB) GetReferenceFood(name string) (*models.ReferenceFood, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	var f models.ReferenceFood
	err := db.conn.QueryRow(
		"SELECT name, base_quantity, unit, calories, protein, carbs, fat FROM reference_foods WHERE name = ? OR ? LIKE '%' || name || '%'",
		name, name,
	).Scan(&f.Name, &f.BaseQuantity, &f.Unit, &f.Macros.Calories, &f.Macros.Protein, &f.Macros.Carbs, &f.Macros.Fat)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (db *DB) CacheFood(f models.ReferenceFood) error {
	description := strings.ToLower(strings.TrimSpace(f.Name))
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO food_cache (description, base_quantity, unit, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?, ?)",
		description, f.BaseQuantity, f.Unit, f.Macros.Calories, f.Macros.Protein, f.Macros.Carbs, f.Macros.Fat,
	)
	return err
}

func (db *DB) GetAllCacheEntries() ([]models.ReferenceFood, error) {
	rows, err := db.conn.Query("SELECT description, base_quantity, unit, calories, protein, carbs, fat FROM food_cache")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.ReferenceFood
	for rows.Next() {
		var f models.ReferenceFood
		if err := rows.Scan(&f.Name, &f.BaseQuantity, &f.Unit, &f.Macros.Calories, &f.Macros.Protein, &f.Macros.Carbs, &f.Macros.Fat); err != nil {
			return nil, err
		}
		entries = append(entries, f)
	}
	return entries, nil
}

func (db *DB) SetGoal(goal models.Goal) error {
	_, err := db.conn.Exec(
		"INSERT INTO goals (timestamp, description) VALUES (?, ?)",
		goal.Timestamp.Format(time.RFC3339Nano),
		goal.Description,
	)
	return err
}

func (db *DB) GetLatestGoal() (*models.Goal, error) {
	var g models.Goal
	var ts string
	err := db.conn.QueryRow(
		"SELECT id, timestamp, description FROM goals ORDER BY timestamp DESC LIMIT 1",
	).Scan(&g.ID, &ts, &g.Description)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.Timestamp = parseTimestamp(ts)
	return &g, nil
}

func (db *DB) RemoveLastEntry() error {
	var entryType string
	var id int64
	var ts string

	query := `
		SELECT 'food' as type, id, timestamp FROM food_entries
		UNION ALL
		SELECT 'water' as type, id, timestamp FROM water_entries
		ORDER BY timestamp DESC
		LIMIT 1
	`
	err := db.conn.QueryRow(query).Scan(&entryType, &id, &ts)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	var deleteQuery string
	if entryType == "food" {
		deleteQuery = "DELETE FROM food_entries WHERE id = ?"
	} else {
		deleteQuery = "DELETE FROM water_entries WHERE id = ?"
	}

	_, err = db.conn.Exec(deleteQuery, id)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
