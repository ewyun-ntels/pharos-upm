import React from 'react';
import {Title} from '@pharos/shared/components/ui-extension';
import {PageBreadcrumb} from '@components/breadcrumb';

interface PageHeaderProps {
  title: React.ReactNode;
  /** 동적 타이틀이 필요할 때만 지정. 없으면 route 기반 자동 생성 */
  breadcrumbLabel?: string;
  actions?: React.ReactNode;
}

export function PageHeader({title, breadcrumbLabel, actions}: PageHeaderProps) {
  return (
    <header className="flex flex-col items-start w-full gap-3 px-5 pt-3">
      <PageBreadcrumb currentLabel={breadcrumbLabel} />
      <div className="flex items-center justify-between w-full">
        <Title>{title}</Title>
        {actions}
      </div>
    </header>
  );
}
