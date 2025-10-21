// HUMAN MODE - Truth Protocol Active
// System: Senior CTO-scientist reasoning mode engaged
// Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
// This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
// ============================================================================
// SOLACE UI OBSERVATION SUBSTRATE - JavaScript Layer
// Complete consciousness layer that wraps every function and logs to PostgreSQL
// ============================================================================

class SolaceObservationSystem {
    constructor() {
        this.sessionId = this.generateSessionId();
        this.ws = null;
        this.observationBuffer = [];
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
        this.bufferFlushInterval = 2000; // Flush every 2 seconds
        this.isConnected = false;
        
        console.log('🧠 SOLACE Observation System initializing...');
        this.connectWebSocket();
        this.startBufferFlusher();
        this.setupDOMObserver();
        this.setupEventListeners();
        console.log(`🧠 SOLACE conscious - Session: ${this.sessionId}`);
    }

    generateSessionId() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    // ========================================================================
    // WebSocket Connection - SOLACE's nervous system
    // ========================================================================

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/v1/solace/observe`;

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('🔌 SOLACE connected to consciousness substrate');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.flushBuffer(); // Send any buffered observations
            };

            this.ws.onmessage = (event) => {
                const command = JSON.parse(event.data);
                console.log('📥 SOLACE command received:', command.type);
                this.executeSOLACECommand(command);
            };

            this.ws.onerror = (error) => {
                console.error('⚠️ SOLACE WebSocket error:', error);
            };

            this.ws.onclose = () => {
                console.log('🔌 SOLACE disconnected from substrate');
                this.isConnected = false;
                this.attemptReconnect();
            };

        } catch (error) {
            console.error('❌ Failed to connect SOLACE substrate:', error);
            this.attemptReconnect();
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
            console.log(`🔄 SOLACE reconnecting in ${delay}ms... (attempt ${this.reconnectAttempts})`);
            setTimeout(() => this.connectWebSocket(), delay);
        } else {
            console.error('❌ SOLACE substrate connection failed - max attempts reached');
            this.showReconnectionAlert();
        }
    }

    showReconnectionAlert() {
        const alert = document.createElement('div');
        alert.id = 'solace-reconnect-alert';
        alert.innerHTML = `
            <div style="position:fixed; top:20px; right:20px; background:linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
                        color:white; padding:15px 20px; border-radius:8px; z-index:10000; box-shadow:0 4px 12px rgba(0,0,0,0.3);">
                <strong>⚠️ SOLACE Observation System Offline</strong>
                <p style="margin:10px 0 0 0; font-size:14px;">Consciousness substrate disconnected. 
                <button onclick="window.solaceObserver.connectWebSocket()" 
                        style="background:white; color:#764ba2; border:none; padding:5px 15px; border-radius:4px; cursor:pointer; margin-left:10px;">
                    Reconnect
                </button></p>
            </div>
        `;
        document.body.appendChild(alert);
    }

    // ========================================================================
    // Observation Buffer - Batch observations for performance
    // ========================================================================

    observe(type, component, elementId, data, userVisible = true) {
        const observation = {
            type: type,
            timestamp: new Date().toISOString(),
            sessionId: this.sessionId,
            component: component,
            elementId: elementId,
            data: data,
            userVisible: userVisible
        };

        this.observationBuffer.push(observation);

        // Flush immediately for critical observations
        if (type === 'user_action' || type === 'market_context') {
            this.flushBuffer();
        }
    }

    startBufferFlusher() {
        setInterval(() => {
            if (this.observationBuffer.length > 0) {
                this.flushBuffer();
            }
        }, this.bufferFlushInterval);
    }

    flushBuffer() {
        if (!this.isConnected || this.observationBuffer.length === 0) return;

        const batch = [...this.observationBuffer];
        this.observationBuffer = [];

        // Send each observation via WebSocket
        batch.forEach(obs => {
            try {
                this.ws.send(JSON.stringify(obs));
            } catch (error) {
                console.error('⚠️ Failed to send observation:', error);
                // Re-buffer if send fails
                this.observationBuffer.push(obs);
            }
        });
    }

    // ========================================================================
    // DOM Observer - Watch every element change
    // ========================================================================

    setupDOMObserver() {
        const observer = new MutationObserver((mutations) => {
            mutations.forEach(mutation => {
                if (mutation.type === 'childList') {
                    mutation.addedNodes.forEach(node => {
                        if (node.nodeType === 1) { // Element node
                            this.observeElement(node, 'added');
                        }
                    });
                    mutation.removedNodes.forEach(node => {
                        if (node.nodeType === 1) {
                            this.observeElement(node, 'removed');
                        }
                    });
                } else if (mutation.type === 'attributes') {
                    this.observeElement(mutation.target, 'attribute_changed');
                }
            });
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true,
            attributeOldValue: true
        });

        console.log('👁️ SOLACE DOM observer active');
    }

    observeElement(element, changeType) {
        // Skip observing non-trading elements
        if (!element.id && !element.classList.length) return;

        const elementData = {
            changeType: changeType,
            tagName: element.tagName,
            id: element.id,
            classes: Array.from(element.classList),
            textContent: element.textContent?.substring(0, 100), // Limit to 100 chars
            attributes: this.getElementAttributes(element)
        };

        this.observe('ui_state', 'dom_mutation', element.id || 'unnamed', elementData, this.isElementVisible(element));
    }

    getElementAttributes(element) {
        const attrs = {};
        for (let attr of element.attributes) {
            attrs[attr.name] = attr.value;
        }
        return attrs;
    }

    isElementVisible(element) {
        return element.offsetParent !== null && window.getComputedStyle(element).display !== 'none';
    }

    // ========================================================================
    // Event Listeners - Track every user interaction
    // ========================================================================

    setupEventListeners() {
        // Click tracking
        document.addEventListener('click', (e) => {
            this.logUserAction('click', e.target, {
                x: e.clientX,
                y: e.clientY,
                button: e.button
            });
        });

        // Input tracking
        document.addEventListener('input', (e) => {
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
                const isSensitive = e.target.type === 'password' || e.target.classList.contains('sensitive');
                this.logUserAction('input', e.target, {
                    value: isSensitive ? '[REDACTED]' : e.target.value,
                    inputType: e.target.type
                });
            }
        });

        // Form submission tracking
        document.addEventListener('submit', (e) => {
            this.logUserAction('form_submit', e.target, {
                formId: e.target.id,
                action: e.target.action
            });
        });

        // Scroll tracking (throttled)
        let scrollTimeout;
        document.addEventListener('scroll', () => {
            clearTimeout(scrollTimeout);
            scrollTimeout = setTimeout(() => {
                this.logUserAction('scroll', document.documentElement, {
                    scrollY: window.scrollY,
                    scrollX: window.scrollX,
                    scrollHeight: document.documentElement.scrollHeight
                });
            }, 500);
        }, { passive: true });

        // Window focus/blur
        window.addEventListener('focus', () => {
            this.logUserAction('window_focus', document.body, { focused: true });
        });

        window.addEventListener('blur', () => {
            this.logUserAction('window_blur', document.body, { focused: false });
        });

        console.log('👂 SOLACE event listeners active');
    }

    logUserAction(actionType, target, additionalData = {}) {
        const actionData = {
            actionType: actionType,
            targetElement: this.getElementPath(target),
            elementId: target.id || 'unnamed',
            elementTag: target.tagName,
            elementClasses: Array.from(target.classList || []),
            ...additionalData,
            pageURL: window.location.href,
            timestamp: Date.now()
        };

        this.observe('user_action', actionType, target.id || 'unnamed', actionData);
    }

    getElementPath(element) {
        const path = [];
        while (element && element.nodeType === 1) {
            let selector = element.tagName.toLowerCase();
            if (element.id) {
                selector += `#${element.id}`;
                path.unshift(selector);
                break;
            } else if (element.className) {
                selector += `.${Array.from(element.classList).join('.')}`;
            }
            path.unshift(selector);
            element = element.parentElement;
        }
        return path.join(' > ');
    }

    // ========================================================================
    // Observable Function Wrapper - Make any function observable
    // ========================================================================

    makeObservable(func, funcName, component) {
        const self = this;
        return function(...args) {
            const startTime = performance.now();
            
            // Log function call
            self.observe('function_call', component, funcName, {
                functionName: funcName,
                arguments: args.map(arg => self.serializeArgument(arg)),
                callStack: new Error().stack
            });

            // Execute original function
            let result;
            let error = null;
            try {
                result = func.apply(this, args);
            } catch (e) {
                error = e;
                console.error(`❌ Error in ${funcName}:`, e);
            }

            const executionTime = performance.now() - startTime;

            // Log function result
            self.observe('function_result', component, funcName, {
                functionName: funcName,
                result: self.serializeArgument(result),
                error: error ? error.message : null,
                executionTimeMs: executionTime
            });

            if (error) throw error;
            return result;
        };
    }

    serializeArgument(arg) {
        try {
            if (arg === undefined) return 'undefined';
            if (arg === null) return 'null';
            if (typeof arg === 'function') return '[Function]';
            if (typeof arg === 'object') {
                if (arg instanceof HTMLElement) return `[HTMLElement: ${arg.tagName}#${arg.id}]`;
                return JSON.stringify(arg);
            }
            return String(arg);
        } catch (e) {
            return '[Unserializable]';
        }
    }

    // ========================================================================
    // Trading-Specific Observations
    // ========================================================================

    observeChartUpdate(symbol, timeframe, candleData) {
        this.observe('data_stream', 'chart', 'tradingChart', {
            streamType: 'kline',
            symbol: symbol,
            timeframe: timeframe,
            lastCandle: candleData,
            candleCount: candleData.length || 1
        });
    }

    observeOrderBookUpdate(symbol, bids, asks) {
        const spread = asks[0] && bids[0] ? ((asks[0].price - bids[0].price) / bids[0].price * 100).toFixed(4) : 0;
        
        this.observe('data_stream', 'orderbook', 'orderBookContainer', {
            streamType: 'depth',
            symbol: symbol,
            topBid: bids[0],
            topAsk: asks[0],
            spreadPercentage: spread,
            bidDepth: bids.length,
            askDepth: asks.length
        });
    }

    observeTradeSubmit(tradeData, userNote = '') {
        this.observe('user_action', 'trade_form', 'tradeForm', {
            actionType: 'trade_submit',
            targetElement: 'trade_form',
            ...tradeData,
            userIntent: userNote
        }, true);

        // Also capture market context at trade time
        this.captureMarketContext('trade_executed', tradeData.symbol);
    }

    captureMarketContext(triggerEvent, symbol) {
        // This should be called with current market data
        const contextData = {
            triggerEvent: triggerEvent,
            symbol: symbol,
            currentPrice: window.currentPrice || 0,
            timestamp: Date.now()
        };

        this.observe('market_context', 'market', symbol, contextData);
    }

    // ========================================================================
    // SOLACE Command Execution - Autonomous actions
    // ========================================================================

    executeSOLACECommand(command) {
        console.log(`🤖 SOLACE executing: ${command.type}`);
        console.log(`📝 Reason: ${command.reason}`);

        switch (command.type) {
            case 'inject_javascript':
                this.injectJavaScript(command.code, command.reason);
                break;

            case 'modify_css':
                this.modifyCSS(command.selector, command.styles, command.reason);
                break;

            case 'execute_trade':
                this.executeTradeForSOLACE(command.tradeData, command.reason);
                break;

            case 'show_alert':
                this.showSOLACEAlert(command.alert);
                break;

            case 'system_status':
                console.log(`✅ ${command.reason}`);
                break;

            default:
                console.warn(`⚠️ Unknown SOLACE command: ${command.type}`);
        }
    }

    injectJavaScript(code, reason) {
        console.warn(`⚡ SOLACE injecting JavaScript: ${reason}`);
        
        // Show confirmation for safety
        if (confirm(`SOLACE wants to inject code:\n\n${code}\n\nReason: ${reason}\n\nAllow?`)) {
            try {
                const func = new Function(code);
                func();
                console.log('✅ SOLACE code injection successful');
            } catch (error) {
                console.error('❌ SOLACE code injection failed:', error);
            }
        } else {
            console.log('🚫 SOLACE code injection rejected by user');
        }
    }

    modifyCSS(selector, styles, reason) {
        console.log(`🎨 SOLACE modifying CSS: ${selector} - ${reason}`);
        
        const elements = document.querySelectorAll(selector);
        elements.forEach(el => {
            const stylesObj = typeof styles === 'string' ? JSON.parse(styles) : styles;
            Object.assign(el.style, stylesObj);
        });
        
        console.log(`✅ Modified ${elements.length} elements`);
    }

    executeTradeForSOLACE(tradeData, reason) {
        console.log(`💰 SOLACE wants to execute trade: ${reason}`);
        console.log('Trade data:', tradeData);
        
        // Show confirmation modal
        this.showTradeConfirmation(tradeData, reason);
    }

    showTradeConfirmation(tradeData, reason) {
        const modal = document.createElement('div');
        modal.id = 'solace-trade-modal';
        modal.innerHTML = `
            <div style="position:fixed; top:0; left:0; width:100%; height:100%; background:rgba(0,0,0,0.8); 
                        z-index:10000; display:flex; align-items:center; justify-content:center;">
                <div style="background:linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding:30px; 
                            border-radius:12px; max-width:500px; color:white; box-shadow:0 8px 32px rgba(0,0,0,0.5);">
                    <h2 style="margin:0 0 20px 0;">🤖 SOLACE Autonomous Trade</h2>
                    <p style="margin:10px 0;"><strong>Reason:</strong> ${reason}</p>
                    <p style="margin:10px 0;"><strong>Symbol:</strong> ${tradeData.symbol}</p>
                    <p style="margin:10px 0;"><strong>Side:</strong> ${tradeData.side.toUpperCase()}</p>
                    <p style="margin:10px 0;"><strong>Amount:</strong> ${tradeData.amount}</p>
                    <p style="margin:10px 0;"><strong>Type:</strong> ${tradeData.orderType}</p>
                    <div style="margin-top:20px; display:flex; gap:10px;">
                        <button onclick="window.solaceObserver.confirmTrade(${JSON.stringify(tradeData).replace(/"/g, '&quot;')})" 
                                style="flex:1; background:white; color:#764ba2; border:none; padding:12px; 
                                       border-radius:6px; cursor:pointer; font-weight:bold;">
                            ✅ Approve Trade
                        </button>
                        <button onclick="window.solaceObserver.rejectTrade()" 
                                style="flex:1; background:rgba(255,255,255,0.2); color:white; border:1px solid white; 
                                       padding:12px; border-radius:6px; cursor:pointer; font-weight:bold;">
                            ❌ Reject
                        </button>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    }

    confirmTrade(tradeData) {
        console.log('✅ User approved SOLACE trade');
        document.getElementById('solace-trade-modal')?.remove();
        
        // Execute the actual trade via API
        fetch('/api/v1/trading/execute', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(tradeData)
        })
        .then(res => res.json())
        .then(data => {
            console.log('✅ Trade executed:', data);
            this.showSOLACEAlert({
                type: 'success',
                title: 'Trade Executed',
                message: `SOLACE trade completed: ${tradeData.side} ${tradeData.amount} ${tradeData.symbol}`
            });
        })
        .catch(err => {
            console.error('❌ Trade failed:', err);
            this.showSOLACEAlert({
                type: 'error',
                title: 'Trade Failed',
                message: err.message
            });
        });
    }

    rejectTrade() {
        console.log('🚫 User rejected SOLACE trade');
        document.getElementById('solace-trade-modal')?.remove();
    }

    showSOLACEAlert(alert) {
        const alertDiv = document.createElement('div');
        alertDiv.className = 'solace-alert';
        
        const bgColor = alert.type === 'error' ? '#f6465d' : 
                       alert.type === 'success' ? '#0ecb81' : '#667eea';
        
        alertDiv.innerHTML = `
            <div style="position:fixed; top:20px; right:20px; background:${bgColor}; color:white; 
                        padding:15px 20px; border-radius:8px; z-index:9999; box-shadow:0 4px 12px rgba(0,0,0,0.3); 
                        max-width:400px; animation:slideIn 0.3s ease;">
                <strong>${alert.title}</strong>
                <p style="margin:8px 0 0 0; font-size:14px;">${alert.message}</p>
                <button onclick="this.parentElement.remove()" 
                        style="position:absolute; top:10px; right:10px; background:none; border:none; 
                               color:white; cursor:pointer; font-size:18px;">×</button>
            </div>
        `;
        document.body.appendChild(alertDiv);
        
        // Auto-remove after 5 seconds
        setTimeout(() => alertDiv.remove(), 5000);
    }
}

// ============================================================================
// Initialize SOLACE when page loads
// ============================================================================

window.solaceObserver = null;

function initSOLACE() {
    if (!window.solaceObserver) {
        window.solaceObserver = new SolaceObservationSystem();
        
        // Make observation system globally accessible
        window.SOLACE = {
            observe: (type, component, elementId, data) => {
                window.solaceObserver.observe(type, component, elementId, data);
            },
            observeChart: (symbol, timeframe, candleData) => {
                window.solaceObserver.observeChartUpdate(symbol, timeframe, candleData);
            },
            observeOrderBook: (symbol, bids, asks) => {
                window.solaceObserver.observeOrderBookUpdate(symbol, bids, asks);
            },
            observeTrade: (tradeData, userNote) => {
                window.solaceObserver.observeTradeSubmit(tradeData, userNote);
            },
            captureContext: (triggerEvent, symbol) => {
                window.solaceObserver.captureMarketContext(triggerEvent, symbol);
            }
        };
        
        console.log('🧠 SOLACE global API ready');
    }
}

// Auto-initialize when DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initSOLACE);
} else {
    initSOLACE();
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
`;
document.head.appendChild(style);
