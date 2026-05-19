/**
 * QueryPipeline - 명확한 쿼리 실행 파이프라인
 * 
 * React의 reactive 패턴 대신 명시적인 제어 흐름 사용
 * - 비즈니스 로직을 클래스로 캡슐화
 * - 각 단계가 명확하게 분리됨
 * - 디버깅 용이
 * 
 * 사용법:
 *   const pipeline = QueryPipeline.for(id, queries).with({ graph, filterMetas, ... });
 *   if (pipeline.ready) {
 *     const result = await pipeline.execute();
 *   } else {
 *     console.log('Waiting:', pipeline.waiting);
 *   }
 */

import { DependencyGraph } from '../utils/dependency-graph';
import { extractVariables } from '@features/dashboard/utils/variable-parser';
import { dashboardProvider } from '@providers/dashboard-provider';
import { transformVariableValue } from '../variables/variable-value-transformer';
import { QUERY_PARAM_START_TIME, QUERY_PARAM_END_TIME, QUERY_PARAM_START_TIME_MS, QUERY_PARAM_END_TIME_MS } from '@lib/query-params';
import { getAbsoluteValueTimestamp, getAbsoluteValueTimestampMs } from '@pharos/shared/components/ui-extension';
import type { DateTimeRangeValue } from '@pharos/shared/components/ui-extension';
import type { ChartQueryRequest, ChartQueryArgs } from '@types';
import type { FilterMeta } from './use-dashboard-store';

// ============================================================================
// Types
// ============================================================================

export interface PipelineConfig {
  mode?: 'filter' | 'panel'; // filter: useDependencyGraph=true, panel: false
  graph?: DependencyGraph;
  filterMetas?: Map<string, FilterMeta>;
  startTime?: DateTimeRangeValue;
  endTime?: DateTimeRangeValue;
  enabled?: boolean;
  dashboardId?: string;
  resource?: 'timeseries' | 'raw';
  panelArgs?: Map<string, string | number | (() => number)>;
}

export interface DependencyInfo {
  /** 쿼리에서 직접 사용하는 변수들 */
  directVars: string[];
  /** 대기해야 하는 모든 변수들 (direct + ancestors) */
  requiredVars: string[];
  /** args에 포함할 변수들 (direct만) */
  argsVars: string[];
}

// ============================================================================
// QueryPipeline
// ============================================================================

export class QueryPipeline {
  private config: Required<PipelineConfig>;

  // Cached values (lazy computation)
  private _deps?: DependencyInfo;
  private _readyCheck?: { ok: boolean; waiting: string[] };
  private _args?: ChartQueryArgs;

  private constructor(
    private id: string,
    private queries: ChartQueryRequest[],
    config: PipelineConfig = {}
  ) {
    this.config = {
      mode: config.mode ?? 'panel',
      graph: config.graph,
      filterMetas: config.filterMetas ?? new Map(),
      startTime: config.startTime,
      endTime: config.endTime,
      enabled: config.enabled ?? true,
      dashboardId: config.dashboardId ?? '',
      resource: config.resource ?? 'timeseries',
    } as Required<PipelineConfig>;
  }

  // ============================================================================
  // Static Factory
  // ============================================================================

  static for(id: string, queries: ChartQueryRequest[]): QueryPipeline {
    return new QueryPipeline(id, queries);
  }

  // ============================================================================
  // Configuration
  // ============================================================================

  with(config: PipelineConfig): this {
    Object.assign(this.config, config);
    // Invalidate cache
    this._deps = undefined;
    this._readyCheck = undefined;
    this._args = undefined;
    return this;
  }

  // ============================================================================
  // Getters (Lazy Computed)
  // ============================================================================

  /**
   * 의존성 정보
   * - directVars: 쿼리에서 직접 사용하는 변수들
   * - requiredVars: 대기해야 하는 모든 변수들 (ancestors 포함)
   * - argsVars: args에 포함할 변수들
   */
  get deps(): DependencyInfo {
    if (!this._deps) {
      this._deps = this.computeDependencies();
    }
    return this._deps;
  }

  /**
   * 실행 준비 완료 여부
   */
  get ready(): boolean {
    if (!this._readyCheck) {
      this._readyCheck = this.checkReady();
    }
    return this._readyCheck.ok;
  }

  /**
   * 대기 중인 변수들
   */
  get waiting(): string[] {
    if (!this._readyCheck) {
      this._readyCheck = this.checkReady();
    }
    return this._readyCheck.waiting;
  }

  /**
   * 쿼리 args (변수 값 + 시간 범위)
   */
  get args(): ChartQueryArgs {
    if (!this._args) {
      this._args = this.buildArgs();
    }
    return this._args;
  }

  // ============================================================================
  // Execution
  // ============================================================================

  /**
   * 쿼리 실행
   * 
   * @throws Error if not ready
   */
  async execute<TData = unknown>(signal?: AbortSignal): Promise<TData> {
    if (!this.config.enabled) {
      throw new Error('Pipeline is disabled');
    }

    if (!this.ready) {
      throw new Error(`Not ready: waiting for ${this.waiting.join(', ')}`);
    }

    try {
      const result = await this.fetch(signal);
      return result as TData;
    } catch (err: any) {
      // Ignore abort errors
      if (err?.name === 'AbortError' || signal?.aborted) {
        throw err;
      }

      console.error(`[${this.id}] ❌ Error:`, err?.message || err);
      throw err;
    }
  }

