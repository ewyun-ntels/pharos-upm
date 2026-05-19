import React from 'react';
import {useNavigate} from 'react-router-dom';
import {useDashboardStore} from './use-dashboard-store';
import {unmountDashboardQueries} from './useDashboardQuery';

/**
 * Dashboard 페이지에서 다른 페이지로 이동 시 저장되지 않은 변경사항을 체크하는 Hook
 * 
 * @param dashboardId 현재 대시보드 ID
 */
export const useDashboardNavigationGuard = (dashboardId: string | null | undefined) => {
  const navigate = useNavigate();

  // ✅ 브라우저 새로고침/닫기 감지
  React.useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      const store = useDashboardStore.getState();
      if (store.hasUnsavedChanges) {
        e.preventDefault();
        e.returnValue = '';
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, []);

  // ✅ SPA 내부 링크 클릭 감지
  React.useEffect(() => {
    let isConfirming = false; // 중복 확인 방지 플래그
    let confirmedNavigation = false; // 확인 완료 후 네비게이션 진행 중 플래그

    const handleClick = (e: MouseEvent) => {
      // 이미 확인 중이거나, 확인 완료 후 네비게이션 중이면 건너뛰기
      if (isConfirming || confirmedNavigation) return;

      const target = e.target as HTMLElement;
      const link = target.closest('a');

      // 링크가 없으면 무시
      if (!link || !link.href) return;

      // 외부 링크나 다운로드 링크는 무시
      if (link.target === '_blank' || link.download || link.href.startsWith('blob:')) {
        return;
      }

      // 링크의 dashboard-id를 URL path에서 추출 (RESTful + Legacy 지원)
      const linkUrl = new URL(link.href);
      let linkDashboardId = null;

      // RESTful pattern: /dashboards/{dashboardId}[/...]
      const restfulMatch = linkUrl.pathname.match(/\/(?:ui\/)?dashboards\/([^\/]+)(?:\/|$)/);
      if (restfulMatch) {
        linkDashboardId = restfulMatch[1];
      }

      // No longer supporting legacy patterns - keeping for historical reference
      // const legacyMatch = linkUrl.pathname.match(/\/(?:ui\/)?dashboards\/(?:show|edit|variables|variableSettings)\/([^\/\?]+)/);
      // if (legacyMatch) {
      //   linkDashboardId = legacyMatch[1];
      // }

      // 같은 대시보드 내 이동이면 팝업 안 띄움 (Edit, Show, Variables, VariableSettings 등)
      if (linkDashboardId === dashboardId) {
        return;
      }

      const store = useDashboardStore.getState();

      // 미저장 변경사항 없음: 클릭 즉시 쿼리 abort 후 정상 네비게이션
      // (unmount cleanup은 새 페이지 렌더링 이후라 lazy chunk 다운로드가 이미 블로킹된 뒤임)
      if (!store.hasUnsavedChanges) {
        unmountDashboardQueries();
        return;
      }

      // 미저장 변경사항 있음: 먼저 이벤트 차단 후 confirm 후 abort
      // (취소 시 쿼리가 이미 abort된 상태가 되는 사이드이펙트 방지)
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();

      // 이미 확인 중이면 중복 실행 방지
      if (isConfirming) return;
      isConfirming = true;

      // 비동기로 confirm 실행 (이벤트 루프 분리)
      setTimeout(() => {
        const confirmed = window.confirm(
          '저장하지 않은 변경사항이 있습니다. 페이지를 나가시겠습니까?'
        );

        if (confirmed) {
          confirmedNavigation = true;

          // 확인 후에 abort (취소 시엔 abort 안 함 → 패널 계속 정상 동작)
          unmountDashboardQueries();

          // Store 초기화하고 페이지 이동 (변경사항 버림)
          store.reset();

          const url = new URL(link.href);
          const pathname = url.pathname.replace(/^\/ui/, '');
          navigate(pathname + url.search);
        }

        // 플래그 리셋
        isConfirming = false;
      }, 0);
    };

    document.addEventListener('click', handleClick, true);
    return () => {
      document.removeEventListener('click', handleClick, true);
      isConfirming = false;
      confirmedNavigation = false;
    };
  }, [dashboardId, navigate]);
};
