import styles from './ConnectButton.module.css';
import { useTranslation } from '../../../i18n';

interface ConnectButtonProps {
  connected?: boolean;
  connecting?: boolean;
  isDisconnecting?: boolean;
  statusText?: string;
  onClick?: () => void;
}

export default function ConnectButton({ 
  connected = false, 
  connecting = false, 
  isDisconnecting = false,
  statusText = '',
  onClick = () => {} 
}: ConnectButtonProps) {
  const { t } = useTranslation();
  
  const statusClass = isDisconnecting
    ? styles.connecting 
    : connecting 
      ? styles.connecting 
      : connected 
        ? styles.connected 
        : styles.disconnected;

  const getLabel = () => {
    if (isDisconnecting) return t('app.disconnecting');
    if (connecting) {
      if (statusText) {
        const translated = t(statusText);
        return translated !== statusText ? translated : statusText;
      }
      return t('app.connecting');
    }
    return connected ? t('app.disconnect') : t('app.connect');
  };

  return (
    <div className={styles.wrapper}>
      <button 
        className={`${styles.button} ${statusClass}`} 
        onClick={onClick}
        disabled={connecting || isDisconnecting}
      >
        <div className={styles.ripple}></div>
        <div className={styles.content}>
          {connecting || isDisconnecting ? (
            <svg className={styles.iconSpin} viewBox="0 0 24 24" fill="none" stroke="#000000" strokeWidth="3">
              <path d="M1 4v6h6M23 20v-6h-6M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" />
            </svg>
          ) : connected ? (
            <svg className={styles.icon} viewBox="0 0 24 24" fill="none" stroke="#000000" strokeWidth="3">
              <rect x="6" y="6" width="12" height="12" rx="2" fill="#000000" />
            </svg>
          ) : (
            <svg className={styles.icon} viewBox="0 0 24 24" fill="#000000">
              <path d="M8 5v14l11-7z" />
            </svg>
          )}
        </div>
      </button>
      <span className={`${styles.label} ${statusClass}`}>
        {getLabel()}
      </span>
    </div>
  );
}
