import styles from '../../views/MainView/MainView.module.css';
import { useTranslation } from '../../i18n';
import Flag from 'react-flagpack';
import type { ServerData } from '../ServerList/ServerList';
import Card from '../../ui/Card';
import Badge from '../../ui/Badge';

interface ServerCardProps {
  activeServer?: ServerData;
  connected: boolean;
  onOpenServers: () => void;
}

const emojiToCode: Record<string, string> = {
  "🇺🇸": "US", "🇬🇧": "GB", "🇩🇪": "DE", "🇫🇷": "FR", "🇳🇱": "NL",
  "🇷🇺": "RU", "🇺🇦": "UA", "🇰🇿": "KZ", "🇯🇵": "JP", "🇰🇷": "KR",
  "🇸🇬": "SG", "🇭🇰": "HK", "🇹🇼": "TW", "🇦🇺": "AU", "🇨🇦": "CA",
  "🇧🇷": "BR", "🇮🇳": "IN", "🇹🇷": "TR", "🇦🇪": "AE", "🇫🇮": "FI"
};

export default function ServerCard({ activeServer, connected, onOpenServers }: ServerCardProps) {
  const { t, getCountryName } = useTranslation();
  const code = activeServer ? emojiToCode[activeServer.flag] : null;

  return (
    <div className={styles.bottomSection}>
      <Card className={styles.serverCard} onClick={onOpenServers}>
        <div className={styles.flagBox}>
          {code ? <Flag code={code} size="l" /> : (
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20M2 12h20"/>
            </svg>
          )}
        </div>
        <div className={styles.serverInfo}>
          <div className={styles.serverNameRow}>
            <span className={styles.serverName}>
              {activeServer ? getCountryName(activeServer.flag, activeServer.name) : t('app.no_servers')}
            </span>
            <svg className={styles.arrowIcon} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </div>
          <Badge text={activeServer ? (activeServer.transport?.toUpperCase() || 'VLESS') : '---'} />
        </div>
        <div className={styles.serverLatency}>
          <span className={connected ? styles.latValue : styles.latDimmed}>
            {(activeServer?.latency_ms ?? 0) > 0 ? `${activeServer?.latency_ms} ms` : '-- ms'}
          </span>
        </div>
      </Card>
    </div>
  );
}
