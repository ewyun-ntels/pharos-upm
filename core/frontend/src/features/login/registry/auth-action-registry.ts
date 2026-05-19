import React from 'react';
import type { User } from '@pharos/shared/types/user';
import type {
  AuthActionPlugin,
  AuthActionBaseProps,
  AuthActionContext,
} from '@pharos/shared/features/auth';

/**
 * Lazy loader 타입 정의
 */
type LazyLoader<T> = () => Promise<{ default: React.ComponentType<T> }>;

/**
 * Lazy loader 여부 체크
 * 함수가 인자 없이 호출되고 Promise를 반환하면 lazy loader로 판단
 */
function isLazyLoader<T>(
  component: React.ComponentType<T> | LazyLoader<T>
): component is LazyLoader<T> {
  // React 컴포넌트는 일반적으로 props를 받으므로 length >= 1
  // Lazy loader는 () => Promise<...> 형태로 length === 0
  return typeof component === 'function' && component.length === 0;
}

/**
 * Auth Action Plugin Registry
 *
 * Dashboard Panel Plugin Registry와 동일한 패턴으로 설계됨
 */
export class AuthActionPluginRegistry {
  private plugins = new Map<string, AuthActionPlugin>();

  /**
   * 플러그인 등록
   * 동일 ID로 재등록 시 오버라이드됨 (Extension 커스터마이징 지원)
   *
   * NOTE: 제네릭 타입의 반공변성(contravariance)으로 인해
   * AuthActionPlugin<T>를 AuthActionPlugin에 직접 할당할 수 없음.
   * 레지스트리 패턴에서 이 캐스트는 안전함 (T extends AuthActionBaseProps 보장)
   */
  register<T extends AuthActionBaseProps>(plugin: AuthActionPlugin<T>): void {
    if (this.plugins.has(plugin.info.id)) {
      console.log(`[AuthAction] Overriding plugin: ${plugin.info.id}`);
    }
    this.plugins.set(plugin.info.id, plugin as AuthActionPlugin);
    console.log(`[AuthAction] Registered: ${plugin.info.id}`);
  }

  /**
   * 플러그인 조회
   */
  get(id: string): AuthActionPlugin | undefined {
    return this.plugins.get(id);
  }

  /**
   * 플러그인 존재 여부
   */
  exists(id: string): boolean {
    return this.plugins.has(id);
  }

  /**
   * 컴포넌트 로드 (lazy loading 지원)
   */
  async loadComponent(
    id: string
  ): Promise<React.ComponentType<AuthActionBaseProps> | null> {
    const plugin = this.get(id);
    if (!plugin) return null;

    try {
      const component = plugin.component;

      if (isLazyLoader(component)) {
        const imported = await component();
        return imported.default;
      }

      return component;
    } catch (error) {
      console.error(`[AuthAction] Failed to load component: ${id}`, error);
      return null;
    }
  }

  /**
   * 추가 Props 로드
   */
  async loadProps(
    id: string,
    user: User | null,
    context?: AuthActionContext
  ): Promise<Partial<AuthActionBaseProps>> {
    const plugin = this.get(id);
    if (!plugin?.getProps) return {};

    try {
      return await plugin.getProps(user, context);
    } catch (error) {
      console.error(`[AuthAction] Failed to load props: ${id}`, error);
      return {};
    }
  }

  /**
   * 플러그인 렌더링
   */
  render(
    id: string,
    props: AuthActionBaseProps,
    fallback?: React.ReactElement
  ): React.ReactElement {
    const registry = this;

    const ActionRenderer: React.FC = () => {
      const [extraProps, setExtraProps] = React.useState<Partial<AuthActionBaseProps>>(
        {}
      );
      const [loading, setLoading] = React.useState(true);

      React.useEffect(() => {
        registry.loadProps(id, props.user, props.context).then((extra) => {
          setExtraProps(extra);
          setLoading(false);
        });
      }, []);

      if (!registry.exists(id)) {
        return React.createElement(
          'div',
          { className: 'flex items-center justify-center h-full text-muted-foreground' },
          `Auth action "${id}" not found`
        );
      }

      if (loading) {
        return fallback || React.createElement('div', null, 'Loading...');
      }

      const LazyComponent = React.lazy(async () => {
        const component = await registry.loadComponent(id);
        if (!component) {
          throw new Error(`Failed to load auth action: ${id}`);
        }
        return { default: component };
      });

      const mergedProps: AuthActionBaseProps = { ...props, ...extraProps };

      return React.createElement(
        React.Suspense,
        { fallback: fallback || React.createElement('div', null, 'Loading...') },
        React.createElement(LazyComponent, mergedProps)
      );
    };

    return React.createElement(ActionRenderer);
  }

  /**
   * 모든 플러그인 제거
   */
  clear(): void {
    this.plugins.clear();
  }
}

export const authActionPluginRegistry = new AuthActionPluginRegistry();
