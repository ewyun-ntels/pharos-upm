

import { Button } from '@pharos/shared/components/ui';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { useList } from '@/lib/data-provider';
import { useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { t } from 'i18next';
import { LayoutGrid } from 'lucide-react';
import {UI_CONFIG_PROVIDER_NAME, UI_CONFIG_RESOURCES} from '@providers/ui-config-provider';

export default function Home() {
  const navigate = useNavigate();

  const { query: { data: homeData, isLoading } } = useList<{ dashboard_id: string }>({
    resource: UI_CONFIG_RESOURCES.HOME,
    dataProviderName: UI_CONFIG_PROVIDER_NAME,
  });

  const menuHome = (homeData?.data as any)?.dashboard_id ?? null;

  useEffect(() => {
    if (!isLoading && menuHome) {
      navigate(`/dashboards/${menuHome}`);
    }
  }, [isLoading, menuHome, navigate]);

  if (isLoading) {
    return (
      <LoadingIndicator className="h-screen" />
    );
  }

  if (!menuHome) {
    return (
      <div className="flex flex-1 h-screen flex-col justify-center items-center gap-6">
        <LayoutGrid className="w-16 h-16 text-muted-foreground/40" />
        <div className="flex flex-col items-center gap-2">
          <p className="text-lg font-medium">{t('label.home.noHome.title', '홈 대시보드가 설정되지 않았습니다')}</p>
          <p className="text-sm text-muted-foreground text-center">
            {t('label.home.noHome.description', '대시보드 목록에서 원하는 대시보드를 홈으로 설정해주세요.')}
          </p>
        </div>
        <Button onClick={() => navigate('/dashboards')}>
          {t('label.home.noHome.action', '대시보드 목록 보기')}
        </Button>
      </div>
    );
  }

  return null;
}
