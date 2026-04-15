import { useState } from 'preact/hooks';
import styles from './ModeSelector.module.css';
import { useTranslation } from '../../../i18n';

interface ModeSelectorProps {
  useTun?: boolean;
  disabled?: boolean;
  onChange?: (useTun: boolean) => void;
}

export default function ModeSelector({ useTun = false, disabled = false, onChange = () => {} }: ModeSelectorProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);

  const toggle = () => {
    if (disabled) return;
    setIsOpen(!isOpen);
  };

  const select = (val: boolean) => {
    onChange(val);
    setIsOpen(false);
  };

  return (
    <div className={styles.container}>
      <button className={styles.trigger} onClick={toggle} disabled={disabled}>
        <span>{useTun ? t('app.vpn_mode') : t('app.proxy_mode')}</span>
        <svg 
          width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"
          className={isOpen ? styles.iconOpen : ''}
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>

      {isOpen && (
        <>
          <div className={styles.overlay} onClick={() => setIsOpen(false)} />
          <div className={styles.dropdown}>
            <button 
              className={`${styles.option} ${useTun ? styles.active : ''}`}
              onClick={() => select(true)}
            >
              {t('app.vpn_mode')}
            </button>
            <button 
              className={`${styles.option} ${!useTun ? styles.active : ''}`}
              onClick={() => select(false)}
            >
              {t('app.proxy_mode')}
            </button>
          </div>
        </>
      )}
    </div>
  );
}