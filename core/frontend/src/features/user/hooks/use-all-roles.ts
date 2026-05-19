import { useList } from '@/lib/data-provider';

import { RoleMetadata } from '@pharos/shared/types/role';
import { ROLE_PROVIDER_NAME } from '@providers/role-provider';

interface RoleGroup {
	group: string; // "core", "catv", etc.
	displayName: string; // "Core Permissions", "CATV Quality Monitor"
	roles: RoleMetadata[];
}

/**
 * Fetch all roles (Core + Extension) from backend
 *
 * 통합 API를 통해 Core role과 Extension role을 모두 조회합니다.
 * GET /api/roles
 *
 * @returns Query result with all roles grouped by source (core/extension)
 *
 * @example
 * const { data: roleGroups } = useAllRoles();
 *
 * roleGroups?.forEach(group => {
 *   console.log(group.displayName); // "Core Permissions", "CATV Quality Monitor"
 *   group.roles.forEach(role => {
 *     console.log(role.displayName); // "사용자 조회", "품질 지표 조회"
 *   });
 * });
 */
const EMPTY_ARRAY: never[] = [];

export function useAllRoles() {
	const { query } = useList<RoleGroup>({
		resource: 'all-roles',
		dataProviderName: ROLE_PROVIDER_NAME,
		queryOptions: {
			staleTime: 0,           // 항상 최신 데이터 사용 (role hide 설정 즉시 반영)
			gcTime: 5 * 60 * 1000, // 5분간 캐시 보존
		},
	});

	return {
		data: query.data?.data || EMPTY_ARRAY,
		isLoading: query.isLoading,
		error: query.error,
	};
}
