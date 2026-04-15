import styles from './StatsBar.module.css';

export interface StatsData {
  upload: number;
  download: number;
}

interface StatsBarProps {
  connected?: boolean;
  useTun?: boolean;
  stats?: StatsData;
  serverName?: string;
}

export default function StatsBar({ 
  connected = false, 
  useTun = false, 
  stats = { upload: 0, download: 0 },
  serverName = ""
}: StatsBarProps) {
  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  };

  return (
    <div className={styles.container}>
      {connected && serverName && (
        <div className={styles.serverInfo}>
          <span className={styles.dot}></span>
          <span className={styles.serverName}>{serverName}</span>
        </div>
      )}
      <div className={styles.stats}>
        {connected && (
          <span className={styles.statItem}>
            <span className={styles.statLabel}>{useTun ? 'VPN' : 'Proxy'}</span>
          </span>
        )}
        <span className={styles.statItem}>
          <span className={styles.statUp}>↑</span>
          <span className={styles.val}>{formatBytes(stats.upload)}</span>
        </span>
        <span className={styles.statItem}>
          <span className={styles.statDown}>↓</span>
          <span className={styles.val}>{formatBytes(stats.download)}</span>
        </span>
      </div>
    </div>
  );
}