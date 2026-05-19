import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {useOne, useUpdate, useInvalidate} from '@/lib/data-provider';
import validator from '@rjsf/validator-ajv8';
import {ScrollArea} from '@pharos/shared/components/ui';
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@pharos/shared/components/ui';
import {USER_PROVIDER_NAME, USER_RESOURCES} from '@providers/user-provider';
import {useToast} from '@hooks/use-toast';
import {JsonEditor} from '@components/editor';
import Form from '@components/rjsf/Form/Form';
import {SchemaEditorActions} from './SchemaEditorActions';

const DEFAULT_SCHEMA = {
  type: 'object',
  properties: {},
};
const DEFAULT_UI_SCHEMA = {};
const DEFAULT_SCHEMA_STR = JSON.stringify(DEFAULT_SCHEMA, null, 2);
const DEFAULT_UI_SCHEMA_STR = JSON.stringify(DEFAULT_UI_SCHEMA, null, 2);

export function SchemaEditorTab() {
  const {toast} = useToast();
  const invalidate = useInvalidate();

  const {
    query: {data: configData, isLoading},
  } = useOne({
    resource: USER_RESOURCES.USER_METADATA_CONFIG,
    id: 'active',
    dataProviderName: USER_PROVIDER_NAME,
    queryOptions: {staleTime: 0},
  });

  // Server-loaded strings (for dirty tracking)
  const [serverSchemaStr, setServerSchemaStr] = useState(DEFAULT_SCHEMA_STR);
  const [serverUiSchemaStr, setServerUiSchemaStr] = useState(DEFAULT_UI_SCHEMA_STR);

  // Editor strings (live edit state)
  const [schemaStr, setSchemaStr] = useState(DEFAULT_SCHEMA_STR);
  const [uiSchemaStr, setUiSchemaStr] = useState(DEFAULT_UI_SCHEMA_STR);

  // Parsed objects for RJSF preview
  const [parsedSchema, setParsedSchema] = useState<Record<string, any>>(DEFAULT_SCHEMA);
  const [parsedUiSchema, setParsedUiSchema] = useState<Record<string, any>>(DEFAULT_UI_SCHEMA);

  // FormData from RJSF onChange (for FormData output panel)
  const [formData, setFormData] = useState<any>(undefined);
  const [formDataStr, setFormDataStr] = useState('{}');

  const [isSaving, setIsSaving] = useState(false);

  const {mutateAsync: updateConfig} = useUpdate();

  // Sync editor state when server data loads
  useEffect(() => {
    const raw = configData?.data?.data ?? configData?.data;
    if (!raw) return;

    const schema = raw.schema ?? DEFAULT_SCHEMA;
    const uiSchema = raw.uiSchema ?? DEFAULT_UI_SCHEMA;
    const sStr = JSON.stringify(schema, null, 2);
    const uStr = JSON.stringify(uiSchema, null, 2);

    setServerSchemaStr(sStr);
    setServerUiSchemaStr(uStr);
    setSchemaStr(sStr);
    setUiSchemaStr(uStr);
    setParsedSchema(schema);
    setParsedUiSchema(uiSchema);
    setFormData(undefined);
    setFormDataStr('{}');
  }, [configData]);

  const isDirty = schemaStr !== serverSchemaStr || uiSchemaStr !== serverUiSchemaStr;

  const handleSchemaChange = useCallback((parsed: Record<string, any>) => {
    setParsedSchema(parsed);
    setFormData(undefined);
    setFormDataStr('{}');
  }, []);

  const handleUiSchemaChange = useCallback((parsed: Record<string, any>) => {
    setParsedUiSchema(parsed);
  }, []);

  const handleReset = useCallback(() => {
    setSchemaStr(serverSchemaStr);
    setUiSchemaStr(serverUiSchemaStr);
    try {
      setParsedSchema(JSON.parse(serverSchemaStr));
    } catch {}
    try {
      setParsedUiSchema(JSON.parse(serverUiSchemaStr));
    } catch {}
    setFormData(undefined);
    setFormDataStr('{}');
  }, [serverSchemaStr, serverUiSchemaStr]);

  const handleSave = useCallback(
    async (description?: string) => {
      let schema: Record<string, any>;
      let uiSchema: Record<string, any>;
      try {
        schema = JSON.parse(schemaStr);
        uiSchema = JSON.parse(uiSchemaStr);
      } catch {
        toast({description: 'Invalid JSON. Please fix errors before saving.', variant: 'destructive'});
        return;
      }

      setIsSaving(true);
      try {
        await updateConfig({
          resource: USER_RESOURCES.USER_METADATA_CONFIG,
          id: 'active',
          values: {config: {schema, uiSchema}, description},
          dataProviderName: USER_PROVIDER_NAME,
        });
        const sStr = JSON.stringify(schema, null, 2);
        const uStr = JSON.stringify(uiSchema, null, 2);
        setServerSchemaStr(sStr);
        setServerUiSchemaStr(uStr);
        invalidate({
          resource: USER_RESOURCES.USER_METADATA_CONFIG,
          dataProviderName: USER_PROVIDER_NAME,
          invalidates: ['detail'],
        });
        invalidate({
          resource: USER_RESOURCES.USER_METADATA_CONFIG_HISTORY,
          dataProviderName: USER_PROVIDER_NAME,
          invalidates: ['list'],
        });
        toast({description: 'Configuration saved successfully.'});
      } catch {
        toast({description: 'Failed to save configuration.', variant: 'destructive'});
      } finally {
        setIsSaving(false);
      }
    },
    [schemaStr, uiSchemaStr, updateConfig, invalidate, toast],
  );

  const previewUiSchema = useMemo(
    () => ({
      ...parsedUiSchema,
      'ui:submitButtonOptions': {norender: true},
    }),
    [parsedUiSchema],
  );

  if (isLoading) {
    return (
      <LoadingIndicator className="h-full" />
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-end gap-2 px-4 pb-2 border-b shrink-0">
        <SchemaEditorActions
          isDirty={isDirty}
          isSaving={isSaving}
          onSave={handleSave}
          onReset={handleReset}
        />
      </div>

      <ResizablePanelGroup orientation="horizontal" className="flex-1 min-h-0">
        {/* Left: Editors */}
        <ResizablePanel defaultSize={60} minSize={30}>
          <ResizablePanelGroup orientation="vertical">
            {/* JSON Schema editor */}
            <ResizablePanel defaultSize={45} minSize={15}>
              <div className="flex flex-col h-full">
                <p className="text-xs text-muted-foreground px-3 py-1.5 font-medium uppercase tracking-wide border-b shrink-0">
                  JSON Schema
                </p>
                <div className="flex-1 min-h-0">
                  <JsonEditor
                    setValue={schemaStr}
                    onChangeValue={setSchemaStr}
                    onChange={handleSchemaChange}
                    height="100%"
                  />
                </div>
              </div>
            </ResizablePanel>

            <ResizableHandle withHandle />

            {/* UISchema editor */}
            <ResizablePanel defaultSize={30} minSize={15}>
              <div className="flex flex-col h-full">
                <p className="text-xs text-muted-foreground px-3 py-1.5 font-medium uppercase tracking-wide border-b shrink-0">
                  UI Schema
                </p>
                <div className="flex-1 min-h-0">
                  <JsonEditor
                    setValue={uiSchemaStr}
                    onChangeValue={setUiSchemaStr}
                    onChange={handleUiSchemaChange}
                    height="100%"
                  />
                </div>
              </div>
            </ResizablePanel>

            <ResizableHandle withHandle />

            {/* FormData output */}
            <ResizablePanel defaultSize={25} minSize={10}>
              <div className="flex flex-col h-full">
                <p className="text-xs text-muted-foreground px-3 py-1.5 font-medium uppercase tracking-wide border-b shrink-0">
                  Form Data
                </p>
                <div className="flex-1 min-h-0">
                  <JsonEditor setValue={formDataStr} height="100%" />
                </div>
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>

        <ResizableHandle withHandle />

        {/* Right: RJSF live preview */}
        <ResizablePanel defaultSize={40} minSize={20}>
          <ScrollArea className="h-full">
            <div className="p-4">
              <p className="text-xs text-muted-foreground mb-3 font-medium uppercase tracking-wide">
                Preview
              </p>
              <Form
                schema={parsedSchema}
                uiSchema={previewUiSchema}
                formData={formData}
                validator={validator}
                noHtml5Validate
                onChange={(e) => {
                  setFormData(e.formData);
                  try {
                    setFormDataStr(JSON.stringify(e.formData ?? {}, null, 2));
                  } catch {}
                }}
              />
            </div>
          </ScrollArea>
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
