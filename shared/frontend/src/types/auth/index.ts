import * as z from "zod";


export const CommonSchema = z.object({
    "collect_client_info": z.boolean().optional(),
    "login_block_duration": z.number(),
    "login_retry_reset_timeout": z.number(),
});
export type Common = z.infer<typeof CommonSchema>;

export const PasswordPolicySchema = z.object({
    "min_digits": z.number(),
    "min_length": z.number(),
    "min_lowercase": z.number(),
    "min_special": z.number(),
    "min_uppercase": z.number(),
    "password_ttl": z.number().optional(),
});
export type PasswordPolicy = z.infer<typeof PasswordPolicySchema>;

export const UserPolicySchema = z.object({
    "email_allowed": z.boolean(),
    "max_length": z.number(),
    "min_length": z.number(),
});
export type UserPolicy = z.infer<typeof UserPolicySchema>;

export const AuthConfigSchema = z.object({
    "common": CommonSchema,
    "password": PasswordPolicySchema,
    "user": UserPolicySchema,
});
export type AuthConfig = z.infer<typeof AuthConfigSchema>;
