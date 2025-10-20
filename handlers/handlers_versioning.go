package handlers

import (
	"ares_api/internal/trading"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// VersioningHandlers groups all strategy versioning API endpoints
type VersioningHandlers struct {
	versionManager *trading.StrategyVersionManager
}

// NewVersioningHandlers creates handlers for versioning endpoints
func NewVersioningHandlers(db *sql.DB) *VersioningHandlers {
	return &VersioningHandlers{
		versionManager: trading.NewStrategyVersionManager(db),
	}
}

// HandleListVersions - GET /api/v1/strategies/{name}/versions
func (vh *VersioningHandlers) HandleListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	versions, err := vh.versionManager.ListVersions(strategyName)
	if err != nil {
		log.Printf("[ERROR] Failed to list versions for %s: %v", strategyName, err)
		http.Error(w, "Failed to list versions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"strategy": strategyName,
		"versions": versions,
		"count":    len(versions),
	})
}

// HandleGetVersion - GET /api/v1/strategies/{name}/versions/{version}
func (vh *VersioningHandlers) HandleGetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]
	versionNum, err := strconv.Atoi(vars["version"])
	if err != nil {
		http.Error(w, "Invalid version number", http.StatusBadRequest)
		return
	}

	version, err := vh.versionManager.GetVersion(strategyName, versionNum)
	if err != nil {
		log.Printf("[ERROR] Failed to get version %d for %s: %v", versionNum, strategyName, err)
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

// HandleGetActiveVersion - GET /api/v1/strategies/{name}/versions/active
func (vh *VersioningHandlers) HandleGetActiveVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	version, err := vh.versionManager.GetActiveVersion(strategyName)
	if err != nil {
		log.Printf("[ERROR] Failed to get active version for %s: %v", strategyName, err)
		http.Error(w, "Active version not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

// CreateVersionRequest - Request body for creating a new version
type CreateVersionRequest struct {
	ConfigJSON       string `json:"config_json"`
	Notes            string `json:"notes"`
	CreatedBy        string `json:"created_by"`
	BacktestResultID *int   `json:"backtest_result_id,omitempty"`
}

// HandleCreateVersion - POST /api/v1/strategies/{name}/versions
func (vh *VersioningHandlers) HandleCreateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	var req CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate JSON config
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigJSON), &configMap); err != nil {
		http.Error(w, "Invalid JSON in config_json field", http.StatusBadRequest)
		return
	}

	version, err := vh.versionManager.CreateVersion(strategyName, req.ConfigJSON, req.Notes, req.CreatedBy, req.BacktestResultID)
	if err != nil {
		log.Printf("[ERROR] Failed to create version for %s: %v", strategyName, err)
		http.Error(w, "Failed to create version", http.StatusInternalServerError)
		return
	}

	log.Printf("[VERSION] Created version %d for %s by %s", version.Version, strategyName, req.CreatedBy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// RollbackRequest - Request body for rolling back to a version
type RollbackRequest struct {
	TargetVersion int    `json:"target_version"`
	Reason        string `json:"reason"`
	TriggeredBy   string `json:"triggered_by"` // 'manual', 'auto', 'emergency'
	CreatedBy     string `json:"created_by"`
}

// HandleRollback - POST /api/v1/strategies/{name}/rollback
func (vh *VersioningHandlers) HandleRollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	var req RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TriggeredBy == "" {
		req.TriggeredBy = "manual"
	}

	err := vh.versionManager.RollbackToVersion(strategyName, req.TargetVersion, req.Reason, req.TriggeredBy, req.CreatedBy)
	if err != nil {
		log.Printf("[ERROR] Failed to rollback %s to version %d: %v", strategyName, req.TargetVersion, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[ROLLBACK] %s rolled back to version %d by %s (%s): %s", strategyName, req.TargetVersion, req.CreatedBy, req.TriggeredBy, req.Reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"strategy":       strategyName,
		"target_version": req.TargetVersion,
		"reason":         req.Reason,
		"triggered_by":   req.TriggeredBy,
		"timestamp":      time.Now(),
	})
}

// HandleGetRollbackHistory - GET /api/v1/strategies/{name}/rollback-history
func (vh *VersioningHandlers) HandleGetRollbackHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	// Optional limit parameter
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history, err := vh.versionManager.GetRollbackHistory(strategyName, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to get rollback history for %s: %v", strategyName, err)
		http.Error(w, "Failed to get rollback history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"strategy": strategyName,
		"history":  history,
		"count":    len(history),
	})
}

