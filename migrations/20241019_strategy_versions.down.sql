-- Rollback strategy versioning tables
DROP INDEX IF EXISTS idx_rollback_history_created;
DROP INDEX IF EXISTS idx_rollback_history_strategy;
DROP TABLE IF EXISTS strategy_rollback_history;

DROP INDEX IF EXISTS idx_backtest_results_performance;
DROP INDEX IF EXISTS idx_backtest_results_created;
DROP INDEX IF EXISTS idx_backtest_results_version;
DROP INDEX IF EXISTS idx_backtest_results_strategy;
DROP TABLE IF EXISTS backtest_results;

DROP INDEX IF EXISTS idx_strategy_versions_created;
DROP INDEX IF EXISTS idx_strategy_versions_active;
DROP INDEX IF EXISTS idx_strategy_versions_name;
DROP TABLE IF EXISTS strategy_versions;
