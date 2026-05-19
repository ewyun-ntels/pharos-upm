import {Accordion} from '@pharos/shared/components/ui-extension';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  ToggleGroup,
  ToggleGroupItem,
} from '@pharos/shared/components/ui';
import {Trash2, Plus} from 'lucide-react';
import {Button} from '@pharos/shared/components/ui';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {StatPanelOptions} from './types';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<StatPanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'stat'}
        items={[
          {
            id: 'StatStyles',
            title: 'Stat Styles',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Text mode</label>
                  <Select
                    value={options.textMode || 'auto'}
                    onValueChange={(value) => {
                      if (['auto', 'value', 'value_and_name', 'name', 'none'].includes(value)) {
                        handleChange(d => { d.textMode = value as StatPanelOptions['textMode']; });
                      }
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select text mode" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="auto">Auto</SelectItem>
                      <SelectItem value="value">Value</SelectItem>
                      <SelectItem value="value_and_name">Value and name</SelectItem>
                      <SelectItem value="name">Name</SelectItem>
                      <SelectItem value="none">None</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {options.textMode === 'value_and_name' && (
                  <div className="flex flex-col gap-2">
                    <div className="flex flex-col gap-2">
                      <label className="text-sm text-muted-foreground">Wide layout</label>
                      <ToggleGroup
                        type="single"
                        variant="outline"
                        size="sm"
                        value={options.wideLayout ? 'on' : 'off'}
                        onValueChange={(value) => {
                          if (value) {
                            handleChange((d) => {
                              d.wideLayout = value === 'on';
                            });
                          }
                        }}
                        className="justify-start w-[200px] bg-background"
                      >
                        <ToggleGroupItem value="on" aria-label="On" className="flex-1">On</ToggleGroupItem>
                        <ToggleGroupItem value="off" aria-label="Off" className="flex-1">Off</ToggleGroupItem>
                      </ToggleGroup>
                    </div>
                  </div>
                )}

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Text alignment</label>
                  <ToggleGroup
                    type="single"
                    variant="outline"
                    size="sm"
                    value={options.textAlignment === 'justify' ? 'justify' : 'center'}
                    onValueChange={(value) => {
                      if (value) {
                        handleChange(d => { d.textAlignment = value as StatPanelOptions['textAlignment']; });
                      }
                    }}
                    className="justify-start w-[200px] bg-background"
                  >
                    <ToggleGroupItem value="justify" aria-label="Auto" className="flex-1">Auto</ToggleGroupItem>
                    <ToggleGroupItem value="center" aria-label="Center" className="flex-1">Center</ToggleGroupItem>
                  </ToggleGroup>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Graph mode</label>
                  <ToggleGroup
                    type="single"
                    variant="outline"
                    size="sm"
                    value={options.chartType === 'none' ? 'none' : 'area'}
                    onValueChange={(value) => {
                      if (value) {
                        handleChange(d => { d.chartType = value as StatPanelOptions['chartType']; });
                      }
                    }}
                    className="justify-start w-[200px] bg-background"
                  >
                    <ToggleGroupItem value="none" aria-label="None" className="flex-1">None</ToggleGroupItem>
                    <ToggleGroupItem value="area" aria-label="Area" className="flex-1">Area</ToggleGroupItem>
                  </ToggleGroup>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Chart color</label>
                  <Input
                    type="color"
                    value={options.chartColor || '#8884d8'}
                    onChange={(e) => handleChange(d => { d.chartColor = e.target.value; })}
                  />
                </div>
              </div>
            )
          },
          {
            id: 'StatOptions',
            title: 'Stat Options',
            content: (
              <div className="space-y-4">
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Display name</label>
                  <Input
                    placeholder="name"
                    value={options.displayName || ''}
                    onChange={(e) => handleChange(d => { d.displayName = e.target.value; })}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Show type</label>
                  <Select
                    value={options.showType || 'last'}
                    onValueChange={(value) => {
                      if (value === 'last' || value === 'first' || value === 'average' || value === 'max' || value === 'min') {
                        handleChange(d => { d.showType = value; });
                      } else {
                        console.error(`Invalid showType value: ${value}`);
                      }
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select show type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="last">Last</SelectItem>
                      <SelectItem value="first">First</SelectItem>
                      <SelectItem value="average">Average</SelectItem>
                      <SelectItem value="max">Max</SelectItem>
                      <SelectItem value="min">Min</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Fill opacity</label>
                  <Input
                    type="number"
                    min="0"
                    max="1"
                    step="0.1"
                    value={options.fillOpacity || 0.5}
                    onChange={(e) => handleChange(d => { d.fillOpacity = parseFloat(e.target.value); })}
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Decimals</label>
                  <Input
                    type="number"
                    min="0"
                    max="10"
                    step="1"
                    placeholder="Auto"
                    value={options.decimals ?? ''}
                    onChange={(e) => {
                      const val = e.target.value;
                      handleChange(d => {
                        d.decimals = val === '' ? undefined : parseInt(val, 10);
                      });
                    }}
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
                <div className="flex flex-col gap-2">
                  <div className="space-y-2">
                    {(options.thresholds || []).map((t, index) => (
                      <div key={index} className="flex items-center gap-2">
                        <Input
                          type="color"
                          value={t.color}
                          onChange={(e) => {
                            const newColor = e.target.value;
                            handleChange(d => {
                              if (d.thresholds) d.thresholds[index].color = newColor;
                            });
                          }}
                          className="w-12 h-8 p-1"
                        />
                        <Input
                          type="number"
                          value={t.value}
                          onChange={(e) => {
                            const val = parseFloat(e.target.value);
                            handleChange(d => {
                              if (!isNaN(val) && d.thresholds) d.thresholds[index].value = val;
                            });
                          }}
                          className="flex-1 h-8"
                          placeholder="Value"
                        />
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-destructive"
                          onClick={() => {
                            handleChange(d => {
                              if (d.thresholds) d.thresholds.splice(index, 1);
                            });
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                  <Button
                    variant="outline"
                    size="default"
                    onClick={() => {
                      handleChange(d => {
                        if (!d.thresholds) d.thresholds = [];
                        d.thresholds.push({ value: 0, color: '#ff0000' });
                      });
                    }}
                  >
                    <Plus className="h-4 w-4" /> Add Threshold
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
