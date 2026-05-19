import React from 'react';
import {Badge} from '@pharos/shared/components/ui';

export function getBgColorActive(title: string) {
  switch (title.toLowerCase()) {
    case 'critical':
      return `bg-[#B91C1C] text-white group-data-[state=active]:bg-[#B91C1C]`;
    case 'major':
      return 'bg-[#F97316] text-white group-data-[state=active]:bg-[#F97316]';
    case 'minor':
      return 'bg-[#EFA926] text-white group-data-[state=active]:bg-[#EFA926]';
    case 'low':
      return 'bg-[#0891B2] text-white group-data-[state=active]:bg-[#0891B2]';
    case 'normal':
      return 'bg-[#16A34A] text-white group-data-[state=active]:bg-[#16A34A]';
    default:
      return 'bg-[#64748B] text-white group-data-[state=active]:bg-[#64748B]';
  }
}

function getBgColor(title: string) {
  switch (title.toLowerCase()) {
    case 'critical':
      return '#B91C1C';
    case 'major':
      return '#F97316';
    case 'minor':
      return '#EFA926';
    case 'low':
      return '#0891B2';
    case 'normal':
      return '#16A34A';
    default:
      return '#64748B';
  }
}

function capitalize(str: string) {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

export default function AlertBadge({
  id,
  title,
  count,
  filterValue,
  onClick,
  className = '',
}: {
  id: string;
  title?: string;
  count?: number;
  filterValue?: string | null;
  onClick?: (value: string | null) => void;
  className?: string;
}) {
  const displayTitle = title || id;
  const bgColor = getBgColor(displayTitle);

  return (
    <Badge
      variant="outline"
      className={`${onClick ? 'cursor-pointer' : 'cursor-default'} border-none ${className}`}
      style={{
        minWidth: '56px',
        backgroundColor: bgColor,
        color: '#fff',
      }}
      onClick={() => onClick && onClick(id)}
    >
      {capitalize(displayTitle)} {count && count}
    </Badge>
  );
}
