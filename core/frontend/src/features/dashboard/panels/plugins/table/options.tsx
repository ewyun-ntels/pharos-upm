import React, {useState, useEffect} from 'react';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {extractColumnsFromQuery} from './setParam';
import {Accordion} from '@pharos/shared/components/ui-extension';
import {uuid} from '@pharos/shared/components/ui-extension';
import {Switch} from '@pharos/shared/components/ui';
import {Properties} from './properties';
import {MappingDialog} from './mappingDialog';
import {TablePanelOptions, FilterItem} from './types';
import {Button} from '@pharos/shared/components/ui';

interface FieldState {
  id: string;
  columnKey: string;
}

export const Options: React.FC<PanelEditorOptionsProps<TablePanelOptions>> = ({
  dataProvider,
  options = {},
  onOptionsChange,
}) => {
  const query = dataProvider?.chartQuery?.[0]?.query;
  const [fields, setFields] = useState<FieldState[]>([]);
  const [availableColumns, setAvailableColumns] = useState<string[]>([]);
  const [mappingDialogOpen, setMappingDialogOpen] = useState(false);
  const [currentFieldIndex, setCurrentFieldIndex] = useState<number | null>(null);

  // Extract columns from query
  useEffect(() => {
    if (query) {
      const columns = extractColumnsFromQuery(query);
      setAvailableColumns(columns);
    }
  }, [query]);

  // Initialize fields from existing columnConfigs with properties, or columns with dataLinks
  useEffect(() => {
    const columnConfigs = options.columnConfigs || [];
    const dataLinkKeys = Object.keys(options.columnDataLinks || {});

    const fieldKeys = new Set<string>([
      ...columnConfigs.filter((c) => c.properties).map((c) => c.key as string),
      ...dataLinkKeys,
    ]);

    const existingFields = Array.from(fieldKeys).map((key) => ({
      id: uuid(),
      columnKey: key,
    }));

    if (existingFields.length > 0) {
      setFields(existingFields);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // 최초 한번만 실행

  const handleAddField = () => {
    setFields([...fields, {id: uuid(), columnKey: ''}]);
  };

  const handleRemoveField = (id: string) => {
    const field = fields.find((f) => f.id === id);
    if (field && field.columnKey) {
      const newColumnConfigs = options.columnConfigs?.map((config) => {
        if (config.key === field.columnKey) {
          const {properties, ...rest} = config;
          return rest;
        }
        return config;
      });
      const newColumnDataLinks = {...(options.columnDataLinks || {})};
      delete newColumnDataLinks[field.columnKey];
      onOptionsChange?.({...options, columnConfigs: newColumnConfigs, columnDataLinks: newColumnDataLinks});
    }
    setFields(fields.filter((f) => f.id !== id));
  };

  const handleColumnSelect = (id: string, columnKey: string) => {
    setFields(fields.map((f) => (f.id === id ? {...f, columnKey} : f)));
  };

  const handleOpenMappingDialog = (fieldIndex: number) => {
    setCurrentFieldIndex(fieldIndex);
    setMappingDialogOpen(true);
  };

  const handleSaveBadgeData = (newBadgeData: any) => {
    if (currentFieldIndex === null) return;

    const field = fields[currentFieldIndex];
    const columnIndex = options.columnConfigs?.findIndex((c) => c.key === field.columnKey);

    if (columnIndex !== undefined && columnIndex >= 0) {
      const newColumnConfigs = [...(options.columnConfigs || [])];
      if (!newColumnConfigs[columnIndex].properties) {
        newColumnConfigs[columnIndex].properties = {} as any;
      }
      (newColumnConfigs[columnIndex].properties as any).badge = newBadgeData;
      onOptionsChange?.({...options, columnConfigs: newColumnConfigs});
    }
  };

  return (
    <div>
      <Accordion
        id={'table-options'}
        items={[
          {
            id: 'options',
            title: 'Table options',
            content: (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Use search</label>
                  <Switch
                    checked={
                      options?.leftItems?.some((item: FilterItem) => item.type === 'search') || false
                    }
                    onCheckedChange={(val) => {
                      const leftItems: FilterItem[] = val ? [{id: 'search', type: 'search'}] : [];
                      onOptionsChange?.({...options, leftItems});
                    }}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Use export</label>
                  <Switch
                    checked={
                      options?.rightItems?.some((item: FilterItem) => item.type === 'export') || false
                    }
                    onCheckedChange={(val) => {
                      const rightItems: FilterItem[] = val ? [{id: 'export', type: 'export'}] : [];
                      onOptionsChange?.({...options, rightItems});
                    }}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Use pagination</label>
                  <Switch
                    checked={options?.usePagenation || false}
                    onCheckedChange={(val) => {
                      onOptionsChange?.({...options, usePagenation: val});
                    }}
                  />
                </div>
              </div>
            ),
          },
          {
            id: 'fields',
            title: 'Fields',
            content: (
              <div className="space-y-2">
                {fields.map((field, index) => {
                  const columnIndex =
                    options.columnConfigs?.findIndex((c) => c.key === field.columnKey) ?? -1;
                  return (
                    <Accordion
                      key={field.id}
                      id={`accordion_${field.id}`}
                      items={[
                        {
                          id: `accordion_item_${field.id}`,
                          title: `Field ${index + 1}`,
                          content: (
                            <Properties
                              id={field.id}
                              availableColumns={availableColumns}
                              selectedColumn={field.columnKey}
                              columnIndex={columnIndex}
                              tableOptions={options}
                              onColumnSelect={(columnKey) =>
                                handleColumnSelect(field.id, columnKey)
                              }
                              onOptionsChange={onOptionsChange!}
                              onRemove={() => handleRemoveField(field.id)}
                              onOpenMappingDialog={() => handleOpenMappingDialog(index)}
                            />
                          ),
                        },
                      ]}
                      defaultOpenItems={'all'}
                      type="single"
                    />
                  );
                })}
                <Button
                  variant="outline"
                  size="sm"
                  className="h-8 mt-1 w-full"
                  onClick={handleAddField}
                  disabled={availableColumns.length === 0}
                >
                  Add Field
                </Button>
              </div>
            ),
          },
        ]}
        defaultOpenItems={'all'}
        type="multiple"
      />

      {currentFieldIndex !== null && (
        <MappingDialog
          open={mappingDialogOpen}
          onOpenChange={setMappingDialogOpen}
          badgeData={
            options.columnConfigs?.[
              options.columnConfigs.findIndex(
                (c) => c.key === fields[currentFieldIndex].columnKey,
              )
            ]?.properties?.badge as any
          }
          onSave={handleSaveBadgeData}
        />
      )}
    </div>
  );
};
