import { Accordion } from '@pharos/shared/components/ui-extension';
import { Button } from '@pharos/shared/components/ui';
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';
import { MappingDialog } from './mappingDialog';
import { useState } from 'react';
import { PiePanelOptions, ColorMapping } from './types';
import { Input } from '@pharos/shared/components/ui';
import { Switch } from '@pharos/shared/components/ui';
import { Slider } from '@pharos/shared/components/ui';
import { Tabs, TabsList, TabsTrigger } from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import { unitTypeList } from '@pharos/shared/lib/unitUtils';
import { useOptionsChange } from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<PiePanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const [mappingDialogOpen, setMappingDialogOpen] = useState(false);
  const handleChange = useOptionsChange(options, onOptionsChange);

  const handleColorMappingsSave = (newMappings: ColorMapping[]) => {
    handleChange(draft => {
      if (!draft.aliasColors) draft.aliasColors = { colors: [] };
      draft.aliasColors.colors = newMappings;
    });
  };

  return (
    <div>
      <Accordion
        id={info?.id || 'pie'}
        items={[
          {
            id: 'LegendOption',
            title: 'Legend',
            content: (
              <div className="space-y-4">
                {/* Legend rule */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Legend rule</label>
                  <Input
                    value={options.legendRule || ''}
                    onChange={(e) => handleChange(d => { d.legendRule = e.target.value; })}
                    placeholder="Enter legend rule"
                  />
                  <p className="text-xs text-muted-foreground">Template using metric label placeholders. e.g. <code>{'{instance}'} - {'{job}'}</code></p>
                </div>

                {/* Visibility */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Visibility</label>
                  <Switch
                    checked={options.legendEnabled || false}
                    onCheckedChange={(checked) => handleChange(d => { d.legendEnabled = checked; })}
                  />
                </div>

                {options.legendEnabled && (
                  <>
                    {/* Placement */}
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Placement</label>
                      <Tabs
                        value={options.legendAlign || 'right'}
                        onValueChange={(value) => handleChange(d => { d.legendAlign = value as 'right' | 'bottom'; })}
                      >
                        <TabsList className="grid w-full grid-cols-2">
                          <TabsTrigger value="right">Right</TabsTrigger>
                          <TabsTrigger value="bottom">Bottom</TabsTrigger>
                        </TabsList>
                      </Tabs>
                    </div>

                    {/* Mode */}
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Mode</label>
                      <Tabs
                        value={options.legendAsTable ? 'true' : 'false'}
                        onValueChange={(value) => handleChange(d => { d.legendAsTable = value === 'true'; })}
                      >
                        <TabsList className="grid w-full grid-cols-2">
                          <TabsTrigger value="false">List</TabsTrigger>
                          <TabsTrigger value="true">Table</TabsTrigger>
                        </TabsList>
                      </Tabs>
                    </div>

                    {/* Show average */}
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show average</label>
                      <Switch
                        checked={options.showAverage || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showAverage = checked; })}
                      />
                    </div>

                    {/* Show last */}
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show last</label>
                      <Switch
                        checked={options.showLast || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showLast = checked; })}
                      />
                    </div>

                    {/* Show max */}
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show max</label>
                      <Switch
                        checked={options.showMax || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showMax = checked; })}
                      />
                    </div>

                    {/* Show min */}
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show min</label>
                      <Switch
                        checked={options.showMin || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showMin = checked; })}
                      />
                    </div>

                    {/* Show total */}
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show total</label>
                      <Switch
                        checked={options.showTotal || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showTotal = checked; })}
                      />
                    </div>
                  </>
                )}
              </div>
            ),
          },
          {
            id: 'options',
            title: 'Pie chart options',
            content: (
              <div className="space-y-4">
                {/* Chart type */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Chart type</label>
                  <Tabs
                    value={options.chartType || 'donut'}
                    onValueChange={(value) => handleChange(d => { d.chartType = value as 'pie' | 'donut'; })}
                  >
                    <TabsList className="grid w-full grid-cols-2">
                      <TabsTrigger value="pie">Pie</TabsTrigger>
                      <TabsTrigger value="donut">Donut</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>

                {/* Slice labels */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Slice labels</label>
                  <Tabs
                    value={options.showLabel || 'none'}
                    onValueChange={(value) => handleChange(d => { d.showLabel = value as 'none' | 'value' | 'percent'; })}
                  >
                    <TabsList className="grid w-full grid-cols-3">
                      <TabsTrigger value="none">None</TabsTrigger>
                      <TabsTrigger value="value">Value</TabsTrigger>
                      <TabsTrigger value="percent">Percent</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>

                {/* Inner radius (donut only) */}
                {(options.chartType === 'donut' || !options.chartType) && (
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">
                      Inner radius: {options.innerRadius ?? 40}
                    </label>
                    <Slider
                      value={[options.innerRadius ?? 40]}
                      onValueChange={([value]) => handleChange(d => { d.innerRadius = value; })}
                      min={0}
                      max={90}
                      step={1}
                    />
                  </div>
                )}

                {/* Outer radius */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">
                    Outer radius: {options.outerRadius ?? 80}
                  </label>
                  <Slider
                    value={[options.outerRadius ?? 80]}
                    onValueChange={([value]) => handleChange(d => { d.outerRadius = value; })}
                    min={10}
                    max={100}
                    step={1}
                  />
                </div>

                {/* Stroke width */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">
                    Stroke width: {options.strokeWidth ?? 2}
                  </label>
                  <Slider
                    value={[options.strokeWidth ?? 2]}
                    onValueChange={([value]) => handleChange(d => { d.strokeWidth = value; })}
                    min={0}
                    max={10}
                    step={1}
                  />
                </div>

                {/* Unit */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Unit</label>
                  <Select
                    value={options.unit || 'short'}
                    onValueChange={(value) => handleChange(d => { d.unit = value; })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select unit" />
                    </SelectTrigger>
                    <SelectContent>
                      {unitTypeList.map((unit) => (
                        <SelectItem key={unit} value={unit}>
                          {unit}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Decimals */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Decimals</label>
                  <Input
                    type="number"
                    min={0}
                    max={10}
                    value={options.decimals ?? 2}
                    onChange={(e) => handleChange(d => { d.decimals = Math.max(0, parseInt(e.target.value) || 0); })}
                  />
                </div>

                {/* Color mappings */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Color mappings</label>
                  {options.aliasColors?.colors?.map(
                    (item: { legendName: string; color: string }, index: number) => (
                      <div key={index} className="flex items-center gap-2">
                        <span className="text-sm font-medium flex-1 text-center">
                          {item.legendName}
                        </span>
                        <span className="text-sm font-medium flex-1 text-center">{'->'}</span>
                        <span className="text-sm font-medium flex-1 text-center">
                          <span
                            className="w-5 h-5 rounded border border-border inline-block"
                            style={{ backgroundColor: item.color }}
                          />
                        </span>
                      </div>
                    ),
                  )}
                  <Button
                    className="border border-border rounded px-4 py-0.5 bg-gray-400 text-white hover:bg-gray-500 h-auto!"
                    onClick={() => setMappingDialogOpen(true)}
                  >
                    Add color mappings
                  </Button>
                </div>
              </div>
            ),
          },
        ]}
        defaultOpenItems={'all'}
        type="multiple"
      />
      <MappingDialog
        open={mappingDialogOpen}
        onOpenChange={setMappingDialogOpen}
        colorMappings={options.aliasColors?.colors || []}
        onSave={handleColorMappingsSave}
      />
    </div>
  );
};
