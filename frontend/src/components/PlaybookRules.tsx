// ============================================
// FILE 3: src/components/PlaybookRules.tsx
// ============================================

interface PlaybookRule {
  rule_id: string;
  confidence: number;
  helpful_count: number;
  harmful_count: number;
  is_active: boolean;
}

interface PlaybookRulesProps {
  rules: PlaybookRule[];
}

export function PlaybookRules({ rules }: PlaybookRulesProps) {
  if (rules.length === 0) {
    return <div style={styles.emptyState}>No rules yet. ACE will learn from trades.</div>;
  }

  // Show only top 5 rules
  const topRules = rules.slice(0, 5);

  return (
    <div>
      {topRules.map((rule) => {
        const confidence = (rule.confidence || 0) * 100;
        const confidenceClass = confidence > 60 ? 'high' : confidence > 30 ? 'medium' : 'low';
        const cardClass = confidence > 60 ? 'success' : 'warning';

        return (
          <div
            key={rule.rule_id}
            style={{
              ...styles.ruleCard,
              borderLeftColor: cardClass === 'success' ? '#10b981' : '#f59e0b',
            }}
          >
            <div style={styles.ruleHeader}>
              <div style={styles.ruleName}>{rule.rule_id}</div>
              <span
                style={{
                  ...styles.badge,
                  ...(rule.is_active ? styles.badgeSuccess : styles.badgeWarning),
                }}
              >
                {rule.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div style={styles.ruleStats}>
              <span>✓ {rule.helpful_count || 0}</span>
              <span>✗ {rule.harmful_count || 0}</span>
              <span>
                <strong>{confidence.toFixed(0)}%</strong> confidence
              </span>
            </div>
            <div style={styles.confidenceBar}>
              <div
                style={{
                  ...styles.confidenceFill,
                  width: `${confidence}%`,
                  background:
                    confidenceClass === 'high' ? '#10b981' : confidenceClass === 'medium' ? '#f59e0b' : '#ef4444',
                }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}

export type { PlaybookRulesProps, PlaybookRule };

const styles: { [key: string]: React.CSSProperties } = {
  emptyState: {
    textAlign: 'center',
    padding: '40px',
    color: '#999',
  },
  ruleCard: {
    background: '#f9fafb',
    padding: '15px',
    borderRadius: '8px',
    marginBottom: '10px',
    borderLeft: '4px solid #667eea',
  },
  ruleHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '10px',
  },
  ruleName: {
    fontWeight: 600,
    fontSize: '14px',
  },
  ruleStats: {
    display: 'flex',
    gap: '20px',
    fontSize: '13px',
    color: '#666',
    marginBottom: '10px',
  },
  confidenceBar: {
    height: '8px',
    background: '#e0e0e0',
    borderRadius: '4px',
    overflow: 'hidden',
  },
  confidenceFill: {
    height: '100%',
    transition: 'width 0.3s',
  },
  badge: {
    padding: '4px 12px',
    borderRadius: '12px',
    fontSize: '12px',
    fontWeight: 600,
  },
  badgeSuccess: {
    background: '#d1fae5',
    color: '#065f46',
  },
  badgeWarning: {
    background: '#fef3c7',
    color: '#92400e',
  },
};
