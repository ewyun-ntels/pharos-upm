import React from 'react';
import {useBreadcrumb} from '@/lib/data-provider/nav';
import {Link} from 'react-router-dom';
import {Folder} from 'lucide-react';
import {cn} from '@pharos/shared/lib/utils';
import {
  Breadcrumb as BreadcrumbUi,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
  useSidebar,
} from '@pharos/shared/components/ui';

/**
 * 모든 페이지에서 공통으로 사용하는 breadcrumb 컴포넌트.
 *
 * - 기본: useBreadcrumb()가 menuRegistry 기반으로 현재 경로를 자동 분석해 계층 생성
 * - currentLabel: 마지막 segment의 label을 override (동적 타이틀, UUID 제거 등)
 *   e.g. /dashboards/{uuid}/variables → currentLabel="대시보드명 / Variables"
 */
export const PageBreadcrumb: React.FC<{
  className?: string;
  currentLabel?: string;
}> = ({className, currentLabel}) => {
  const {open} = useSidebar();
  const {breadcrumbs} = useBreadcrumb(currentLabel);

  if (breadcrumbs.length === 0) return null;

  return (
    <div className={cn(!open && 'pl-6', className)}>
      <BreadcrumbUi>
        <BreadcrumbList>
          {breadcrumbs.map((crumb, index) => [
            index > 0 && <BreadcrumbSeparator key={`sep-${index}`} />,
            <BreadcrumbItem key={`${crumb.label}-${index}`}>
              {index === breadcrumbs.length - 1 ? (
                <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
              ) : crumb.href ? (
                <BreadcrumbLink asChild>
                  <Link to={crumb.href}>{crumb.label}</Link>
                </BreadcrumbLink>
              ) : (
                <span className="flex items-center text-muted-foreground font-normal">
                  {crumb.isFolder && <Folder className="w-3.5 h-3.5 mr-1.5" />}
                  {crumb.label}
                </span>
              )}
            </BreadcrumbItem>,
          ])}
        </BreadcrumbList>
      </BreadcrumbUi>
    </div>
  );
};
