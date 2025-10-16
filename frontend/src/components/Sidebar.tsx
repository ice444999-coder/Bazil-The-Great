import { useState, useEffect } from 'react';

interface SidebarProps {
  activePage: string;
  onNavigate: (page: string) => void;
  onLogout: () => void;
}

export default function Sidebar({ activePage, onNavigate, onLogout }: SidebarProps) {
  const [username, setUsername] = useState('User');

  useEffect(() => {
    // Get username from localStorage
    const userData = localStorage.getItem('ares_user');
    if (userData) {
      try {
        const user = JSON.parse(userData);
        setUsername(user.username || 'User');
      } catch (e) {
        console.error('Failed to parse user data:', e);
      }
    }
  }, []);

  const navItems = [
    { id: 'dashboard', icon: '🏠', label: 'Dashboard' },
    { id: 'trading', icon: '📊', label: 'Trading' },
    { id: 'chat', icon: '💬', label: 'Chat with SOLACE' },
    { id: 'editor', icon: '📝', label: 'Code Editor' },
    { id: 'memory', icon: '🧠', label: 'Memory System' },
    { id: 'vision', icon: '👁️', label: 'SOLACE Vision' },
    { id: 'health', icon: '❤️', label: 'System Health' },
  ];

  return (
    <div style={styles.sidebar}>
      {/* Logo Section */}
      <div style={styles.logo}>
        <h1 style={styles.logoTitle}>ARES</h1>
        <p style={styles.logoSubtitle}>Autonomous AI Trading System</p>
      </div>

      {/* User Info Section */}
      <div style={styles.userInfo}>
        <div style={styles.username}>{username}</div>
        <div style={styles.role}>Administrator</div>
      </div>

      {/* Navigation Menu */}
      <nav style={styles.navMenu}>
        {navItems.map((item) => (
          <div
            key={item.id}
            style={{
              ...styles.navItem,
              ...(activePage === item.id ? styles.navItemActive : {}),
            }}
            onClick={() => onNavigate(item.id)}
            onMouseEnter={(e) => {
              if (activePage !== item.id) {
                e.currentTarget.style.background = 'rgba(255,255,255,0.1)';
              }
            }}
            onMouseLeave={(e) => {
              if (activePage !== item.id) {
                e.currentTarget.style.background = 'transparent';
              }
            }}
          >
            <span style={styles.navIcon}>{item.icon}</span>
            <span>{item.label}</span>
          </div>
        ))}
      </nav>

      {/* Logout Button */}
      <div
        style={styles.logoutBtn}
        onClick={onLogout}
        onMouseEnter={(e) => {
          e.currentTarget.style.background = 'rgba(255,255,255,0.3)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.background = 'rgba(255,255,255,0.2)';
        }}
      >
        🚪 Logout
      </div>
    </div>
  );
}

const styles: { [key: string]: React.CSSProperties } = {
  sidebar: {
    width: '250px',
    background: 'linear-gradient(180deg, #667eea 0%, #764ba2 100%)',
    color: 'white',
    display: 'flex',
    flexDirection: 'column',
    padding: '20px 0',
    height: '100vh',
    position: 'fixed',
    left: 0,
    top: 0,
  },
  logo: {
    padding: '0 20px 20px',
    borderBottom: '1px solid rgba(255,255,255,0.2)',
    marginBottom: '20px',
  },
  logoTitle: {
    fontSize: '28px',
    marginBottom: '5px',
    fontWeight: 'bold',
  },
  logoSubtitle: {
    fontSize: '12px',
    opacity: 0.8,
    margin: 0,
  },
  userInfo: {
    padding: '0 20px 20px',
    borderBottom: '1px solid rgba(255,255,255,0.2)',
    marginBottom: '20px',
  },
  username: {
    fontWeight: 600,
    fontSize: '14px',
    marginBottom: '5px',
  },
  role: {
    fontSize: '12px',
    opacity: 0.7,
  },
  navMenu: {
    flex: 1,
    overflowY: 'auto',
  },
  navItem: {
    padding: '12px 20px',
    cursor: 'pointer',
    transition: 'background 0.3s',
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    color: 'white',
    textDecoration: 'none',
  },
  navItemActive: {
    background: 'rgba(255,255,255,0.2)',
    borderLeft: '3px solid white',
  },
  navIcon: {
    fontSize: '18px',
  },
  logoutBtn: {
    padding: '12px 20px',
    margin: '20px',
    background: 'rgba(255,255,255,0.2)',
    border: '1px solid rgba(255,255,255,0.3)',
    color: 'white',
    borderRadius: '5px',
    cursor: 'pointer',
    textAlign: 'center',
    transition: 'background 0.3s',
  },
};

export type { SidebarProps };
