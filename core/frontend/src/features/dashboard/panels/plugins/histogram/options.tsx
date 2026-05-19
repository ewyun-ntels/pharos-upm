import { Accordion } from '@pharos/shared/components/ui-extension';
import { Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@pharos/shared/components/ui';
import { Switch } from '@pharos/shared/components/ui';
import { Slider } from '@pharos/shared/components/ui';
import { Tabs, TabsList, TabsTrigger } from '@pharos/shared/components/ui';
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';
import { HistogramPanelOptions } from './types';
import React from 'react';
import { useOptionsChange } from '../common/useOptionsChange';
import { unitTypeList } from '@pharos/shared/lib/unitUtils';

export const Options: React.FC<PanelEditorOptionsProps<HistogramPanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'histogram'}
        defaultOpenItems={'all'}
        type="multiple"
        items={[
          {
            id: 'LegendOption',
            title: 'Legend',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Legend rule</label>
                  <Input
                    value={options.legendRule || ''}
                    onChange={(e) => handleChange(d => { d.legendRule = e.target.value; })}
                    placeholder="Enter legend rule"
                  />
                  <p className="text-xs text-muted-foreground">Template using metric label placeholders. e.g. <code>{'{instance}'} - {'{job}'}</code></p>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Visibility</label>
                  <Switch
                    checked={options.legendEnabled !== false}
                    onCheckedChange={(checked) => handleChange(d => { d.legendEnabled = checked; })}
                  />
                </div>

                {options.legendEnabled !== false && (
                  <>
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Placement</label>
                      <Tabs
                        value={options.legendOption || 'bottom'}
                        onValueChange={(value) => {
                          if (value === 'bottom' || value === 'right') {
                            handleChange(d => { d.legendOption = value; });
                          }
                        }}
                      >
                        <TabsList className="grid w-full grid-cols-2">
                          <TabsTrigger value="bottom">Bottom</TabsTrigger>
                          <TabsTrigger value="right">Right</TabsTrigger>
                        </TabsList>
                      </Tabs>
                    </div>

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

                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show average</label>
                      <Switch
                        checked={options.showAverage || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showAverage = checked; })}
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show last</label>
                      <Switch
                        checked={options.showLast || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showLast = checked; })}
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show max</label>
                      <Switch
                        checked={options.showMax || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showMax = checked; })}
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show min</label>
                      <Switch
                        checked={options.showMin || false}
                        onCheckedChange={(checked) => handleChange(d => { d.showMin = checked; })}
                      />
                    </div>

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
            id: 'ChartOptions',
            title: 'Chart Options',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Bucket count</label>
                  <Input
                    type="number"
                    value={options.bucketCount ?? 10}
                    onChange={(e) => handleChange(d => { d.bucketCount = parseInt(e.target.value) || 10; })}
                    placeholder="Number of bins (default: 10)"
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Fill opacity: {Math.round((options.fillOpacity ?? 0.8) * 100)}%</label>
                  <Slider
                    value={[Math.round((options.fillOpacity ?? 0.8) * 100)]}
                    onValueChange={([value]) => handleChange(d => { d.fillOpacity = value / 100; })}
                    min={0}
                    max={100}
                    step={5}
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Tick font size</label>
                  <Input
                    type="number"
                    value={options.tickFont ?? 12}
                    onChange={(e) => handleChange(d => { d.tickFont = parseInt(e.target.value) || 12; })}
                    placeholder="Font size (default: 12)"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Show grid lines</label>
                  <Switch
                    checked={options.showGrid !== false}
                    onCheckedChange={(checked) => handleChange(d => { d.showGrid = checked; })}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Show Y-axis</label>
                  <Switch
                    checked={options.yaxis || false}
                    onCheckedChange={(checked) => handleChange(d => { d.yaxis = checked; })}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Stack bars</label>
                  <Switch
                    checked={options.stackId || false}
                    onCheckedChange={(checked) => handleChange(d => { d.stackId = checked; })}
                  />
                </div>

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
              </div>
            ),
          },
        ]}
      />
    </div>
  );
};
