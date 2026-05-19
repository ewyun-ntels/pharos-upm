/**
 * CATV Extension permission keys (mirrors pkg/permissions/permissions.go)
 * Keep in sync with the Go constants.
 * Do NOT import this file from outside the catv extension.
 */
export const CATV_PERMISSIONS = {
  Read:   'extension:catv:read',
  Create: 'extension:catv:create',
  Update: 'extension:catv:update',
  Delete: 'extension:catv:delete',
} as const;

export type CatvPermission = (typeof CATV_PERMISSIONS)[keyof typeof CATV_PERMISSIONS];
