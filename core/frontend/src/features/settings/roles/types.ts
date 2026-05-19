// RoleMetadata, RoleGroup — JSON Schema에서 quicktype으로 자동 생성된 타입
export type {RoleMetadata, RoleGroup} from '@pharos/shared/types/role';

// 내부 편집 상태 — schema에 없는 UI 전용 타입
export type AtomicEditState = {
  displayName: string;
  description: string;
  hide: boolean;
};
