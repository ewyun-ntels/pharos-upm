import {Accordion} from '@pharos/shared/components/ui-extension';
import {Button} from '@pharos/shared/components/ui';
import {Input} from '@pharos/shared/components/ui';
import {Switch} from '@pharos/shared/components/ui';
import {Tabs, TabsList, TabsTrigger} from '@pharos/shared/components/ui';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@pharos/shared/components/ui';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {BarChartPanelOptions} from './types';
import {unitTypeList} from '@pharos/shared/lib/unitUtils';
import {useOptionsChange} from '../common/useOptionsChange';
import {useThresholdForm} from '../common/useThresholdForm';
import React from 'react';

export const Options: React.FC<PanelEditorOptionsProps<BarChartPanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);
  const {newThreshold, setNewThreshold, handleAddThreshold, handleRemoveThreshold} = useThresholdForm(handleChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'barChart'}
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
                  <label className="text-sm text-muted-foreground">X Axis</label>
                  <Input
                    value={options.xAxis || ''}
                    onChange={(e) => handleChange(d => { d.xAxis = e.target.value; })}
                    placeholder="timestamp, index, or column name"
                  />
                  <p className="text-[11px] text-muted-foreground">
                    Specify the column to use for X-axis labels (leave empty for timestamp)
                  </p>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Bar Mode</label>
                  <Select
                    value={options.barMode || 'single'}
                    onValueChange={(value) => {
                      if (value === 'single' || value === 'grouped' || value === 'stacked' || value === 'percent') {
                        handleChange(d => { d.barMode = value; });
                      } else {
                        console.error(`Invalid barMode value: ${value}`);
                      }
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select bar mode" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="single">Single</SelectItem>
                      <SelectItem value="grouped">Grouped</SelectItem>
                      <SelectItem value="stacked">Stacked</SelectItem>
                      <SelectItem value="percent">100% Stacked</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Orientation</label>
                  <Tabs
                    value={options.orientation || 'auto'}
                    onValueChange={(value) => {
                      if (value === 'auto' || value === 'vertical' || value === 'horizontal') {
                        handleChange(d => { d.orientation = value; });
                      } else {
                        console.error(`Invalid orientation value: ${value}`);
                      }
                    }}
                  >
                    <TabsList className="grid w-full grid-cols-3">
                      <TabsTrigger value="auto">Auto</TabsTrigger>
                      <TabsTrigger value="vertical">Vertical</TabsTrigger>
                      <TabsTrigger value="horizontal">Horizontal</TabsTrigger>
                    </TabsList>
                  </Tabs>
                </div>

                {options.orientation === 'horizontal' && (
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Value Sort Order</label>
                    <Select
                      value={options.valueSortOrder || 'none'}
                      onValueChange={(value) => {
                        if (value === 'none' || value === 'asc' || value === 'desc') {
                          handleChange(d => { d.valueSortOrder = value; });
                        } else {
                          console.error(`Invalid valueSortOrder value: ${value}`);
                        }
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select sort order" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">None</SelectItem>
                        <SelectItem value="asc">Ascending</SelectItem>
                        <SelectItem value="desc">Descending</SelectItem>
                      </SelectContent>
                    </Select>
                    <p className="text-[11px] text-muted-foreground">
                      Sort bars by their values (based on first series)
                    </p>
                  </div>
                )}

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Bar Color Mode</label>
                  <Select
                    value={options.barColorMode || 'series'}
                    onValueChange={(value) => {
                      if (value === 'single' || value === 'category' || value === 'series') {
                        handleChange(d => { d.barColorMode = value; });
                      } else {
                        console.error(`Invalid barColorMode value: ${value}`);
                      }
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select color mode" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="single">Single Color</SelectItem>
                      <SelectItem value="category">Category Colors</SelectItem>
                      <SelectItem value="series">Series Colors</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-[11px] text-muted-foreground">
                    Single: all bars same color | Category: color by X-axis | Series: color by data series
                  </p>
                </div>

                {options.barColorMode === 'single' && (
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Bar Color</label>
                    <div className="flex items-center gap-2">
                      <Input
                        type="color"
                        value={options.barColor || '#4169E1'}
                        onChange={(e) => handleChange(d => { d.barColor = e.target.value; })}
                        className="w-20 h-10"
                      />
                      <Input
                        value={options.barColor || 'hsl(221, 83%, 53%)'}
                        onChange={(e) => handleChange(d => { d.barColor = e.target.value; })}
                        placeholder="hsl(221, 83%, 53%)"
                        className="flex-1"
                      />
                    </div>
                    <p className="text-[11px] text-muted-foreground">
                      Use color picker or enter HSL/RGB/HEX value
                    </p>
                  </div>
                )}

                <div className="flex items-center justify-between">
                  <label className="text-sm text-muted-foreground">Show value labels</label>
                  <Switch
                    checked={options.showValueLabels || false}
                    onCheckedChange={(checked) => handleChange(d => { d.showValueLabels = checked; })}
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
                  <label className="text-sm text-muted-foreground">Tooltip limit</label>
                  <Input
                    type="number"
                    min="1"
                    value={options.tooltipLimit || 1}
                    onChange={(e) => handleChange(d => { d.tooltipLimit = parseInt(e.target.value) || 1; })}
                    placeholder="1"
                  />
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
            id: 'VariableInteraction',
            title: 'Variable Interaction',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">On Click Variable</label>
                  <Input
                    value={options.onClickVariable || ''}
                    onChange={(e) => handleChange(d => { d.onClickVariable = e.target.value; })}
                    placeholder="Enter variable ID"
                  />
                  <p className="text-[11px] text-muted-foreground">
                    Variable ID to set with the legend value when a bar is clicked
                  </p>
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
                  <div key={index} className="flex items-center gap-2 p-2 border rounded">
                    <div className="flex-1">
                      <div className="text-sm">Value: {threshold.value}</div>
                      <div className="text-sm">Type: {threshold.type || 'solid'}</div>
                    </div>
                    <div
                      className="w-8 h-8 rounded border"
                      style={{backgroundColor: threshold.color}}
                    />
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleRemoveThreshold(index)}
                    >
                      Remove
                    </Button>
                  </div>
                ))}

                <div className="space-y-2 p-4 border rounded">
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Value</label>
                    <Input
                      type="number"
                      value={newThreshold.value}
                      onChange={(e) => setNewThreshold({...newThreshold, value: e.target.value})}
                      placeholder="Enter threshold value"
                    />
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Color</label>
                    <Input
                      type="color"
                      value={newThreshold.color}
                      onChange={(e) => setNewThreshold({...newThreshold, color: e.target.value})}
                    />
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-sm text-muted-foreground">Line type</label>
                    <Select
                      value={newThreshold.type}
                      onValueChange={(value) => {
                        if (value === 'solid' || value === 'dash' || value === 'dot') {
                          setNewThreshold({...newThreshold, type: value});
                        } else {
                          console.error(`Invalid threshold type value: ${value}`);
                        }
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="solid">Solid</SelectItem>
                        <SelectItem value="dash">Dash</SelectItem>
                        <SelectItem value="dot">Dot</SelectItem>
                      </SelectContent>
                    </Select>
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
