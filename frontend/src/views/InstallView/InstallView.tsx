import styles from './InstallView.module.css';

interface InstallViewProps {
  statusText: string;
  progress: { percentage: number; label: string };
}

export default function InstallView({ statusText, progress }: InstallViewProps) {
  const message = statusText === 'app.status_installing_core' 
    ? 'Downloading Xray Core...' 
    : statusText === 'app.status_installing_tun'
      ? 'Downloading Tun2Proxy...'
      : statusText === 'app.status_checking_assets' 
        ? 'Updating Geo Assets...' 
        : 'Setting up your application...';

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <div className={styles.loader}></div>
        <h1 className={styles.title}>Welcome to Xray Client</h1>
        <p className={styles.subtitle}>{message}</p>
        
        <div className={styles.progressContainer}>
          <div className={styles.progressBar}>
            <div 
              className={styles.progressFill} 
              style={{ width: `${progress.percentage}%` }}
            ></div>
          </div>
          <div className={styles.progressStats}>
            <span>{progress.label}</span>
            <span>{progress.percentage}%</span>
          </div>
        </div>

        <p className={styles.description}>
          This only happens once. We are downloading the necessary components to ensure your connection is secure and fast.
        </p>
      </div>
    </div>
  );
}
