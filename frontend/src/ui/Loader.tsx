interface LoaderProps {
  size?: 'small' | 'medium' | 'large';
  color?: string;
}

const sizes = { small: 16, medium: 24, large: 40 };

export default function Loader({ size = 'medium', color = 'currentColor' }: LoaderProps) {
  const s = sizes[size];
  return (
    <svg width={s} height={s} viewBox="0 0 24 24" fill="none" style={{ animation: 'spin 1s linear infinite' }}>
      <circle cx="12" cy="12" r="10" stroke={color} strokeWidth="3" strokeDasharray="32" strokeLinecap="round" opacity="0.3" />
      <path d="M12 2a10 10 0 0 1 10 10" stroke={color} strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}
