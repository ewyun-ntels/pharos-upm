import {useState, useEffect, useRef} from 'react';
import {panelPluginRegistry} from '@pharos/core/panel-registry';
import type {Panel, Action} from '@pharos/shared/types/dashboard';

/**
 * transformedData SWR 캐시 — 패널 remount 시 스피너 제거
 *
 * - 키: `${panel.id}:${panel.renderType}:${JSON.stringify(panel.dataProvider)}:${permission}`
 * - 에디터에서 dataProvider/renderType을 변경하면 키가 달라지므로 자동 무효화
 * - 변경 없이 에디터 ↔ 대시보드 이동 시 캐시 히트 → transformedData 즉시 복원
 *
 * 메모리: 패널 설정 객체(수KB)를 저장하므로 MAX_TRANSFORMED_CACHE로 항목 수 제한 (LRU)
 */
const MAX_TRANSFORMED_CACHE = 200;
const transformedDataCache = new Map<string, unknown>();

function setTransformedCache(key: string, value: unknown) {
  if (transformedDataCache.has(key)) transformedDataCache.delete(key);
  transformedDataCache.set(key, value);
  if (transformedDataCache.size > MAX_TRANSFORMED_CACHE) {
    transformedDataCache.delete(transformedDataCache.keys().next().value!);
  }
}

interface UseTransformedPanelDataOptions {
  panel: Panel | undefined;
  permission: Action;
  /** true: panel.options 변경 시 toPanelData 재실행 (에디터 전용) */
  watchOptions?: boolean;
  /** true: LRU 캐시 사용 (대시보드 전용, remount 시 즉시 복원) */
  enableCache?: boolean;
}

interface UseTransformedPanelDataResult {
  /** toPanelData() 결과. enableCache=true일 때 fallback은 panel 원본, false일 때 빈 객체 */
  transformedData: any;
  /** 초기 로딩 완료 여부 (undefined가 아닌지) */
  isResolved: boolean;
}

/**
 * toPanelData() 호출 로직을 공통화하는 훅
 *
 * PanelRenderer, LeftTopPanel 양쪽에서 동일한 패턴으로 사용:
 * - panelPluginRegistry.loadEditorConfig → toPanelData 실행
 * - fallback 처리 (toPanelData 없는 플러그인)
 * - error 처리
 * - isMounted 패턴으로 메모리 누수 방지
 *
 * 모드별 차이는 파라미터로 제어:
 * - watchOptions: 에디터에서 Fields 속성 변경 시 즉시 미리보기 반영
 * - enableCache: 대시보드에서 remount 시 스피너 제거 (SWR 패턴)
 */
export function useTransformedPanelData({
  panel,
  permission,
  watchOptions = false,
  enableCache = false,
}: UseTransformedPanelDataOptions): UseTransformedPanelDataResult {

  // SWR: 캐시에서 초기값 복원 (enableCache일 때만)
  // options도 캐시 키에 포함 — columnDataLinks 등 options 변경 시 캐시 무효화 보장
  const cacheKey = enableCache && panel
    ? `${panel.id}:${panel.renderType}:${JSON.stringify(panel.dataProvider)}:${JSON.stringify(panel.options)}:${permission}`
    : null;
  const cachedData = cacheKey ? transformedDataCache.get(cacheKey) ?? null : null;

  const [transformedData, setTransformedData] = useState<any>(
    cachedData !== null ? cachedData : undefined,
  );

  // watchOptions 변경에 의한 불필요한 effect 재실행 방지
  const watchOptionsRef = useRef(watchOptions);
  watchOptionsRef.current = watchOptions;

  useEffect(() => {
    if (!panel?.renderType) {
      setTransformedData(enableCache ? null : undefined);
      return;
    }

    let isMounted = true;

    panelPluginRegistry
      .loadEditorConfig(panel.renderType)
      .then(async (editorConfig) => {
        if (!isMounted) return;

        if (!editorConfig?.toPanelData) {
          // toPanelData 없는 플러그인
          // - 대시보드(enableCache): 원본 panel 사용
          // - 에디터(!enableCache): 빈 객체로 확정 (기본값 사용)
          const fallback = enableCache ? panel : {};
          if (isMounted) setTransformedData(fallback);
          return;
        }

        try {
          const result = await editorConfig.toPanelData(panel, permission);
          if (isMounted) {
            if (cacheKey) setTransformedCache(cacheKey, result);
            setTransformedData(result);
          }
        } catch (error) {
          console.error(
            `[useTransformedPanelData] Failed to transform panel data for renderType: ${panel.renderType}`,
            error,
          );
          if (isMounted) {
            setTransformedData(enableCache ? panel : {}); // fallback
          }
        }
      })
      .catch((error) => {
        console.error(
          `[useTransformedPanelData] Failed to load editor config for renderType: ${panel?.renderType}`,
          error,
        );
        if (isMounted) {
          setTransformedData(enableCache ? panel : {}); // fallback
        }
      });

    return () => {
      isMounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    panel?.id,
    panel?.renderType,
    panel?.dataProvider,
    permission,
    enableCache,
    // watchOptions=true일 때만 panel.options가 변경되면 재실행
    // eslint-disable-next-line react-hooks/exhaustive-deps
    ...(watchOptions ? [panel?.options] : []),
  ]);

  return {
    transformedData,
    isResolved: transformedData !== undefined,
  };
}
