import * as z from "zod";

// All available permission keys in the system

export const PermissionKeysSchema = z.enum([
    "attr:password_change",
    "attr:password_retry_unlimited",
    "attr:skip_temporarily_block",
    "attr:temporary_user",
    "role:alert_create",
    "role:alert_delete",
    "role:alert_read",
    "role:alert_update",
    "role:dashboard_create",
    "role:dashboard_read",
    "role:group_add",
    "role:notification_create",
    "role:notification_delete",
    "role:notification_read",
    "role:notification_update",
    "role:super_admin",
    "role:user_create",
    "role:user_delete",
    "role:user_read",
    "role:user_update",
]);
export type PermissionKeys = z.infer<typeof PermissionKeysSchema>;

export const RoleMetadataSchema = z.object({
    "description": z.string(),
    "displayName": z.string(),
    "group": z.string(),
    "hide": z.boolean().optional(),
    "key": z.string(),
    "roles": z.record(z.string(), z.boolean()).optional(),
});
export type RoleMetadata = z.infer<typeof RoleMetadataSchema>;

export const RoleMetadataConfigEntitySchema = z.object({
    "config": z.array(RoleMetadataSchema),
    "created_at": z.string(),
    "description": z.string().optional(),
    "id": z.string(),
    "updated_by": z.string().optional(),
});
export type RoleMetadataConfigEntity = z.infer<typeof RoleMetadataConfigEntitySchema>;

export const RoleMetadataConfigResponseSchema = z.object({
    "config": z.array(RoleMetadataSchema).optional(),
    "message": z.string().optional(),
});
export type RoleMetadataConfigResponse = z.infer<typeof RoleMetadataConfigResponseSchema>;

export const RoleGroupSchema = z.object({
    "displayName": z.string(),
    "group": z.string(),
    "roles": z.array(RoleMetadataSchema),
});
export type RoleGroup = z.infer<typeof RoleGroupSchema>;

export const RoleMetadataConfigHistoryResponseSchema = z.object({
    "history": z.array(RoleMetadataConfigEntitySchema),
});
export type RoleMetadataConfigHistoryResponse = z.infer<typeof RoleMetadataConfigHistoryResponseSchema>;
