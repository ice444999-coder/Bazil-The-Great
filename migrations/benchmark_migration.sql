-- Minimal migration for benchmark testing
-- Creates essential tables for AI query performance testing

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create schemas
CREATE SCHEMA IF NOT EXISTS trading_core;
CREATE SCHEMA IF NOT EXISTS memory_system;
CREATE SCHEMA IF NOT EXISTS solace_core;
CREATE SCHEMA IF NOT EXISTS tool_system;

-- Trading core tables
CREATE TABLE IF NOT EXISTS trading_core.trades (
    trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(4) NOT NULL, -- 'BUY' or 'SELL'
    quantity DECIMAL(10,2) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trading_core.strategies (
    strategy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_name VARCHAR(100) UNIQUE NOT NULL,
    strategy_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    performance_metrics JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    embedding vector(1536)
);

-- Memory system tables
CREATE TABLE IF NOT EXISTS memory_system.memory_embeddings (
    embedding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type VARCHAR(50) NOT NULL,
    content_hash VARCHAR(64) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS memory_system.conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_message TEXT NOT NULL,
    ai_response TEXT NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- SOLACE core tables
CREATE TABLE IF NOT EXISTS solace_core.solace_decisions (
    decision_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_type VARCHAR(50) NOT NULL,
    context JSONB NOT NULL,
    decision TEXT NOT NULL,
    confidence_score DECIMAL(3,2),
    outcome JSONB,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tool system tables
CREATE TABLE IF NOT EXISTS tool_system.tool_registry (
    tool_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50),
    embedding vector(1536),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert some sample data for benchmarking
INSERT INTO trading_core.trades (symbol, side, quantity, price) VALUES
('AAPL', 'BUY', 100, 150.00),
('GOOGL', 'SELL', 50, 2800.00),
('MSFT', 'BUY', 75, 300.00);

INSERT INTO memory_system.memory_embeddings (content_type, content_hash, content, embedding) VALUES
('conversation', 'hash1', 'Sample conversation content', NULL),
('decision', 'hash2', 'Sample decision content', NULL);

INSERT INTO solace_core.solace_decisions (decision_type, context, decision, confidence_score, embedding) VALUES
('trading', '{"symbol": "AAPL"}', 'Buy AAPL based on analysis', 0.85, NULL);

INSERT INTO tool_system.tool_registry (tool_name, description, category, embedding) VALUES
('data_analyzer', 'Analyzes market data', 'analysis', NULL),
('trade_executor', 'Executes trades', 'execution', NULL);