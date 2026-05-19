import React from 'react';
import {Badge} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {Input} from '@pharos/shared/components/ui';
import {Switch} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {tableUnitTypeList} from '@pharos/shared/lib/unitUtils';
import {TrashIcon} from '@pharos/shared/components';
import {TablePanelOptions, DataLink} from './types';

type BadgeMapping = {
  [key: string]: {
    color: string;
    textColor?: string;
    displayText?: string;
  };
};

export const Properties = ({
  availableColumns,
  selectedColumn,
  columnIndex,
  tableOptions,
  onColumnSelect,
  onOptionsChange,
  onRemove,
  onOpenMappingDialog,
}: {
  id: string;
  availableColumns: string[];
  selectedColumn: string;
  columnIndex: number;
  tableOptions: TablePanelOptions;
  onColumnSelect: (columnKey: string) => void;
  onOptionsChange: (newOptions: TablePanelOptions) => void;
  onRemove: () => void;
  onOpenMappingDialog: () => void;
}) => {
  const isValidIndex = columnIndex >= 0 && columnIndex < (tableOptions.columnConfigs?.length ?? 0);
  const columnConfig = isValidIndex ? tableOptions.columnConfigs?.[columnIndex] : undefined;
  const properties = columnConfig?.properties || {};
  const columnKey = selectedColumn;
  const dataLinks: DataLink[] = (tableOptions.columnDataLinks?.[columnKey] || []);

  // properties나 dataLinks가 변경될 때마다 Add property Select를 리셋
  const selectResetKey = `${Object.keys(properties).sort().join(',')}_${dataLinks.length}`;

  /** columnConfigs를 불변성 보장하며 복사 후 해당 컬럼 properties를 updater로 수정 */
  const updateColumnProperties = (updater: (props: Record<string, unknown>) => void) => {
    if (!isValidIndex) return;
    const newColumnConfigs = (tableOptions.columnConfigs || []).map((c, i) =>
      i === columnIndex ? {...c, properties: {...(c.properties || {})}} : c,
    );
    updater(newColumnConfigs[columnIndex].properties as Record<string, unknown>);
    onOptionsChange({...tableOptions, columnConfigs: newColumnConfigs});
  };

  const handlePropertyChange = (
    propertyKey: keyof typeof properties,
    value: string | number | boolean | BadgeMapping | unknown[] | object,
  ) => {
    updateColumnProperties((props) => {
      props[propertyKey] = value;
    });
  };

  const handleDeleteProperty = (propertyKey: keyof typeof properties) => {
    updateColumnProperties((props) => {
      delete props[propertyKey];
    });
  };

  const handleAddProperty = (propertyKey: string) => {
    if (propertyKey === 'dataLinks') {
      const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
      if (!newColumnDataLinks[columnKey]) {
        newColumnDataLinks[columnKey] = [{title: '', url: '', targetBlank: false}];
      }
      onOptionsChange({...tableOptions, columnDataLinks: newColumnDataLinks});
      return;
    }

    const defaults: Record<string, unknown> = {
      hide: false,
      title: '',
      unit: undefined,
      size: undefined,
      group: '',
      badge: {},
    };
    const moreDefaults: Record<string, unknown> = {
      colorRules: [],
      gauge: {min: 0, max: 100, color: '#3b82f6'},
    };
    if (propertyKey in defaults) {
      updateColumnProperties((props) => {
        props[propertyKey] = defaults[propertyKey];
      });
    } else if (propertyKey in moreDefaults) {
      updateColumnProperties((props) => {
        props[propertyKey] = moreDefaults[propertyKey];
      });
    }
  };

  const handleColumnChange = (newColumnKey: string) => {
    let newColumnConfigs = (tableOptions.columnConfigs || []).map((c) => ({...c}));
    let newIndex = newColumnConfigs.findIndex((col) => col.key === newColumnKey);

    if (newIndex >= 0) {
      newColumnConfigs[newIndex] = {
        ...newColumnConfigs[newIndex],
        properties: {...(columnConfig?.properties || {})},
      };
    } else {
      newColumnConfigs.push({key: newColumnKey as any});
      newIndex = newColumnConfigs.length - 1;
    }

    // 이전 컬럼의 properties 제거
    if (isValidIndex && columnIndex !== newIndex) {
      const {properties: _, ...rest} = newColumnConfigs[columnIndex];
      newColumnConfigs[columnIndex] = rest as any;
    }

    // dataLinks도 새 컬럼 키로 이동
    const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
    if (columnKey && columnKey !== newColumnKey && newColumnDataLinks[columnKey]) {
      newColumnDataLinks[newColumnKey] = newColumnDataLinks[columnKey];
      delete newColumnDataLinks[columnKey];
    }

    onOptionsChange({...tableOptions, columnConfigs: newColumnConfigs, columnDataLinks: newColumnDataLinks});
    onColumnSelect(newColumnKey);
  };

  const handleAddDataLink = () => {
    const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
    newColumnDataLinks[columnKey] = [...(newColumnDataLinks[columnKey] || []), {title: '', url: '', targetBlank: false}];
    onOptionsChange({...tableOptions, columnDataLinks: newColumnDataLinks});
  };

  const handleDataLinkChange = (index: number, field: keyof DataLink, value: string | boolean) => {
    const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
    const links = [...(newColumnDataLinks[columnKey] || [])];
    links[index] = {...links[index], [field]: value};
    newColumnDataLinks[columnKey] = links;
    onOptionsChange({...tableOptions, columnDataLinks: newColumnDataLinks});
  };

  const handleDeleteDataLink = (index: number) => {
    const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
    const links = [...(newColumnDataLinks[columnKey] || [])];
    links.splice(index, 1);
    newColumnDataLinks[columnKey] = links;
    onOptionsChange({...tableOptions, columnDataLinks: newColumnDataLinks});
  };

  const handleDeleteAllDataLinks = () => {
    const newColumnDataLinks = {...(tableOptions.columnDataLinks || {})};
    delete newColumnDataLinks[columnKey];
    onOptionsChange({...tableOptions, columnDataLinks: newColumnDataLinks});
  };

  const renderPropertyItem = (key: string) => {
    if (key === 'hide') return (
      <div key="hide" className="flex items-center justify-between">
        <label className="text-sm text-muted-foreground">Hide in table</label>
        <div className="flex items-center gap-2">
          <Switch
            checked={properties.hide || false}
            onCheckedChange={(val) => handlePropertyChange('hide', val)}
          />
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('hide')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
      </div>
    );
    if (key === 'title') return (
      <div key="title" className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Display name</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('title')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <Input value={properties.title || ''} onChange={(e) => handlePropertyChange('title', e.target.value)} />
      </div>
    );
    if (key === 'badge') return (
      <div key="badge" className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Value mappings</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('badge')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <p className="text-[11px] text-muted-foreground">
          Displays a colored badge based on the cell value. Click the button below to configure condition → color mappings.
        </p>
        {Object.entries(properties.badge || {}).map(([k, value]: [string, any]) => (
          <div key={k} className="flex items-center gap-2">
            <span className="text-sm font-medium flex-1 text-center">{k}</span>
            <span className="text-sm font-medium flex-1 text-center">{'->'}</span>
            <span className="text-sm font-medium flex-1 text-center">
              <Badge
                variant="outline"
                className={`${value.textColor ? 'border' : 'border-0'} text-white shadow-none`}
                style={{backgroundColor: value.color, color: value.textColor || '#FFFFFF', borderColor: value.textColor}}
              >
                {value.displayText || k}
              </Badge>
            </span>
          </div>
        ))}
        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={onOpenMappingDialog}>
          {Object.keys(properties.badge || {}).length === 0 ? 'Add value mappings' : 'Edit value mappings'}
        </Button>
      </div>
    );
    if (key === 'unit') return (
      <div key="unit" className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Unit</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('unit')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <Select value={properties.unit || ''} onValueChange={(val) => handlePropertyChange('unit', val)}>
          <SelectTrigger><SelectValue placeholder="Choose unit" /></SelectTrigger>
          <SelectContent>
            {tableUnitTypeList.map((unit) => <SelectItem key={unit} value={unit}>{unit}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
    );
    if (key === 'size') return (
      <div key="size" className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Column width</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('size')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <Input type="number" value={properties.size ?? ''} onChange={(e) => handlePropertyChange('size', Number(e.target.value))} />
      </div>
    );
    if (key === 'group') return (
      <div key="group" className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Group name</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('group')}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <Input value={properties.group || ''} onChange={(e) => handlePropertyChange('group', e.target.value)} />
      </div>
    );
    if (key === 'colorRules') {
      const rules = (properties.colorRules as any[]) || [];
      const conditionLabels: Record<string, string> = {gt: '>', gte: '>=', lt: '<', lte: '<=', eq: '='};
      const handleAddRule = () => {
        handlePropertyChange('colorRules', [...rules, {condition: 'gte', value: 0, color: '#ef4444', textColor: '#ffffff'}]);
      };
      const handleUpdateRule = (i: number, field: string, val: any) => {
        const updated = rules.map((r, idx) => idx === i ? {...r, [field]: val} : r);
        handlePropertyChange('colorRules', updated);
      };
      const handleDeleteRule = (i: number) => {
        handlePropertyChange('colorRules', rules.filter((_, idx) => idx !== i));
      };
      return (
        <div key="colorRules" className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <label className="text-sm text-muted-foreground">Color rules</label>
            <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('colorRules')}>
              <TrashIcon className="w-4 h-4" />
            </Button>
          </div>
          <p className="text-[11px] text-muted-foreground">Applies background color when the numeric value matches a condition. Rules are checked top to bottom; the first match wins.</p>
          {rules.map((rule: any, i: number) => (
            <div key={i} className="flex items-center gap-1 flex-wrap">
              <Select value={rule.condition} onValueChange={(v) => handleUpdateRule(i, 'condition', v)}>
                <SelectTrigger className="w-16 h-7 text-xs"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(conditionLabels).map(([k, label]) => <SelectItem key={k} value={k}>{label}</SelectItem>)}
                </SelectContent>
              </Select>
              <Input type="number" value={rule.value} onChange={(e) => handleUpdateRule(i, 'value', Number(e.target.value))} className="w-20 h-7 text-xs" />
              <Input type="color" value={rule.color || '#ef4444'} onChange={(e) => handleUpdateRule(i, 'color', e.target.value)} className="w-8 h-7 p-0 cursor-pointer" title="Background color" />
              <Input type="color" value={rule.textColor || '#ffffff'} onChange={(e) => handleUpdateRule(i, 'textColor', e.target.value)} className="w-8 h-7 p-0 cursor-pointer" title="Text color" />
              <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteRule(i)}>
                <TrashIcon className="w-3 h-3" />
              </Button>
            </div>
          ))}
          <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleAddRule}>+ Add rule</Button>
        </div>
      );
    }
    if (key === 'gauge') {
      const gauge = (properties.gauge as any) || {min: 0, max: 100, color: '#3b82f6'};
      const handleGaugeChange = (field: string, val: any) => {
        handlePropertyChange('gauge', {...gauge, [field]: val});
      };
      return (
        <div key="gauge" className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <label className="text-sm text-muted-foreground">Gauge</label>
            <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteProperty('gauge')}>
              <TrashIcon className="w-4 h-4" />
            </Button>
          </div>
          <p className="text-[11px] text-muted-foreground">Displays a bar gauge inside the cell. Set the min/max range and bar color.</p>
          <div className="flex items-center gap-2">
            <label className="text-xs text-muted-foreground w-8">Min</label>
            <Input type="number" value={gauge.min ?? 0} onChange={(e) => handleGaugeChange('min', Number(e.target.value))} className="h-7 text-xs" />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-xs text-muted-foreground w-8">Max</label>
            <Input type="number" value={gauge.max ?? 100} onChange={(e) => handleGaugeChange('max', Number(e.target.value))} className="h-7 text-xs" />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-xs text-muted-foreground w-8">Color</label>
            <Input type="color" value={gauge.color || '#3b82f6'} onChange={(e) => handleGaugeChange('color', e.target.value)} className="w-10 h-7 p-0 cursor-pointer" />
          </div>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="w-full space-y-3">
      {/* Field selector */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <label className="text-sm text-muted-foreground">Field</label>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={onRemove}>
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
        <Select value={selectedColumn} onValueChange={handleColumnChange}>
          <SelectTrigger>
            <SelectValue placeholder="Choose" />
          </SelectTrigger>
          <SelectContent>
            {availableColumns.map((col) => (
              <SelectItem key={col} value={col}>{col}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Data links — shown before other properties so newly added properties appear below */}
      {dataLinks.length > 0 && (
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <label className="text-sm text-muted-foreground">Data links</label>
            <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={handleDeleteAllDataLinks}>
              <TrashIcon className="w-4 h-4" />
            </Button>
          </div>
          <div className="rounded bg-muted px-3 py-2 mb-1 text-[11px] text-muted-foreground space-y-1">
            <div className="font-medium text-xs mb-1">URL variable examples</div>
            <code className="block">{'{{value}}'} — cell value</code>
            <code className="block">{'{{row.hostname}}'} — hostname column in the same row</code>
            <code className="block">{'{{env}}'} — dashboard filter variable (empty if not set)</code>
            <div className="pt-1 font-medium text-xs">URL examples (paste directly from browser address bar)</div>
            <code className="block">{'/ui/dashboards/uuid?id={{value}}'}</code>
            <code className="block">{'/ui/dashboards/uuid?host={{row.hostname}}&dc={{dc}}'}</code>
            <code className="block">{'https://other.example.com/page?id={{value}}'}</code>
          </div>
          {dataLinks.map((link, i) => (
            <div key={i} className="border border-border rounded p-2 space-y-2">
              <div className="flex items-center gap-1">
                <Input
                  placeholder="Title"
                  value={link.title}
                  onChange={(e) => handleDataLinkChange(i, 'title', e.target.value)}
                  className="flex-1 h-7 text-xs"
                />
                <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => handleDeleteDataLink(i)}>
                  <TrashIcon className="w-4 h-4" />
                </Button>
              </div>
              <Input
                placeholder="URL (e.g. /ui/dashboards/detail?host={{row.hostname}})"
                value={link.url}
                onChange={(e) => handleDataLinkChange(i, 'url', e.target.value)}
                className="h-7 text-xs"
              />
              <div className="flex items-center justify-between">
                <small className="text-[11px] text-muted-foreground">Open in new tab</small>
                <Switch
                  checked={link.targetBlank ?? false}
                  onCheckedChange={(val) => handleDataLinkChange(i, 'targetBlank', val)}
                />
              </div>
            </div>
          ))}
          <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleAddDataLink}>
            + Add link
          </Button>
        </div>
      )}

      {/* Properties rendered in insertion order — always added below data links */}
      {Object.keys(properties).map((key) => renderPropertyItem(key))}

      {/* Add property selector — always at the bottom */}
      {selectedColumn && isValidIndex && (
        <div className="flex flex-col gap-2 border-t border-border pt-3">
          <label className="text-sm text-muted-foreground">Add property</label>
          <Select key={selectResetKey} onValueChange={handleAddProperty}>
            <SelectTrigger>
              <SelectValue placeholder="Choose additional" />
            </SelectTrigger>
            <SelectContent>
              {properties.hide === undefined && <SelectItem value="hide">Hide in table</SelectItem>}
              {properties.title === undefined && <SelectItem value="title">Display name</SelectItem>}
              {properties.badge === undefined && <SelectItem value="badge">Value mappings</SelectItem>}
              {properties.unit === undefined && <SelectItem value="unit">Unit</SelectItem>}
              {properties.size === undefined && <SelectItem value="size">Column width</SelectItem>}
              {properties.group === undefined && <SelectItem value="group">Group</SelectItem>}
              {properties.colorRules === undefined && <SelectItem value="colorRules">Color rules</SelectItem>}
              {properties.gauge === undefined && <SelectItem value="gauge">Gauge</SelectItem>}
              {dataLinks.length === 0 && <SelectItem value="dataLinks">Data links</SelectItem>}
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  );
};
