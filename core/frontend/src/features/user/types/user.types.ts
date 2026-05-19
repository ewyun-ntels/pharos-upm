import type { User, Role } from '@pharos/shared/types/user';
import type { PermissionKeys } from '@pharos/shared/types/role';

/**
 * User feature types
 * 
 * Backend schema를 기반으로 한 타입 정의
 * shared/schema/user/User.schema에서 생성됨
 */

// Permission item used in UI selections (multiselect 등)
export type PermissionItem = { 
  id: PermissionKeys; 
  label: string 
};

// Table row data (User + UI 확장)
export type UserRowData = User & {
  // 테이블 UI를 위한 추가 필드
  permission?: PermissionItem[];
  rowData?: UserRowData; // nested row callback용
};

// Form data (생성/편집)
export type UserFormData = {
  name: string;
  password?: string;
  roles?: Role[];
  blocked?: boolean;
};

// Re-export shared types
export type { User, Role };
