import styles from './ServerList.module.css';
import Flag from 'react-flagpack';
import Loader from '../Loader/Loader';
import { useTranslation } from '../../i18n';

export interface ServerData {
  name: string;
  flag: string;
  type?: string;
  transport?: string;
  latency_ms?: number;
  link?: string;
}

interface ServerListProps {
  servers?: ServerData[];
  selected?: number;
  pinging?: Set<number>;
  onSelect?: (index: number) => void;
  onPing?: (index: number) => void;
}

const emojiToCode: Record<string, string> = {
  "🇺🇸": "US", "🇬🇧": "GB", "🇩🇪": "DE", "🇫🇷": "FR", "🇳🇱": "NL",
  "🇷🇺": "RU", "🇺🇦": "UA", "🇰🇿": "KZ", "🇯🇵": "JP", "🇰🇷": "KR",
  "🇸🇬": "SG", "🇭🇰": "HK", "🇹🇼": "TW", "🇦🇺": "AU", "🇨🇦": "CA",
  "🇧🇷": "BR", "🇮🇳": "IN", "🇹🇷": "TR", "🇦🇪": "AE", "🇫🇮": "FI"
};

export default function ServerList({ 
  servers = [], 
  selected = 0, 
  pinging = new Set(),
  onSelect = () => {}, 
  onPing = () => {} 
}: ServerListProps) {
  const { t, getCountryName } = useTranslation();
  
  const getBadges = (server: ServerData) => {
    const badges: { text: string, type: 'protocol' | 'service' | 'location' }[] = [];
    
    // Protocol badges
    if (server.transport) {
      badges.push({ text: server.transport.toLowerCase(), type: 'protocol' });
    }
    if (server.type && server.type !== 'vless') {
      badges.push({ text: server.type.toLowerCase(), type: 'protocol' });
    }

    // Service badges based on location
    if (server.flag === "🇷🇺") {
      badges.push({ text: t('badges.russia'), type: 'location' });
      badges.push({ text: t('badges.banks'), type: 'service' });
    } else {
      badges.push({ text: t('badges.youtube'), type: 'service' });
      badges.push({ text: t('badges.chatgpt'), type: 'service' });
      badges.push({ text: t('badges.telegram'), type: 'service' });
    }

    return badges;
  };

  return (
    <div className={styles.servers}>
      {servers.map((s, i) => {
        const code = emojiToCode[s.flag];
        const badges = getBadges(s);
        
        return (
          <div 
            key={i} 
            className={`${styles.server} ${i === selected ? styles.sel : ''}`.trim()} 
            onClick={() => onSelect(i)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onSelect(i);
              }
            }}
          >
            {code ? (
              <div className={styles.flagIcon}>
                <Flag code={code} size="l" hasDropShadow={true} hasBorder={true} />
              </div>
            ) : (
              <span className={styles.flag}>{s.flag}</span>
            )}
            <div className={styles.info}>
              <div className={styles.nameRow}>
                <span className={styles.name}>{getCountryName(s.flag, s.name)}</span>
              </div>
              <div className={styles.badgeRow}>
                {badges.map((b, bi) => (
                  <span key={bi} className={`${styles.badge} ${styles[b.type]}`}>
                    {b.text}
                  </span>
                ))}
              </div>
            </div>
            <div className={styles.action}>
              {pinging.has(i) ? (
                <Loader size="small" color="green" />
              ) : s.latency_ms === -1 ? (
                <span className={styles.latencyError}>Error</span>
              ) : (s.latency_ms ?? 0) > 0 ? (
                <span className={styles.latency}>{s.latency_ms}ms</span>
              ) : (
                <button 
                  className={styles.pingBtn} 
                  onClick={(e) => {
                    e.stopPropagation();
                    onPing(i);
                  }}
                  tabIndex={-1}
                >
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
                  </svg>
                </button>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}