  // ============================================================================
  // Private: Dependency Computation
  // ============================================================================

  private computeDependencies(): DependencyInfo {
    const { graph } = this.config;

    // 쿼리에서 변수 추출
    const allQueries = this.queries.map(q => q.query || '').join('\n');
    const directVars = extractVariables(allQueries);

    // Graph가 없거나 변수가 없으면 바로 반환
    if (!graph || directVars.length === 0) {
      return {
        directVars,
        requiredVars: directVars,
        argsVars: directVars,
      };
    }

    // Graph에서 transitive dependencies 추가
    // 쿼리에서 host를 사용하고, host가 datacenter → region 체인이면
    // host, datacenter, region 모두 완료될 때까지 기다려야 함
    const requiredSet = new Set<string>(directVars);
    directVars.forEach(varName => {
      const ancestors = graph.getAncestors(varName);
      ancestors.forEach(ancestor => requiredSet.add(ancestor));
    });

    return {
      directVars,
      requiredVars: Array.from(requiredSet),
      argsVars: directVars, // args에는 직접 사용하는 변수만
    };
  }

  // ============================================================================
  // Private: Ready Check
  // ============================================================================

  private checkReady(): { ok: boolean; waiting: string[] } {
    const { enabled, filterMetas } = this.config;

    if (!enabled) {
      return { ok: false, waiting: [] };
    }

    const { requiredVars } = this.deps;

    if (requiredVars.length === 0) {
      return { ok: true, waiting: [] };
    }

    const waiting: string[] = [];

    requiredVars.forEach(varName => {
      const meta = filterMetas.get(varName);

      // 대시보드에 등록되지 않은 변수는 대기하지 않음.
      // 백엔드 Pongo2의 |default: 필터가 폴백 처리하므로 프론트에서 block 불필요.
      if (!meta) return;

      const isFetching = meta.isFetching ?? false;
      const isSuccess = meta.isQuerySuccess ?? false;

      // 쿼리가 완료된 상태면 ready: 값이 있는 경우(정상)와 no data인 경우 모두 포함.
      // 값 없이 완료된 경우(no data)는 blockedByNoData에서 감지하여 쿼리 스킵 처리.
      const ready = !isFetching && isSuccess;
      if (!ready) {
        waiting.push(varName);
      }
    });

    return {
      ok: waiting.length === 0,
      waiting,
    };
  }

  // ============================================================================
  // Private: Args Building
  // ============================================================================

  private buildArgs(): ChartQueryArgs {
    const { filterMetas, startTime, endTime, panelArgs } = this.config;
    const { argsVars } = this.deps;

    const args = new Map<string, string | number | (() => number)>();

    // 1. Panel에서 전달한 args 먼저 추가 (step 등)
    if (panelArgs) {
      panelArgs.forEach((value, key) => {
        args.set(key, value);
      });
    }

    // 2. 변수 값 추가 (변환 레이어 사용)
    // ewyun-20260515: Prometheus datasource 변수는 regex 포맷(val1|val2) 사용
    // SQL datasource는 기존 'val1','val2' 포맷 유지
    // 패널 쿼리 datasource 기준으로 판단 (변수 datasource보다 더 신뢰할 수 있음)
    const isPanelPrometheus = this.queries.some(
      q => q.datasourceName?.toLowerCase().includes('prometheus')
    );
    argsVars.forEach(varName => {
      const meta = filterMetas.get(varName);
      if (meta?.value !== undefined) {
        const isPrometheus = isPanelPrometheus ||
          meta.datasourceName?.toLowerCase().includes('prometheus');
        const transformed = transformVariableValue(meta.value, {
          format: isPrometheus ? 'regex' : 'sql',
        });
        args.set(varName, transformed);
      }
    });

    // 3. 시간 범위 추가
    if (startTime) {
      args.set(QUERY_PARAM_START_TIME, () => getAbsoluteValueTimestamp(startTime));
      args.set(QUERY_PARAM_START_TIME_MS, () => getAbsoluteValueTimestampMs(startTime));
    } else {
      args.set(QUERY_PARAM_START_TIME, () => Math.floor((Date.now() - 3600000) / 1000));
      args.set(QUERY_PARAM_START_TIME_MS, () => Date.now() - 3600000);
    }

    if (endTime) {
      args.set(QUERY_PARAM_END_TIME, () => getAbsoluteValueTimestamp(endTime));
      args.set(QUERY_PARAM_END_TIME_MS, () => getAbsoluteValueTimestampMs(endTime));
    } else {
      args.set(QUERY_PARAM_END_TIME, () => Math.floor(Date.now() / 1000));
      args.set(QUERY_PARAM_END_TIME_MS, () => Date.now());
    }

    return args;
  }

  // ============================================================================
  // Private: Fetch
  // ============================================================================

  private async fetch(signal?: AbortSignal): Promise<unknown> {
    const { resource, dashboardId } = this.config;
    const args = this.args;

    // queries에 args 추가
    const chartQueryRequests = this.queries.map(q => ({
      ...q,
      dashboardId,
      args,
    }));

    // 모든 resource가 variables.value로 전달해야 함
    const meta = {
      variables: { value: chartQueryRequests },
      signal
    };

    const result = await dashboardProvider.getOne({
      resource,
      id: `${this.id}-query`,
      meta,
    });

    return result.data;
  }
}
