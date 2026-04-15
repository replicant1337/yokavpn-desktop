interface BadgeProps {
  text: string;
}

export default function Badge({ text }: BadgeProps) {
  return (
    <span style={{
      background: 'rgba(255,255,255,0.06)',
      padding: '2px 8px',
      borderRadius: '6px',
      fontSize: '10px',
      fontWeight: 700,
      color: 'var(--text-muted)',
      textTransform: 'uppercase',
      letterSpacing: '0.5px'
    }}>
      {text}
    </span>
  );
}
