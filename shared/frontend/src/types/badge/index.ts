import * as z from "zod";


export const QuerySpecSchema = z.object({
    "query": z.string(),
    "time_label": z.string().optional(),
    "variable_label": z.string().optional(),
    "variables": z.record(z.string(), z.any()).optional(),
});
export type QuerySpec = z.infer<typeof QuerySpecSchema>;

export const BadgeValueResponseSchema = z.object({
    "cache_hit": z.boolean().optional(),
    "last_updated": z.coerce.date().optional(),
    "name": z.string(),
    "value": z.any(),
});
export type BadgeValueResponse = z.infer<typeof BadgeValueResponseSchema>;

export const BadgeSpecSchema = z.object({
    "cache_ttl": z.number().optional(),
    "column": z.string(),
    "datasource": z.string(),
    "name": z.string(),
    "query": QuerySpecSchema,
});
export type BadgeSpec = z.infer<typeof BadgeSpecSchema>;
