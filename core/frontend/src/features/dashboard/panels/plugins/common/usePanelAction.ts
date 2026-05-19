import { useCallback } from 'react';
import type { PanelAction } from '@pharos/core/panel-registry';

/**
 * usePanelAction
 *
 * 플러그인 컴포넌트에서 액션을 emit하기 위한 공통 훅.
 * 베이스(PanelRenderer)는 액션 타입을 모르고, 플러그인이 type/payload를 소유.
 *
 * 사용 예:
 *   const dispatch = usePanelAction(onPanelAction);
 *   dispatch('resize-column', { colId: 'name', width: 120 });
 */
export function usePanelAction(
  onPanelAction?: (action: PanelAction) => void,
) {
  return useCallback(
    <TPayload = unknown>(type: string, payload: TPayload) => {
      onPanelAction?.({ type, payload });
    },
    [onPanelAction],
  );
}
