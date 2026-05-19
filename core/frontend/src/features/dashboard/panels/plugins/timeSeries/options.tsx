import { Accordion } from '@pharos/shared/components/ui-extension';
import { Button } from '@pharos/shared/components/ui';
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
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';
import { TimeSeriesPanelOptions } from '@pharos/shared/components/charts';
import { unitTypeList } from '@pharos/shared/lib/unitUtils';
import React from 'react';
import { useOptionsChange } from '../common/useOptionsChange';
import { useThresholdForm } from '../common/useThresholdForm';

export const Options: React.FC<PanelEditorOptionsProps<TimeSeriesPanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);
  const { newThreshold, setNewThreshold, handleAddThreshold, handleRemoveThreshold, handleUpdateThreshold } = useThresholdForm(handleChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'timeSeries'}
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
                    checked={options.legendEnabled || false}
                    onCheckedChange={(checked) => handleChange(d => { d.legendEnabled = checked; })}
                  />
                </div>

                {options.legendEnabled && (
                  <>
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Placement</label>
                      <Tabs
                        value={options.legendAlign || 'right'}
                        onValueChange={(value) => {
                          if (value === 'right' || value === 'bottom') {
                            handleChange(d => { d.legendAlign = value; });
                          } else {
                            console.error(`Invalid legendAlign value: ${value}`);
                          }
                        }}
                      >
                        <TabsList className="grid w-full grid-cols-2">
                          <TabsTrigger value="right">Right</TabsTrigger>
                          <TabsTrigger value="bottom">Bottom</TabsTrigger>
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
                  <label className="text-sm text-muted-foreground">Chart type</label>
                  <Tabs
                    value={options.chartType || 'line'}
                    onValueChange={(value) => {
                      if (value === 'line' || value === 'points' || value === 'bars' || value === 'area') {
                        handleChange(d => { d.chartType = value; });
                      }
                    }}
                  >
                    <TabsList className="grid w-full grid-cols-4">
                      <TabsTrigger value="line">Line</TabsTrigger>
                      <TabsTrigger value="points">Points</TabsTrigger>
                      <TabsTrigger value="bars">Bars</TabsTrigger>
                      <TabsTrigger value="area">Area</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>

                {/* Line-specific options */}
                {(options.chartType === 'line' || !options.chartType) && (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Show points</label>
                      <Switch
                        checked={options.showPoints !== false}
                        onCheckedChange={(checked) => handleChange(d => { d.showPoints = checked; })}
                      />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Line width: {options.lineWidth ?? 1}px</label>
                      <Slider
                        value={[options.lineWidth ?? 1]}
                        onValueChange={([value]) => handleChange(d => { d.lineWidth = value; })}
                        min={1}
                        max={5}
                        step={1}
                      />
                    </div>
                  </div>
                )}

                {/* Points-specific options */}
                {options.chartType === 'points' && (
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Point size: {options.pointSize ?? 8}px</label>
                    <Slider
                      value={[options.pointSize ?? 8]}
                      onValueChange={([value]) => handleChange(d => { d.pointSize = value; })}
                      min={4}
                      max={20}
                      step={1}
                    />
                  </div>
                )}

                {/* Bars-specific options */}
                {options.chartType === 'bars' && (
                  <div className="space-y-3">
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Bar width: {Math.round((options.barWidth ?? 0.6) * 100)}%</label>
                      <Slider
                        value={[Math.round((options.barWidth ?? 0.6) * 100)]}
                        onValueChange={([value]) => handleChange(d => { d.barWidth = value / 100; })}
                        min={10}
                        max={100}
                        step={5}
                      />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Fill opacity: {Math.round((options.fillOpacity ?? 1) * 100)}%</label>
                      <Slider
                        value={[Math.round((options.fillOpacity ?? 1) * 100)]}
                        onValueChange={([value]) => handleChange(d => { d.fillOpacity = value / 100; })}
                        min={0}
                        max={100}
                        step={5}
                      />
                    </div>
                    <div className="flex items-center justify-between">
                      <label className="text-sm text-muted-foreground">Stack bars</label>
                      <Switch
                        checked={options.barStack ?? false}
                        onCheckedChange={(checked) => handleChange(d => { d.barStack = checked; })}
                      />
                    </div>
                  </div>
                )}

                {/* Area-specific options */}
                {options.chartType === 'area' && (
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Fill opacity: {Math.round((options.fillOpacity ?? 0.1) * 100)}%</label>
                    <Slider
                      value={[Math.round((options.fillOpacity ?? 0.1) * 100)]}
                      onValueChange={([value]) => handleChange(d => { d.fillOpacity = value / 100; })}
                      min={0}
                      max={100}
                      step={5}
                    />
                  </div>
                )}

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Decimals</label>
                  <Input
                    type="number"
                    min={0}
                    max={10}
                    value={options.decimals ?? 2}
                    onChange={(e) => handleChange(d => { d.decimals = Math.max(0, parseInt(e.target.value) || 0); })}
                    placeholder="2"
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Unit</label>
                  <Select
                    value={options.unit || 'none'}
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

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Cursor sync</label>
                  <Tabs
                    value={options.syncId === undefined ? 'default' : options.syncId === '' ? 'off' : 'custom'}
                    onValueChange={(value) => {
                      if (value === 'default') handleChange(d => { d.syncId = undefined; });
                      else if (value === 'off') handleChange(d => { d.syncId = ''; });
                    }}
                  >
                    <TabsList className="grid w-full grid-cols-3">
                      <TabsTrigger value="default">Default</TabsTrigger>
                      <TabsTrigger value="custom">Custom</TabsTrigger>
                      <TabsTrigger value="off">Off</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  {options.syncId !== undefined && options.syncId !== '' && (
                    <Input
                      value={options.syncId}
                      onChange={(e) => handleChange(d => { d.syncId = e.target.value; })}
                      placeholder="Group name"
                    />
                  )}
                  {options.syncId === undefined && (
                    <p className="text-xs text-muted-foreground">Cursor position syncs across all timeSeries panels on the same dashboard.</p>
                  )}
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Y-axis min</label>
                  <Input
                    value={options.scales?.yMin || ''}
                    onChange={(e) => handleChange(d => {
                      if (!d.scales) d.scales = {};
                      d.scales.yMin = e.target.value;
                    })}
                    placeholder="Auto"
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Y-axis max</label>
                  <Input
                    value={options.scales?.yMax || ''}
                    onChange={(e) => handleChange(d => {
                      if (!d.scales) d.scales = {};
                      d.scales.yMax = e.target.value;
                    })}
                    placeholder="Auto"
                  />
                </div>
              </div>
            ),
          },
          {
            id: 'TooltipOptions',
            title: 'Tooltip',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Show values</label>
                  <Tabs
                    value={(options.tooltipLimit ?? 1) > 1 ? 'all' : 'closest'}
                    onValueChange={(value) => {
                      handleChange(d => { d.tooltipLimit = value === 'all' ? 999 : 1; });
                    }}
                  >
                    <TabsList className="grid w-full grid-cols-2">
                      <TabsTrigger value="closest">Closest</TabsTrigger>
                      <TabsTrigger value="all">All</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <p className="text-xs text-muted-foreground">
                    {(options.tooltipLimit ?? 1) > 1
                      ? 'All series values are shown in the tooltip.'
                      : 'Only the series closest to the cursor is shown.'}
                  </p>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <label className="text-sm text-muted-foreground">Annotation button</label>
                    <p className="text-xs text-muted-foreground">Click chart to lock cursor and show button</p>
                  </div>
                  <Switch
                    checked={options.showAnnotationButton ?? true}
                    onCheckedChange={(checked) => handleChange(d => { d.showAnnotationButton = checked; })}
                  />
                </div>
              </div>
            ),
          },
          {
            id: 'ThresholdOptions',
            title: 'Thresholds',
            content: (
              <div className="space-y-4">
                {options.thresholds?.map((threshold, index) => (
                  <div key={index} className="space-y-2 p-3 border rounded">
                    {/* Row 1: Value + Color */}
                    <div className="flex items-center gap-2">
                      <Input
                        type="number"
                        className="flex-1"
                        value={threshold.value || ''}
                        onChange={(e) => handleUpdateThreshold(index, { value: e.target.value })}
                        placeholder="Value"
                      />
                      <Input
                        type="color"
                        className="w-10 h-9 p-1 cursor-pointer"
                        value={threshold.color || '#ff0000'}
                        onChange={(e) => handleUpdateThreshold(index, { color: e.target.value })}
                      />
                    </div>
                    {/* Row 2: Line type + Width */}
                    <div className="flex items-center gap-2">
                      <Select
                        value={threshold.type || 'solid'}
                        onValueChange={(value) => {
                          if (value === 'solid' || value === 'dash' || value === 'dot') {
                            handleUpdateThreshold(index, { type: value });
                          }
                        }}
                      >
                        <SelectTrigger className="flex-1">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="solid">Solid</SelectItem>
                          <SelectItem value="dash">Dash</SelectItem>
                          <SelectItem value="dot">Dot</SelectItem>
                        </SelectContent>
                      </Select>
                      <div className="flex items-center gap-1 w-24">
                        <Slider
                          min={1}
                          max={6}
                          step={1}
                          value={[threshold.lineWidth ?? 2]}
                          onValueChange={([v]) => handleUpdateThreshold(index, { lineWidth: v })}
                        />
                        <span className="text-xs text-muted-foreground w-4 text-right">{threshold.lineWidth ?? 2}</span>
                      </div>
                    </div>
                    <Button
                      variant="destructive"
                      className="w-full"
                      onClick={() => handleRemoveThreshold(index)}
                    >
                      Remove
                    </Button>
                  </div>
                ))}

                {/* Add new threshold */}
                <div className="space-y-2 p-3 border rounded">
                  <div className="flex items-center gap-2">
                    <Input
                      type="number"
                      className="flex-1"
                      value={newThreshold.value}
                      onChange={(e) => setNewThreshold({ ...newThreshold, value: e.target.value })}
                      placeholder="Value"
                    />
                    <Input
                      type="color"
                      className="w-10 h-9 p-1 cursor-pointer"
                      value={newThreshold.color}
                      onChange={(e) => setNewThreshold({ ...newThreshold, color: e.target.value })}
                    />
                  </div>
                  <div className="flex items-center gap-2">
                    <Select
                      value={newThreshold.type}
                      onValueChange={(value) => {
                        if (value === 'solid' || value === 'dash' || value === 'dot') {
                          setNewThreshold({ ...newThreshold, type: value });
                        }
                      }}
                    >
                      <SelectTrigger className="flex-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="solid">Solid</SelectItem>
                        <SelectItem value="dash">Dash</SelectItem>
                        <SelectItem value="dot">Dot</SelectItem>
                      </SelectContent>
                    </Select>
                    <div className="flex items-center gap-1 w-24">
                      <Slider
                        min={1}
                        max={6}
                        step={1}
                        value={[newThreshold.lineWidth ?? 2]}
                        onValueChange={([v]) => setNewThreshold({ ...newThreshold, lineWidth: v })}
                      />
                      <span className="text-xs text-muted-foreground w-4 text-right">{newThreshold.lineWidth ?? 2}</span>
                    </div>
                  </div>
                  <Button onClick={handleAddThreshold} className="w-full">
                    Add Threshold
                  </Button>
                </div>
              </div>
            ),
          },
        ]}
        defaultOpenItems={'all'}
        type="multiple"
      />
    </div>
  );
};
