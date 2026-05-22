/**
 * Dependency Graph — Variable 의존성 그래프 관리
 * 
 * - 자동 의존성 추출 (variable-parser 사용)
 * - 순환 참조 감지 (DFS)
 * - 토폴로지 정렬 (Kahn's algorithm)
 */

import { extractGrafanaVariables, extractVariables, isGrafanaClassicVariableQuery } from './variable-parser';
import { isPrometheusDatasource } from './datasource';

/**
 * Variable 인터페이스 — 의존성 분석용
 * id: variable 식별자
 * query: 의존성을 추출할 쿼리 문자열
 */
export interface Variable {
  id: string;
  query: string;
  datasourceName?: string;
  datasourceType?: string;
}

export class DependencyGraph {
  // id -> 이 variable에 의존하는 variables (dependents)
  private adjacencyList = new Map<string, Set<string>>();
  
  // id -> 이 variable이 의존하는 variables (dependencies)
  private reverseDeps = new Map<string, Set<string>>();

  /**
   * Variable 목록으로부터 의존성 그래프 구축
   */
  buildFromVariables(variables: Variable[]): void {
    this.adjacencyList.clear();
    this.reverseDeps.clear();

    // 모든 variable ID 등록
    variables.forEach(v => {
      this.adjacencyList.set(v.id, new Set());
      this.reverseDeps.set(v.id, new Set());
    });

    // 의존성 추출 및 그래프 구축
    variables.forEach(variable => {
      const deps = isPrometheusDatasource(variable.datasourceName || '', variable.datasourceType) ||
        isGrafanaClassicVariableQuery(variable.query)
        ? [...new Set([...extractVariables(variable.query), ...extractGrafanaVariables(variable.query)])]
        : extractVariables(variable.query);
      
      deps.forEach(depId => {
        // depId가 실제 존재하는 variable인 경우만 의존성 추가
        if (this.adjacencyList.has(depId)) {
          // variable은 depId에 의존
          this.reverseDeps.get(variable.id)?.add(depId);
          
          // depId는 variable에 의해 사용됨 (dependent)
          this.adjacencyList.get(depId)?.add(variable.id);
        }
      });
    });
  }

  /**
   * 순환 참조 감지 (DFS)
   * @returns 순환 경로 배열 (없으면 빈 배열)
   */
  detectCycles(): string[][] {
    const visited = new Set<string>();
    const recursionStack = new Set<string>();
    const cycles: string[][] = [];

    const dfs = (nodeId: string, path: string[]): void => {
      if (recursionStack.has(nodeId)) {
        // 순환 발견: 순환 시작 지점부터 현재까지
        const cycleStart = path.indexOf(nodeId);
        cycles.push([...path.slice(cycleStart), nodeId]);
        return;
      }
      
      if (visited.has(nodeId)) return;
      
      visited.add(nodeId);
      recursionStack.add(nodeId);
      
      this.adjacencyList.get(nodeId)?.forEach(dependent => {
        dfs(dependent, [...path, nodeId]);
      });
      
      recursionStack.delete(nodeId);
    };

    this.adjacencyList.forEach((_, id) => {
      if (!visited.has(id)) {
        dfs(id, []);
      }
    });

    return cycles;
  }

  /**
   * 토폴로지 정렬 — 의존성 우선 로딩 순서 (Kahn's algorithm)
   * @returns 로딩 순서 배열
   * @throws 순환 참조가 있으면 에러
   */
  getLoadOrder(): string[] {
    const indegree = new Map<string, number>();
    const result: string[] = [];
    const queue: string[] = [];

    // indegree 계산 (각 노드가 의존하는 개수)
    this.adjacencyList.forEach((_, id) => {
      indegree.set(id, this.reverseDeps.get(id)?.size || 0);
    });

    // indegree가 0인 노드들로 시작 (의존성 없음)
    indegree.forEach((deg, id) => {
      if (deg === 0) queue.push(id);
    });

    while (queue.length > 0) {
      const current = queue.shift()!;
      result.push(current);

      // current에 의존하는 노드들의 indegree 감소
      this.adjacencyList.get(current)?.forEach(dependent => {
        const nextDegree = indegree.get(dependent)! - 1;
        indegree.set(dependent, nextDegree);
        
        if (nextDegree === 0) {
          queue.push(dependent);
        }
      });
    }

    // 모든 노드가 처리되지 않았으면 순환 참조 존재
    if (result.length !== this.adjacencyList.size) {
      throw new Error('Circular dependency detected');
    }

    return result;
  }

  /**
   * 특정 variable이 의존하는 variables (dependencies)
   */
  getDependencies(id: string): string[] {
    return [...(this.reverseDeps.get(id) || [])];
  }

  /**
   * 특정 variable에 의존하는 variables (dependents)
   */
  getDependents(id: string): string[] {
    return [...(this.adjacencyList.get(id) || [])];
  }

  /**
   * 특정 variable의 모든 조상(ancestors) 가져오기 (재귀)
   * Grafana 방식: queryKey에 모든 조상 값 포함하여 cascade refetch
   */
  getAncestors(id: string): string[] {
    const visited = new Set<string>();
    const ancestors: string[] = [];

    const dfs = (currentId: string) => {
      const deps = this.reverseDeps.get(currentId) || new Set();
      deps.forEach(depId => {
        if (!visited.has(depId)) {
          visited.add(depId);
          ancestors.push(depId);
          dfs(depId); // 재귀
        }
      });
    };

    dfs(id);
    return ancestors;
  }

  /**
   * 특정 variable의 모든 후손(descendants/dependents) 가져오기 (재귀)
   * Cascade 시 모든 하위 변수들을 초기화하기 위해 사용
   */
  getAllDependents(id: string): string[] {
    const visited = new Set<string>();
    const descendants: string[] = [];

    const dfs = (currentId: string) => {
      const deps = this.adjacencyList.get(currentId) || new Set();
      deps.forEach(depId => {
        if (!visited.has(depId)) {
          visited.add(depId);
          descendants.push(depId);
          dfs(depId); // 재귀
        }
      });
    };

    dfs(id);
    return descendants;
  }

  /**
   * 그래프 통계
   */
  getStats() {
    return {
      totalNodes: this.adjacencyList.size,
      edges: [...this.adjacencyList.values()].reduce((sum, deps) => sum + deps.size, 0),
    };
  }
}
