/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package binance

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// HistoricalDataManager manages fetching and caching of historical candles
type HistoricalDataManager struct {
	db            *sql.DB
	binanceClient *BinanceClient
}

// NewHistoricalDataManager creates a new data manager
func NewHistoricalDataManager(db *sql.DB) *HistoricalDataManager {
	return &HistoricalDataManager{
		db:            db,
		binanceClient: NewBinanceClient(),
	}
}

// GetHistoricalCandles fetches candles with intelligent caching
// 1. Check database cache first
// 2. Fetch missing data from Binance
// 3. Cache new data in database
func (m *HistoricalDataManager) GetHistoricalCandles(symbol string, interval KlineInterval, startTime, endTime time.Time) ([]HistoricalCandle, error) {
	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	log.Printf("[HISTORICAL] Fetching %s %s candles from %s to %s",
		symbol, interval,
		startTime.Format("2006-01-02 15:04"),
		endTime.Format("2006-01-02 15:04"))

	// Step 1: Check database cache
	cachedCandles, err := m.getCachedCandles(symbol, interval, startMs, endMs)
	if err != nil {
		log.Printf("[HISTORICAL][WARN] Failed to query cache: %v", err)
		// Continue anyway, will fetch from API
	}

	// Step 2: Identify missing gaps
	gaps := m.findMissingGaps(cachedCandles, startMs, endMs, interval)

	if len(gaps) == 0 {
		log.Printf("[HISTORICAL] ✅ All %d candles found in cache (no API calls needed)", len(cachedCandles))
		return cachedCandles, nil
	}

	log.Printf("[HISTORICAL] Found %d cached candles, %d gaps to fill from API", len(cachedCandles), len(gaps))

	// Step 3: Fetch missing data from Binance
	var newCandles []HistoricalCandle
	for i, gap := range gaps {
		log.Printf("[HISTORICAL] Fetching gap %d/%d: %s to %s",
			i+1, len(gaps),
			time.UnixMilli(gap.Start).Format("2006-01-02 15:04"),
			time.UnixMilli(gap.End).Format("2006-01-02 15:04"))

		gapCandles, err := m.binanceClient.GetKlinesBatch(symbol, interval, gap.Start, gap.End)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch gap %d: %w", i+1, err)
		}

		newCandles = append(newCandles, gapCandles...)

		// Step 4: Cache new candles
		if err := m.cacheCandles(symbol, interval, gapCandles); err != nil {
			log.Printf("[HISTORICAL][WARN] Failed to cache %d candles: %v", len(gapCandles), err)
			// Continue anyway, we have the data
		}
	}

	// Step 5: Merge cached and new candles, sort by timestamp
	allCandles := append(cachedCandles, newCandles...)
	allCandles = m.sortAndDeduplicate(allCandles)

	log.Printf("[HISTORICAL] ✅ Total %d candles (%d from cache, %d from API)",
		len(allCandles), len(cachedCandles), len(newCandles))

	return allCandles, nil
}

// TimeGap represents a missing time range
type TimeGap struct {
	Start int64
	End   int64
}

// findMissingGaps identifies time ranges not covered by cached candles
func (m *HistoricalDataManager) findMissingGaps(cachedCandles []HistoricalCandle, startMs, endMs int64, interval KlineInterval) []TimeGap {
	if len(cachedCandles) == 0 {
		// No cached data, entire range is a gap
		return []TimeGap{{Start: startMs, End: endMs}}
	}

	intervalMs := m.getIntervalMilliseconds(interval)
	var gaps []TimeGap

	// Sort cached candles by timestamp (should already be sorted from DB)
	sortedCandles := m.sortAndDeduplicate(cachedCandles)

	// Check if there's a gap at the beginning
	firstCachedMs := sortedCandles[0].Timestamp.UnixMilli()
	if startMs < firstCachedMs-intervalMs {
		gaps = append(gaps, TimeGap{Start: startMs, End: firstCachedMs - intervalMs})
	}

	// Check for gaps between cached candles
	for i := 0; i < len(sortedCandles)-1; i++ {
		currentMs := sortedCandles[i].Timestamp.UnixMilli()
		nextMs := sortedCandles[i+1].Timestamp.UnixMilli()
		expectedNextMs := currentMs + intervalMs

		// If there's a gap larger than one interval
		if nextMs > expectedNextMs+intervalMs {
			gaps = append(gaps, TimeGap{Start: expectedNextMs, End: nextMs - intervalMs})
		}
	}

	// Check if there's a gap at the end
	lastCachedMs := sortedCandles[len(sortedCandles)-1].Timestamp.UnixMilli()
	if endMs > lastCachedMs+intervalMs {
		gaps = append(gaps, TimeGap{Start: lastCachedMs + intervalMs, End: endMs})
	}

	return gaps
}

// getCachedCandles retrieves candles from database
func (m *HistoricalDataManager) getCachedCandles(symbol string, interval KlineInterval, startMs, endMs int64) ([]HistoricalCandle, error) {
	query := `
		SELECT timestamp, open, high, low, close, volume
		FROM historical_candles
		WHERE symbol = ? AND interval = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp ASC
	`

	rows, err := m.db.Query(query, symbol, string(interval), startMs, endMs)
	if err != nil {
		return nil, fmt.Errorf("failed to query cached candles: %w", err)
	}
	defer rows.Close()

	var candles []HistoricalCandle
	for rows.Next() {
		var timestampMs int64
		var candle HistoricalCandle

		if err := rows.Scan(&timestampMs, &candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume); err != nil {
			return nil, fmt.Errorf("failed to scan candle: %w", err)
		}

		candle.Timestamp = time.UnixMilli(timestampMs)
		candles = append(candles, candle)
	}

	return candles, nil
}

