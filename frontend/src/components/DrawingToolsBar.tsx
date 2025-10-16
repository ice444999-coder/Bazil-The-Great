import { useState } from 'react';

interface DrawingToolsBarProps {
  onToolSelect: (toolId: string) => void;
  onClearDrawings?: () => void;
  onUndoLast?: () => void;
}

interface DrawingTool {
  id: string;
  name: string;
  icon: string;
  description: string;
  hotkey?: string;
}

const DRAWING_TOOLS: DrawingTool[] = [
  {
    id: 'cursor',
    name: 'Cursor',
    icon: '↖️',
    description: 'Select and move objects',
    hotkey: 'ESC',
  },
  {
    id: 'trendline',
    name: 'Trend Line',
    icon: '📈',
    description: 'Draw trend lines',
    hotkey: 'T',
  },
  {
    id: 'horizontal',
    name: 'Horizontal Line',
    icon: '➖',
    description: 'Draw horizontal support/resistance',
    hotkey: 'H',
  },
  {
    id: 'vertical',
    name: 'Vertical Line',
    icon: '|',
    description: 'Draw vertical time markers',
    hotkey: 'V',
  },
  {
    id: 'rectangle',
    name: 'Rectangle',
    icon: '▭',
    description: 'Draw price ranges',
    hotkey: 'R',
  },
  {
    id: 'fibonacci',
    name: 'Fibonacci',
    icon: '𝝋',
    description: 'Fibonacci retracement levels',
    hotkey: 'F',
  },
  {
    id: 'text',
    name: 'Text',
    icon: 'T',
    description: 'Add text annotations',
    hotkey: 'X',
  },
  {
    id: 'arrow',
    name: 'Arrow',
    icon: '→',
    description: 'Draw directional arrows',
    hotkey: 'A',
  },
];

export default function DrawingToolsBar({ 
  onToolSelect, 
  onClearDrawings,
  onUndoLast 
}: DrawingToolsBarProps) {
  const [selectedTool, setSelectedTool] = useState('cursor');
  const [showTooltip, setShowTooltip] = useState<string | null>(null);

  const handleToolClick = (tool: DrawingTool) => {
    setSelectedTool(tool.id);
    onToolSelect(tool.id);
  };

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        background: '#1E2329',
        padding: '8px 12px',
        borderRadius: 6,
        border: '1px solid #2B3139',
      }}
    >
      {/* Label */}
      <div
        style={{
          fontSize: 12,
          color: '#848E9C',
          fontWeight: 600,
          marginRight: 4,
        }}
      >
        Draw:
      </div>

      {/* Drawing Tools */}
      <div style={{ display: 'flex', gap: 4, flex: 1 }}>
        {DRAWING_TOOLS.map((tool) => {
          const isSelected = selectedTool === tool.id;

          return (
            <div
              key={tool.id}
              style={{ position: 'relative' }}
              onMouseEnter={() => setShowTooltip(tool.id)}
              onMouseLeave={() => setShowTooltip(null)}
            >
              <button
                onClick={() => handleToolClick(tool)}
                style={{
                  width: 36,
                  height: 36,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: isSelected ? '#F0B90B' : 'transparent',
                  color: isSelected ? '#0B0E11' : '#FFFFFF',
                  border: isSelected ? 'none' : '1px solid #2B3139',
                  borderRadius: 4,
                  fontSize: 16,
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  if (!isSelected) {
                    e.currentTarget.style.background = '#2B3139';
                    e.currentTarget.style.borderColor = '#F0B90B';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isSelected) {
                    e.currentTarget.style.background = 'transparent';
                    e.currentTarget.style.borderColor = '#2B3139';
                  }
                }}
              >
                {tool.icon}
              </button>

              {/* Tooltip */}
              {showTooltip === tool.id && (
                <div
                  style={{
                    position: 'absolute',
                    top: '100%',
                    left: '50%',
                    transform: 'translateX(-50%)',
                    marginTop: 8,
                    background: '#0B0E11',
                    border: '1px solid #F0B90B',
                    borderRadius: 6,
                    padding: '8px 12px',
                    minWidth: 160,
                    zIndex: 1000,
                    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.5)',
                  }}
                >
                  <div
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: '#FFFFFF',
                      marginBottom: 4,
                    }}
                  >
                    {tool.name}
                  </div>
                  <div
                    style={{
                      fontSize: 11,
                      color: '#848E9C',
                      marginBottom: tool.hotkey ? 4 : 0,
                    }}
                  >
                    {tool.description}
                  </div>
                  {tool.hotkey && (
                    <div
                      style={{
                        fontSize: 10,
                        color: '#F0B90B',
                        fontWeight: 600,
                      }}
                    >
                      Hotkey: {tool.hotkey}
                    </div>
                  )}

                  {/* Tooltip Arrow */}
                  <div
                    style={{
                      position: 'absolute',
                      bottom: '100%',
                      left: '50%',
                      transform: 'translateX(-50%)',
                      width: 0,
                      height: 0,
                      borderLeft: '6px solid transparent',
                      borderRight: '6px solid transparent',
                      borderBottom: '6px solid #F0B90B',
                    }}
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Divider */}
      <div
        style={{
          width: 1,
          height: 24,
          background: '#2B3139',
          margin: '0 4px',
        }}
      />

      {/* Action Buttons */}
      <div style={{ display: 'flex', gap: 4 }}>
        {/* Undo Button */}
        {onUndoLast && (
          <button
            onClick={onUndoLast}
            style={{
              padding: '8px 12px',
              background: 'transparent',
              border: '1px solid #2B3139',
              borderRadius: 4,
              color: '#FFFFFF',
              fontSize: 12,
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'all 0.2s',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = '#F0B90B';
              e.currentTarget.style.color = '#F0B90B';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = '#2B3139';
              e.currentTarget.style.color = '#FFFFFF';
            }}
          >
            ↶ Undo
          </button>
        )}

        {/* Clear All Button */}
        {onClearDrawings && (
          <button
            onClick={onClearDrawings}
            style={{
              padding: '8px 12px',
              background: 'transparent',
              border: '1px solid #2B3139',
              borderRadius: 4,
              color: '#F6465D',
              fontSize: 12,
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'all 0.2s',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = '#F6465D';
              e.currentTarget.style.background = '#F6465D10';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = '#2B3139';
              e.currentTarget.style.background = 'transparent';
            }}
          >
            🗑️ Clear
          </button>
        )}
      </div>
    </div>
  );
}

export { DRAWING_TOOLS };
export type { DrawingTool };
