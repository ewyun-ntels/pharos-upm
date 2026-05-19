import * as z from "zod";


export const StatisticsSchema = z.object({
    "elapsed": z.number().optional(),
});
export type Statistics = z.infer<typeof StatisticsSchema>;

export const IndexSchema = z.object({
    "data": z.array(z.record(z.string(), z.any())),
    "meta": z.array(z.record(z.string(), z.string())),
    "rows": z.number(),
    "sql": z.string().optional(),
    "statistics": StatisticsSchema.optional(),
});
export type Index = z.infer<typeof IndexSchema>;