// cacheCandles stores candles in database
func (m *HistoricalDataManager) cacheCandles(symbol string, interval KlineInterval, candles []HistoricalCandle) error {
	if len(candles) == 0 {
		return nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO historical_candles (symbol, interval, timestamp, open, high, low, close, volume)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, candle := range candles {
		result, err := stmt.Exec(
			symbol,
			string(interval),
			candle.Timestamp.UnixMilli(),
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
		if err != nil {
			log.Printf("[HISTORICAL][WARN] Failed to cache candle at %s: %v", candle.Timestamp, err)
			continue
		}

		rows, _ := result.RowsAffected()
		inserted += int(rows)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[HISTORICAL] Cached %d/%d new candles for %s %s", inserted, len(candles), symbol, interval)
	return nil
}

// sortAndDeduplicate sorts candles by timestamp and removes duplicates
func (m *HistoricalDataManager) sortAndDeduplicate(candles []HistoricalCandle) []HistoricalCandle {
	if len(candles) == 0 {
		return candles
	}

	// Create map to deduplicate by timestamp
	candleMap := make(map[int64]HistoricalCandle)
	for _, candle := range candles {
		timestampMs := candle.Timestamp.UnixMilli()
		// Keep first occurrence (could also keep latest)
		if _, exists := candleMap[timestampMs]; !exists {
			candleMap[timestampMs] = candle
		}
	}

	// Extract sorted unique candles
	timestamps := make([]int64, 0, len(candleMap))
	for ts := range candleMap {
		timestamps = append(timestamps, ts)
	}

	// Sort timestamps
	for i := 0; i < len(timestamps)-1; i++ {
		for j := i + 1; j < len(timestamps); j++ {
			if timestamps[i] > timestamps[j] {
				timestamps[i], timestamps[j] = timestamps[j], timestamps[i]
			}
		}
	}

	// Build sorted candle slice
	result := make([]HistoricalCandle, 0, len(timestamps))
	for _, ts := range timestamps {
		result = append(result, candleMap[ts])
	}

	return result
}

// getIntervalMilliseconds converts interval to milliseconds
func (m *HistoricalDataManager) getIntervalMilliseconds(interval KlineInterval) int64 {
	switch interval {
	case Interval1m:
		return 60 * 1000
	case Interval5m:
		return 5 * 60 * 1000
	case Interval15m:
		return 15 * 60 * 1000
	case Interval1h:
		return 60 * 60 * 1000
	case Interval4h:
		return 4 * 60 * 60 * 1000
	case Interval1d:
		return 24 * 60 * 60 * 1000
	default:
		return 60 * 1000
	}
}

// GetCacheStats returns statistics about cached data
func (m *HistoricalDataManager) GetCacheStats(symbol string, interval KlineInterval) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_candles,
			MIN(timestamp) as first_candle,
			MAX(timestamp) as last_candle,
			MIN(created_at) as first_cached,
			MAX(created_at) as last_cached
		FROM historical_candles
		WHERE symbol = ? AND interval = ?
	`

	var stats struct {
		TotalCandles int64
		FirstCandle  sql.NullInt64
		LastCandle   sql.NullInt64
		FirstCached  sql.NullString
		LastCached   sql.NullString
	}

	err := m.db.QueryRow(query, symbol, string(interval)).Scan(
		&stats.TotalCandles,
		&stats.FirstCandle,
		&stats.LastCandle,
		&stats.FirstCached,
		&stats.LastCached,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	result := map[string]interface{}{
		"symbol":        symbol,
		"interval":      string(interval),
		"total_candles": stats.TotalCandles,
	}

	if stats.FirstCandle.Valid {
		result["first_candle"] = time.UnixMilli(stats.FirstCandle.Int64).Format("2006-01-02 15:04:05")
		result["last_candle"] = time.UnixMilli(stats.LastCandle.Int64).Format("2006-01-02 15:04:05")
	}

	if stats.FirstCached.Valid {
		result["first_cached"] = stats.FirstCached.String
		result["last_cached"] = stats.LastCached.String
	}

	return result, nil
}

// CleanupOldCandles removes candles older than specified duration
func (m *HistoricalDataManager) CleanupOldCandles(olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	cutoffMs := cutoffTime.UnixMilli()

	result, err := m.db.Exec(`
		DELETE FROM historical_candles
		WHERE timestamp < ?
	`, cutoffMs)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old candles: %w", err)
	}

	deleted, _ := result.RowsAffected()
	log.Printf("[HISTORICAL] Cleaned up %d candles older than %s", deleted, cutoffTime.Format("2006-01-02"))

	return deleted, nil
}

// TestConnection tests Binance API connectivity
func (m *HistoricalDataManager) TestConnection() error {
	return m.binanceClient.TestConnection()
}

// GetLatestPrice fetches current price from Binance
func (m *HistoricalDataManager) GetLatestPrice(symbol string) (float64, error) {
	return m.binanceClient.GetLatestPrice(symbol)
}
