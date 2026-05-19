import { z } from 'zod';

/**
 * MarkdownViewer Panel Options Schema
 * Defines the configuration options for MarkdownViewer panel type
 */
export const MarkdownViewerPanelOptionsSchema = z.object({
  // Markdown content
  content: z.string().optional(),
  
  // Custom message when data doesn't exist
  chartDataNotExistMessage: z.string().optional(),
});

/**
 * MarkdownViewer Panel Options Type
 * Inferred from the Zod schema
 */
export type MarkdownViewerPanelOptions = z.infer<typeof MarkdownViewerPanelOptionsSchema>;
