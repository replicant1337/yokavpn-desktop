import styles from './ProxyInfo.module.css';
import { useTranslation } from '../../i18n';

export interface ProxyInfoData {
  ip: string;
  port: number;
  user?: string;
  pass?: string;
}

interface ProxyInfoProps {
  connected?: boolean;
  useTun?: boolean;
  systemProxy?: boolean;
  onSystemProxyChange?: (val: boolean) => void;
  proxyInfo?: ProxyInfoData;
}

export default function ProxyInfo({ 
  connected = false, 
  useTun = false, 
  systemProxy = true,
  onSystemProxyChange = () => {},
  proxyInfo = { ip: '127.0.0.1', port: 10808 } 
}: ProxyInfoProps) {
  const { t } = useTranslation();
  
  if (useTun) return null;

  const handleCopy = () => {
    let text = `${proxyInfo.ip}:${proxyInfo.port}`;
    if (proxyInfo.user && proxyInfo.pass) {
      text = `${proxyInfo.user}:${proxyInfo.pass}@${text}`;
    }
    navigator.clipboard.writeText(text);
  };

  return (
    <div className={styles.container}>
      <div className={styles.modeToggle}>
        <button 
          className={`${styles.toggleBtn} ${systemProxy ? styles.active : ''}`}
          onClick={() => !connected && onSystemProxyChange(true)}
          disabled={connected}
        >
          {t('app.system_proxy')}
        </button>
        <button 
          className={`${styles.toggleBtn} ${!systemProxy ? styles.active : ''}`}
          onClick={() => !connected && onSystemProxyChange(false)}
          disabled={connected}
        >
          {t('app.local_proxy')}
        </button>
      </div>

      {connected && (
        <div className={styles.proxyDetails}>
          <div className={styles.proxyInline}>
            <span className={styles.proxyType}>{systemProxy ? 'HTTP' : 'SOCKS5'}</span>
            <span className={styles.proxyIp}>
              {proxyInfo.ip}<span className={styles.colon}>:</span>{proxyInfo.port}
            </span>
            <button className={styles.copySmall} onClick={handleCopy} title={t('app.copy_all')}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="9" y="9" width="13" height="13" rx="2"/>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
            </button>
          </div>
          {proxyInfo.user && proxyInfo.pass && !systemProxy && (
            <div className={styles.credentials}>
              <code>{proxyInfo.user}</code>:<code>{proxyInfo.pass}</code>
            </div>
          )}
        </div>
      )}
    </div>
  );
}