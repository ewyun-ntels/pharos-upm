import * as z from "zod";


export const IndexSchema = z.object({
    "build_time": z.coerce.date(),
    "commit": z.string(),
    "module": z.string(),
    "updated_at": z.coerce.date(),
    "version": z.string(),
});
export type Index = z.infer<typeof IndexSchema>;
