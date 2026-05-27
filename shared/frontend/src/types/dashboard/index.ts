import * as z from "zod";

// global: stored in DashboardConfig.annotations, shown on all timeseries panels. panel:
// stored in panel.options.annotations, shown only on that panel.

export const ScopeSchema = z.enum([
    "global",
    "panel",
]);
export type Scope = z.infer<typeof ScopeSchema>;

// Panel/filter kind used for backend query routing. Determines where a panel or filter is
// located: headerL (top-left), headerR (top-right), left (sidebar), panels (chart panel
// grid).

export const KindSchema = z.enum([
    "headerL",
    "headerR",
    "left",
    "panels",
]);
export type Kind = z.infer<typeof KindSchema>;

// Filter type: select, datetime, step, refreshInterval, custom, tableSearch

export const FilterConfigTypeSchema = z.enum([
    "custom",
    "datetime",
    "refreshInterval",
    "select",
    "step",
    "tableSearch",
]);
export type FilterConfigType = z.infer<typeof FilterConfigTypeSchema>;

// Layout type: card (normal panel) or row (container for nested panels)

export const PanelLayoutTypeSchema = z.enum([
    "card",
    "row",
]);
export type PanelLayoutType = z.infer<typeof PanelLayoutTypeSchema>;

// User permission level for this dashboard

export const ActionSchema = z.enum([
    "editor",
    "owner",
    "viewer",
]);
export type Action = z.infer<typeof ActionSchema>;

export const DashboardCreateResponseSchema = z.object({
    "id": z.string(),
});
export type DashboardCreateResponse = z.infer<typeof DashboardCreateResponseSchema>;

export const AnnotationSchema = z.object({
    "color": z.string().optional(),
    "id": z.string(),
    "panel_id": z.string().optional(),
    "scope": ScopeSchema,
    "text": z.string(),
    "time": z.number(),
});
export type Annotation = z.infer<typeof AnnotationSchema>;

export const FilterConfigSchema = z.object({
    "datasourceName": z.string().optional(),
    "id": z.string(),
    "kind": KindSchema,
    "label": z.string().optional(),
    "options": z.record(z.string(), z.any()).optional(),
    "query": z.string().optional(),
    "type": FilterConfigTypeSchema,
});
export type FilterConfig = z.infer<typeof FilterConfigSchema>;

export const ChartQuerySchema = z.object({
    "datasourceName": z.string(),
    "forceRefresh": z.number().optional(),
    "label": z.string(),
    "query": z.string(),
    "templateStyle": z.string().optional(),
});
export type ChartQuery = z.infer<typeof ChartQuerySchema>;

export const PanelLayoutSchema = z.object({
    "collapsed": z.boolean().optional(),
    "h": z.number(),
    "isDraggable": z.boolean().optional(),
    "isResizable": z.boolean().optional(),
    "maxH": z.number().optional(),
    "maxW": z.number().optional(),
    "minH": z.number().optional(),
    "minW": z.number().optional(),
    "static": z.boolean().optional(),
    "type": PanelLayoutTypeSchema,
    "w": z.number(),
    "x": z.number(),
    "y": z.number(),
});
export type PanelLayout = z.infer<typeof PanelLayoutSchema>;

export const DashboardFolderSchema = z.object({
    "createdAt": z.coerce.date(),
    "id": z.string(),
    "name": z.string(),
    "updatedAt": z.coerce.date(),
});
export type DashboardFolder = z.infer<typeof DashboardFolderSchema>;

export const DashboardPermissionSchema = z.object({
    "action": ActionSchema,
    "kind": z.string(),
    "object": z.string(),
    "subject": z.string(),
});
export type DashboardPermission = z.infer<typeof DashboardPermissionSchema>;

export const InspectInfoSchema = z.object({
    "datasourceName": z.string(),
    "error": z.string().optional(),
    "executedQuery": z.string(),
    "rawQuery": z.string(),
    "timestamp": z.coerce.date(),
    "variables": z.record(z.string(), z.any()),
});
export type InspectInfo = z.infer<typeof InspectInfoSchema>;

export const QueryInspectResponseStatisticsSchema = z.object({
    "elapsed": z.number(),
});
export type QueryInspectResponseStatistics = z.infer<typeof QueryInspectResponseStatisticsSchema>;

export const QueryResponseStatisticsSchema = z.object({
    "elapsed": z.number(),
});
export type QueryResponseStatistics = z.infer<typeof QueryResponseStatisticsSchema>;

export const DataProviderSchema = z.object({
    "chartQuery": z.array(ChartQuerySchema).optional(),
    "dataProviderName": z.string().optional(),
    "resource": z.string().optional(),
});
export type DataProvider = z.infer<typeof DataProviderSchema>;

export const QueryInspectResponseSchema = z.object({
    "data": z.array(z.record(z.string(), z.any())),
    "inspect": InspectInfoSchema,
    "meta": z.array(z.record(z.string(), z.string())),
    "rows": z.number(),
    "statistics": QueryInspectResponseStatisticsSchema,
});
export type QueryInspectResponse = z.infer<typeof QueryInspectResponseSchema>;

export const QueryResponseSchema = z.object({
    "data": z.array(z.record(z.string(), z.any())),
    "meta": z.array(z.record(z.string(), z.string())),
    "rows": z.number(),
    "statistics": QueryResponseStatisticsSchema,
});
export type QueryResponse = z.infer<typeof QueryResponseSchema>;

export const PanelSchema = z.object({
    "bgTransparent": z.boolean().optional(),
    "dataProvider": DataProviderSchema.optional(),
    "description": z.string().optional(),
    "id": z.string(),
    "layout": PanelLayoutSchema,
    "options": z.record(z.string(), z.any()).optional(),
    "renderType": z.string().optional(),
    "title": z.string(),
});
export type Panel = z.infer<typeof PanelSchema>;

export const DashboardConfigSchema = z.object({
    "annotations": z.array(AnnotationSchema).optional(),
    "description": z.string(),
    "displayName": z.string(),
    "favorite": z.boolean(),
    "filters": z.array(FilterConfigSchema),
    "groupName": z.string().optional(),
    "panels": z.array(PanelSchema).optional(),
    "title": z.string(),
    "type": z.string(),
});
export type DashboardConfig = z.infer<typeof DashboardConfigSchema>;

export const DashboardDataSchema = z.object({
    "annotations": z.array(AnnotationSchema).optional(),
    "config": DashboardConfigSchema,
    "folderId": z.string().optional(),
    "id": z.string(),
    "permission": ActionSchema,
});
export type DashboardData = z.infer<typeof DashboardDataSchema>;
