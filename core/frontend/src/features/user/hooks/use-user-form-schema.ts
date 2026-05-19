import { useMemo } from 'react';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import type { UserMetadataConfig } from '@pharos/shared/types/user';

type RoleMetadata = { key: string; displayName: string; description: string };
type RoleGroup = { group: string; displayName: string; roles: RoleMetadata[] };

export function useUserFormSchema({
  roleGroups,
  isSuperAdmin,
  userMetadataConfig,
  isLoadingRoles,
  isLoadingUserMetadata,
}: {
  roleGroups: RoleGroup[] | undefined;
  isSuperAdmin: boolean;
  userMetadataConfig: UserMetadataConfig | null;
  isLoadingRoles: boolean;
  isLoadingUserMetadata: boolean;
}) {
  const schema = useMemo(() => {
    if (isLoadingRoles || isLoadingUserMetadata) {
      return { type: 'object', properties: {}, additionalProperties: false };
    }

    const permissionProperties: Record<string, any> = {};
    if (roleGroups) {
      roleGroups.forEach((group) => {
        group.roles.forEach((role) => {
          if (role.key === PermissionKeysSchema.enum['role:super_admin'] && !isSuperAdmin) return;
          const title =
            group.group === 'core'
              ? role.displayName
              : `[${group.displayName}] ${role.displayName}`;
          permissionProperties[role.key] = {
            type: 'boolean',
            title,
            default: false,
            description: role.description,
          };
        });
      });
    }

    const properties: Record<string, any> = {
      permissions: { title: 'Role', type: 'object', properties: permissionProperties },
    };
    if (userMetadataConfig?.schema) {
      properties.userinfo = userMetadataConfig.schema;
    }

    return { type: 'object', properties, additionalProperties: false };
  }, [roleGroups, isSuperAdmin, userMetadataConfig, isLoadingRoles, isLoadingUserMetadata]);

  const uiSchema = useMemo(() => {
    const base = {
      permissions: { 'ui:widget': 'hidden' },
      'ui:submitButtonOptions': { norender: true },
    };
    if (userMetadataConfig?.uiSchema) {
      return { ...base, userinfo: userMetadataConfig.uiSchema };
    }
    return base;
  }, [userMetadataConfig]);

  return { schema, uiSchema };
}
