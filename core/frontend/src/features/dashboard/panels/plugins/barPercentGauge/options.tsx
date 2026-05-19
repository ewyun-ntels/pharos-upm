import React from 'react';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {BarPercentGaugePanelOptions} from './types';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<BarPercentGaugePanelOptions>> = ({
  options = {},
  onOptionsChange,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">Legend Rule</label>
        <Input
          value={options.legendRule || ''}
          onChange={(e) => handleChange(d => { d.legendRule = e.target.value; })}
          placeholder="Enter legend rule"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Gauge Max Value</label>
        <Input
          type="number"
          value={options.gaugeMaxValue || ''}
          onChange={(e) => handleChange(d => { d.gaugeMaxValue = Number(e.target.value); })}
          placeholder="Enter max value"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Sub Name</label>
        <Input
          value={options.subName || ''}
          onChange={(e) => handleChange(d => { d.subName = e.target.value; })}
          placeholder="Enter sub name"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Show Type</label>
        <Select
          value={options.showType || 'last'}
          onValueChange={(value) => handleChange(d => { d.showType = value; })}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select show type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="min">Min</SelectItem>
            <SelectItem value="max">Max</SelectItem>
            <SelectItem value="last">Last</SelectItem>
            <SelectItem value="average">Average</SelectItem>
            <SelectItem value="total">Total</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Chart Color</label>
        <Input
          value={options.chartColor || ''}
          onChange={(e) => handleChange(d => { d.chartColor = e.target.value; })}
          placeholder="Enter chart color"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Unit Type</label>
        <Input
          value={options.unitType || ''}
          onChange={(e) => handleChange(d => { d.unitType = e.target.value; })}
          placeholder="Enter unit type"
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">No Data Message</label>
        <Input
          value={options.chartDataNotExistMessage || ''}
          onChange={(e) => handleChange(d => { d.chartDataNotExistMessage = e.target.value; })}
          placeholder="Custom message when data doesn't exist"
        />
      </div>
    </div>
  );
};
