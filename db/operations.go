package db

import (
	"database/sql"
	"strings"
	"time"

	"calorie-tracker/models"
)

// Food operations

func (db *DB) AddFoodEntry(entry models.FoodEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO food_entries (timestamp, description, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?)",
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.Description, entry.Calories, entry.Protein, entry.Carbs, entry.Fat,
	)
	return err
}

func (db *DB) GetDailyFoodEntries(t time.Time) ([]models.FoodEntry, error) {
	t = t.Local()
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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

func (db *DB) CacheFood(f models.ReferenceFood) error {
	description := strings.ToLower(strings.TrimSpace(f.Name))
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO food_cache (description, base_quantity, unit, calories, protein, carbs, fat) VALUES (?, ?, ?, ?, ?, ?, ?)",
		description, f.BaseQuantity, f.Unit, f.Macros.Calories, f.Macros.Protein, f.Macros.Carbs, f.Macros.Fat,
	)
	return err
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

func (db *DB) GetReferenceFood(name string) (*models.ReferenceFood, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	var f models.ReferenceFood
	err := db.conn.QueryRow(
		"SELECT name, base_quantity, unit, calories, protein, carbs, fat FROM reference_foods WHERE name = ?",
		name,
	).Scan(&f.Name, &f.BaseQuantity, &f.Unit, &f.Macros.Calories, &f.Macros.Protein, &f.Macros.Carbs, &f.Macros.Fat)

	if err == nil {
		return &f, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	err = db.conn.QueryRow(
		"SELECT name, base_quantity, unit, calories, protein, carbs, fat FROM reference_foods WHERE ? LIKE '%' || name || '%'",
		name,
	).Scan(&f.Name, &f.BaseQuantity, &f.Unit, &f.Macros.Calories, &f.Macros.Protein, &f.Macros.Carbs, &f.Macros.Fat)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Water operations

func (db *DB) AddWaterEntry(entry models.WaterEntry) error {
	_, err := db.conn.Exec(
		"INSERT INTO water_entries (timestamp, amount_ml) VALUES (?, ?)",
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.AmountML,
	)
	return err
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

// Stats operations

func (db *DB) GetStatsRange(days int) ([]models.DailyStats, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStartUTC := todayStart.AddDate(0, 0, -days).UTC()

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

// Goal operations

func (db *DB) SetGoal(goal models.Goal) error {
	_, err := db.conn.Exec(
		"INSERT INTO goals (timestamp, description) VALUES (?, ?)",
		goal.Timestamp.UTC().Format(time.RFC3339Nano),
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
	g.Timestamp = parseTimestamp(ts).Local()
	return &g, nil
}

// Other operations

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
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}
