import * as z from "zod";


export const BlockUserRequestSchema = z.object({
    "block": z.boolean(),
});
export type BlockUserRequest = z.infer<typeof BlockUserRequestSchema>;

export const ChangeMyPasswordRequestSchema = z.object({
    "current_password": z.string(),
    "new_password": z.string(),
});
export type ChangeMyPasswordRequest = z.infer<typeof ChangeMyPasswordRequestSchema>;

export const ChangePasswordRequestSchema = z.object({
    "password": z.string(),
});
export type ChangePasswordRequest = z.infer<typeof ChangePasswordRequestSchema>;

export const AttributesSchema = z.object({
    "info": z.record(z.string(), z.any()).optional(),
    "roles": z.record(z.string(), z.boolean()).optional(),
});
export type Attributes = z.infer<typeof AttributesSchema>;

export const RoleSchema = z.object({
    "description": z.string().optional(),
    "display_name": z.string(),
    "group": z.string().optional(),
    "role": z.string(),
});
export type Role = z.infer<typeof RoleSchema>;

export const SetAttributesRequestSchema = z.object({
    "attributes": AttributesSchema,
});
export type SetAttributesRequest = z.infer<typeof SetAttributesRequestSchema>;

export const UpdatePasswordExpirationRequestSchema = z.object({
    "expired_at": z.union([z.coerce.date(), z.null()]),
});
export type UpdatePasswordExpirationRequest = z.infer<typeof UpdatePasswordExpirationRequestSchema>;

export const UserMetadataConfigSchema = z.object({
    "schema": z.record(z.string(), z.any()),
    "uiSchema": z.record(z.string(), z.any()).optional(),
});
export type UserMetadataConfig = z.infer<typeof UserMetadataConfigSchema>;

export const CreateUserRequestSchema = z.object({
    "attributes": AttributesSchema.optional(),
    "password": z.string(),
    "username": z.string(),
});
export type CreateUserRequest = z.infer<typeof CreateUserRequestSchema>;

export const UserSchema = z.object({
    "attributes": AttributesSchema.optional(),
    "blocked": z.union([z.boolean(), z.null()]).optional(),
    "created_at": z.union([z.coerce.date(), z.null()]).optional(),
    "is_me": z.boolean().optional(),
    "name": z.string(),
    "password_expired_at": z.union([z.coerce.date(), z.null()]).optional(),
    "prepare": z.array(z.string()).optional(),
    "prepare_labels": z.array(z.string()).optional(),
    "roles": z.array(RoleSchema).optional(),
    "temporary_blocked_expires_at": z.union([z.coerce.date(), z.null()]).optional(),
});
export type User = z.infer<typeof UserSchema>;

export const ListUserResponseSchema = z.object({
    "users": z.array(UserSchema).optional(),
});
export type ListUserResponse = z.infer<typeof ListUserResponseSchema>;

export const GetUserResponseSchema = z.object({
    "user": UserSchema,
});
export type GetUserResponse = z.infer<typeof GetUserResponseSchema>;
