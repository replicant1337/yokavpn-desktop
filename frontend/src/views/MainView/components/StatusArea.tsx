import styles from '../MainView.module.css';
import ConnectButton from '../../../components/ConnectButton/ConnectButton';
import type { StatsData } from '../../../components/StatsBar/StatsBar';

interface StatusAreaProps {
  connected: boolean;
  connecting: boolean;
  isDisconnecting: boolean;
  statusText?: string;
  stats: StatsData;
  toggleConnection: () => void;
}

export default function StatusArea({ 
  connected, 
  connecting, 
  isDisconnecting,
  statusText,
  stats, 
  toggleConnection 
}: StatusAreaProps) {
  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  };

  return (
    <div className={styles.mainContent}>
      <ConnectButton 
        connected={connected} 
        connecting={connecting} 
        isDisconnecting={isDisconnecting}
        statusText={statusText}
        onClick={toggleConnection} 
      />

      {connected && (
        <div className={styles.statsFloat}>
          <div className={styles.statItem}>
            <span className={styles.statUp}>↑</span>
            <span>{formatBytes(stats.upload)}</span>
          </div>
          <div className={styles.statItem}>
            <span className={styles.statDown}>↓</span>
            <span>{formatBytes(stats.download)}</span>
          </div>
        </div>
      )}
    </div>
  );
}


