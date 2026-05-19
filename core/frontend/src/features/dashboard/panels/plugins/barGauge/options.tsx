import {Accordion} from '@pharos/shared/components/ui-extension';
import {unitTypeList} from '@pharos/shared/lib/unitUtils';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {BarGaugePanelOptions} from './types';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import React from 'react';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<BarGaugePanelOptions>> = ({
  options = {},
  onOptionsChange,
  info,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div>
      <Accordion
        id={info?.id || 'barGauge'}
        items={[
          {
            id: 'barchartOptions',
            title: 'Bar chart options',
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
                </div>

                {/* Unit */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">Unit</label>
                  <Select
                    value={options.unit || 'default'}
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

                {/* Custom message */}
                <div className="flex flex-col gap-2">
                  <label className="text-sm text-muted-foreground">
                    Chart data not exist message
                  </label>
                  <Input
                    value={options.chartDataNotExistMessage || ''}
                    onChange={(e) => handleChange(d => { d.chartDataNotExistMessage = e.target.value; })}
                    placeholder="Enter custom message"
                  />
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