// BacktestRequest - Request body for running a backtest
type BacktestRequest struct {
	Version         *int    `json:"version,omitempty"` // Specific version to test, or active if nil
	Symbol          string  `json:"symbol"`
	Timeframe       string  `json:"timeframe"`
	NumCandles      int     `json:"num_candles"`
	StartingBalance float64 `json:"starting_balance,omitempty"`
	PositionSize    float64 `json:"position_size,omitempty"`
}

// HandleRunBacktest - POST /api/v1/strategies/{name}/backtest
func (vh *VersioningHandlers) HandleRunBacktest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	var req BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Defaults
	if req.Symbol == "" {
		req.Symbol = "BTC/USDT"
	}
	if req.NumCandles == 0 {
		req.NumCandles = 500
	}
	if req.StartingBalance == 0 {
		req.StartingBalance = 10000
	}
	if req.PositionSize == 0 {
		req.PositionSize = 2.0
	}

	// TODO: Load real historical data instead of synthetic
	// For now, generate synthetic data
	_ = trading.GenerateSyntheticData(req.Symbol, req.NumCandles, 50000.0)

	// Get version to test (active or specific)
	var version *trading.StrategyVersion
	var err error
	if req.Version != nil {
		version, err = vh.versionManager.GetVersion(strategyName, *req.Version)
	} else {
		version, err = vh.versionManager.GetActiveVersion(strategyName)
	}
	if err != nil {
		log.Printf("[ERROR] Failed to get version for backtest: %v", err)
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	// TODO: Create strategy instance with config from version.ConfigJSON
	// For now, return placeholder response
	log.Printf("[BACKTEST] Running backtest for %s version %d on %s (%d candles)",
		strategyName, version.Version, req.Symbol, req.NumCandles)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "Backtest functionality requires strategy factory implementation",
		"strategy":         strategyName,
		"version":          version.Version,
		"symbol":           req.Symbol,
		"num_candles":      req.NumCandles,
		"starting_balance": req.StartingBalance,
		"note":             "Use existing backtester_test.go for now",
	})
}

// HandleListBacktestResults - GET /api/v1/strategies/{name}/backtest-results
func (vh *VersioningHandlers) HandleListBacktestResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	// Optional limit parameter
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := vh.versionManager.ListBacktestResults(strategyName, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to list backtest results for %s: %v", strategyName, err)
		http.Error(w, "Failed to list backtest results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"strategy": strategyName,
		"results":  results,
		"count":    len(results),
	})
}

// HandleGetBacktestResult - GET /api/v1/backtest-results/{id}
func (vh *VersioningHandlers) HandleGetBacktestResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid backtest result ID", http.StatusBadRequest)
		return
	}

	result, err := vh.versionManager.GetBacktestResult(id)
	if err != nil {
		log.Printf("[ERROR] Failed to get backtest result %d: %v", id, err)
		http.Error(w, "Backtest result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RegisterVersioningRoutes registers all versioning endpoints
func RegisterVersioningRoutes(router *mux.Router, db *sql.DB) {
	vh := NewVersioningHandlers(db)

	// Version management
	router.HandleFunc("/api/v1/strategies/{name}/versions", vh.HandleListVersions).Methods("GET")
	router.HandleFunc("/api/v1/strategies/{name}/versions/active", vh.HandleGetActiveVersion).Methods("GET")
	router.HandleFunc("/api/v1/strategies/{name}/versions/{version:[0-9]+}", vh.HandleGetVersion).Methods("GET")
	router.HandleFunc("/api/v1/strategies/{name}/versions", vh.HandleCreateVersion).Methods("POST")

	// Rollback
	router.HandleFunc("/api/v1/strategies/{name}/rollback", vh.HandleRollback).Methods("POST")
	router.HandleFunc("/api/v1/strategies/{name}/rollback-history", vh.HandleGetRollbackHistory).Methods("GET")

	// Backtesting
	router.HandleFunc("/api/v1/strategies/{name}/backtest", vh.HandleRunBacktest).Methods("POST")
	router.HandleFunc("/api/v1/strategies/{name}/backtest-results", vh.HandleListBacktestResults).Methods("GET")
	router.HandleFunc("/api/v1/backtest-results/{id:[0-9]+}", vh.HandleGetBacktestResult).Methods("GET")

	log.Println("[ROUTES] Registered 9 strategy versioning endpoints")
}
