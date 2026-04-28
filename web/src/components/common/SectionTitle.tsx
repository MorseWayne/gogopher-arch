import type { LucideIcon } from 'lucide-react';
import styles from './SectionTitle.module.css';

interface SectionTitleProps {
  icon: LucideIcon;
  label: string;
}

export function SectionTitle({ icon: Icon, label }: SectionTitleProps) {
  return (
    <div className={styles.sectionTitle}>
      <Icon size={16} />
      <span>{label}</span>
    </div>
  );
}
