import styles from './ConnectionStatus.module.css';
import Loader from '../../ui/Loader';

interface ConnectionStatusProps {
  text?: string;
  connecting?: boolean;
}

export default function ConnectionStatus({ text = '', connecting = false }: ConnectionStatusProps) {
  if (!text && !connecting) return null;

  return (
    <div className={styles.status}>
      {connecting && <Loader size="small" color="green" />}
      {text && <span>{text}</span>}
    </div>
  );
}