import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '100vh',
          backgroundColor: '#0B0E11',
          color: '#EAECEF',
          fontFamily: 'monospace',
          padding: '20px'
        }}>
          <div style={{
            backgroundColor: '#1E2329',
            border: '2px solid #F6465D',
            borderRadius: '8px',
            padding: '32px',
            maxWidth: '800px'
          }}>
            <h1 style={{ color: '#F6465D', marginTop: 0 }}>
              ⚠️ Something went wrong
            </h1>
            <p style={{ color: '#848E9C', marginBottom: '16px' }}>
              The application encountered an error and couldn't recover.
            </p>
            <details style={{ marginBottom: '24px' }}>
              <summary style={{ cursor: 'pointer', color: '#F0B90B' }}>
                Error Details
              </summary>
              <pre style={{
                backgroundColor: '#0B0E11',
                padding: '16px',
                borderRadius: '4px',
                overflow: 'auto',
                marginTop: '12px',
                fontSize: '12px'
              }}>
                {this.state.error?.toString()}
                {'\n\n'}
                {this.state.error?.stack}
              </pre>
            </details>
            <button
              onClick={() => window.location.reload()}
              style={{
                backgroundColor: '#0ECB81',
                color: 'white',
                border: 'none',
                padding: '12px 24px',
                borderRadius: '6px',
                cursor: 'pointer',
                fontWeight: 'bold',
                fontSize: '14px'
              }}
            >
              Reload Page
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
