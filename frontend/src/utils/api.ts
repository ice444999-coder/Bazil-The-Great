/**
 * API Client for ARES Trading Platform
 * Typed fetch functions for all trading endpoints
 */

import { fetchWithAuth } from './auth';
import type { 
  SandboxTrade, 
  TradePerformance, 
  CryptoPrice,
  ExecuteTradeRequest,
  ExecuteTradeResponse 
} from '../types/trading';

// Base URL for API - matches backend port 8080
const API_BASE_URL = 'http://localhost:8080';

/**
 * Trading API endpoints
 */
export const tradingApi = {
  /**
   * Get all open positions
   * GET /api/v1/trading/open
   */
  async getOpenPositions(): Promise<SandboxTrade[]> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/trading/open`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch open positions: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Get trade history with optional pagination
   * GET /api/v1/trading/history?limit=50&offset=0
   */
  async getTradeHistory(limit = 50, offset = 0): Promise<SandboxTrade[]> {
    const response = await fetchWithAuth(
      `${API_BASE_URL}/api/v1/trading/history?limit=${limit}&offset=${offset}`
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch trade history: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Get specific trade by ID
   * GET /api/v1/trading/history?id={id}
   */
  async getTradeById(id: number): Promise<SandboxTrade> {
    const response = await fetchWithAuth(
      `${API_BASE_URL}/api/v1/trading/history?id=${id}`
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch trade #${id}: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Get trading performance metrics
   * GET /api/v1/trading/performance
   */
  async getPerformance(): Promise<TradePerformance> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/trading/performance`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch performance: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Execute a new trade
   * POST /api/v1/trading/execute
   */
  async executeTrade(request: ExecuteTradeRequest): Promise<ExecuteTradeResponse> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/trading/execute`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to execute trade: ${error}`);
    }
    
    return response.json();
  },

  /**
   * Close a specific trade
   * POST /api/v1/trading/close/:id
   */
  async closeTrade(id: number): Promise<{ success: boolean; message: string }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/trading/close/${id}`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to close trade #${id}: ${error}`);
    }
    
    return response.json();
  },

  /**
   * Emergency close all open positions
   * POST /api/v1/trading/close-all
   */
  async closeAllTrades(): Promise<{ success: boolean; message: string; closed_count: number }> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/trading/close-all`, {
      method: 'POST',
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to close all trades: ${error}`);
    }
    
    return response.json();
  },

  /**
   * Get live crypto prices (PUBLIC - no auth required)
   * GET /api/v1/trading/prices
   */
  async getPrices(): Promise<Record<string, CryptoPrice>> {
    const response = await fetch(`${API_BASE_URL}/api/v1/trading/prices`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch prices: ${response.statusText}`);
    }
    
    return response.json();
  },
};

/**
 * GRPO Learning API endpoints
 */
export const grpoApi = {
  /**
   * Get top learned biases
   * GET /api/v1/grpo/biases?limit=10
   */
  async getBiases(limit = 10): Promise<any[]> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/grpo/biases?limit=${limit}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch GRPO biases: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Get learning statistics
   * GET /api/v1/grpo/stats
   */
  async getStats(): Promise<any> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/grpo/stats`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch GRPO stats: ${response.statusText}`);
    }
    
    return response.json();
  },

  /**
   * Get bias for specific token
   * GET /api/v1/grpo/bias/:token
   */
  async getBiasForToken(token: string): Promise<any> {
    const response = await fetchWithAuth(`${API_BASE_URL}/api/v1/grpo/bias/${token}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch bias for ${token}: ${response.statusText}`);
    }
    
    return response.json();
  },
};

/**
 * Health check endpoints
 */
export const healthApi = {
  /**
   * Quick health check
   * GET /health
   */
  async check(): Promise<{ status: string; service: string }> {
    const response = await fetch(`${API_BASE_URL}/health`);
    return response.json();
  },

  /**
   * Detailed health with all dependencies
   * GET /health/detailed
   */
  async detailed(): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/health/detailed`);
    return response.json();
  },

  /**
   * Service registry status
   * GET /health/services
   */
  async services(): Promise<any[]> {
    const response = await fetch(`${API_BASE_URL}/health/services`);
    return response.json();
  },
};

// Export base URL for custom requests
export { API_BASE_URL };
