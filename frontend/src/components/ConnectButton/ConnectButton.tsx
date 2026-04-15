import styles from './ConnectButton.module.css';
import { useTranslation } from '../../i18n';

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
    ? styles.connecting // Use same style as connecting for now
    : connecting 
      ? styles.connecting 
      : connected 
        ? styles.connected 
        : styles.disconnected;

  const getLabel = () => {
    if (isDisconnecting) {
      return t('app.disconnecting');
    }
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
          {connecting ? (
            <svg className={styles.iconSpin} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
              <path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83" />
            </svg>
          ) : connected ? (
            <svg className={styles.icon} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
              <rect x="6" y="6" width="12" height="12" rx="2" fill="currentColor" />
            </svg>
          ) : (
            <svg className={styles.icon} viewBox="0 0 24 24" fill="currentColor">
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
