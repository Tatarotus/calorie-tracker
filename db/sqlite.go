package db

import (
	"database/sql"
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
	return nil
}

func (db *DB) AddFoodEntry(entry models.FoodEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO food_entries (timestamp, description, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?)",
		entry.Timestamp.UTC(), entry.Description, entry.Calories, entry.Protein, entry.Carbs, entry.Fat,
	)
	return err
}

func (db *DB) AddWaterEntry(entry models.WaterEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO water_entries (timestamp, amount_ml) VALUES (?, ?)",
		entry.Timestamp.UTC(), entry.AmountML,
	)
	return err
}

func (db *DB) GetDailyFoodEntries(t time.Time) ([]models.FoodEntry, error) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	rows, err := db.conn.Query(
		"SELECT id, timestamp, description, calories, protein, carbs, fat FROM food_entries WHERE timestamp >= ? AND timestamp < ?",
		start, end,
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
		// Try parsing different common formats or use driver's direct Scan if possible
		// The format found was "2006-01-02 15:04:05.999999999 +0000 UTC"
		parsedTs, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", ts)
		if err != nil {
			// Fallback if needed
			parsedTs, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", ts)
		}
		e.Timestamp = parsedTs
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetDailyWaterEntries(t time.Time) ([]models.WaterEntry, error) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	rows, err := db.conn.Query(
		"SELECT id, timestamp, amount_ml FROM water_entries WHERE timestamp >= ? AND timestamp < ?",
		start, end,
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
		parsedTs, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", ts)
		if err != nil {
			parsedTs, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", ts)
		}
		e.Timestamp = parsedTs
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetStatsRange(days int) ([]models.DailyStats, error) {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -days)

	query := `
		SELECT 
			strftime('%Y-%m-%d', timestamp) as date,
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
			strftime('%Y-%m-%d', timestamp) as date,
			0 as calories,
			0 as protein,
			0 as carbs,
			0 as fat,
			SUM(amount_ml) as water_ml
		FROM water_entries
		WHERE timestamp >= ?
		GROUP BY date
	`
	
	rows, err := db.conn.Query(query, start, start)
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
		if !date.Valid {
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
	start := time.Now().UTC().AddDate(0, 0, -days)
	rows, err := db.conn.Query(
		"SELECT id, timestamp, description, calories, protein, carbs, fat FROM food_entries WHERE timestamp >= ?",
		start,
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
		// Try parsing different common formats or use driver's direct Scan if possible
		// The format found was "2006-01-02 15:04:05.999999999 +0000 UTC"
		parsedTs, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", ts)
		if err != nil {
			// Fallback if needed
			parsedTs, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", ts)
		}
		e.Timestamp = parsedTs
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *DB) GetCachedFood(description string) (*models.FoodEntry, error) {
	description = strings.ToLower(strings.TrimSpace(description))
	var e models.FoodEntry
	err := db.conn.QueryRow(
		"SELECT calories, protein, carbs, fat FROM food_cache WHERE description = ?",
		description,
	).Scan(&e.Calories, &e.Protein, &e.Carbs, &e.Fat)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Description = description
	return &e, nil
}

func (db *DB) CacheFood(entry models.FoodEntry) error {
	description := strings.ToLower(strings.TrimSpace(entry.Description))
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO food_cache (description, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?)",
		description, entry.Calories, entry.Protein, entry.Carbs, entry.Fat,
	)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
