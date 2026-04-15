import styles from './Loader.module.css';

interface LoaderProps {
  size?: 'small' | 'medium' | 'large';
  color?: 'green' | 'white';
}

export default function Loader({ size = 'medium', color = 'green' }: LoaderProps) {
  const loaderClass = `${styles.loader} ${styles[size]} ${styles[color]}`.trim();
  
  return <div className={loaderClass}></div>;
}