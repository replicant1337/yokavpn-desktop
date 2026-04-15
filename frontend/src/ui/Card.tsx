import { ComponentChildren } from 'preact';

interface CardProps {
  children: ComponentChildren;
  className?: string;
  onClick?: () => void;
}

export default function Card({ children, className = '', onClick }: CardProps) {
  return (
    <div className={className} onClick={onClick} style={{ cursor: onClick ? 'pointer' : undefined }}>
      {children}
    </div>
  );
}
