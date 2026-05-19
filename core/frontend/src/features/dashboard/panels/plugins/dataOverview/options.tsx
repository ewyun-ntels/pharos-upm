import React from 'react';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {DataOverviewPanelOptions} from './types';
import {Input} from '@pharos/shared/components/ui';
import {Switch} from '@pharos/shared/components/ui';
import {Label} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<DataOverviewPanelOptions>> = ({
  options = {},
  onOptionsChange,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  const handleDataChange = (index: number, field: 'name' | 'path', value: string) => {
    handleChange(draft => {
      if (!draft.datas) draft.datas = [];
      draft.datas[index] = {...draft.datas[index], [field]: value};
    });
  };

  const addData = () => {
    handleChange(draft => {
      if (!draft.datas) draft.datas = [];
      draft.datas.push({name: '', path: ''});
    });
  };

  const removeData = (index: number) => {
    handleChange(draft => {
      draft.datas = draft.datas?.filter((_, i) => i !== index);
    });
  };

  return (
    <div className="space-y-4">
      <div>
        <Label className="mb-2">Data Paths</Label>
        {(options.datas || []).map((data, index) => (
          <div key={index} className="flex gap-2 mb-2">
            <Input
              placeholder="Name"
              value={data.name}
              onChange={(e) => handleDataChange(index, 'name', e.target.value)}
              className="flex-1"
            />
            <Input
              placeholder="JSONPath"
              value={data.path}
              onChange={(e) => handleDataChange(index, 'path', e.target.value)}
              className="flex-1"
            />
            <Button variant="destructive" size="sm" onClick={() => removeData(index)}>
              Remove
            </Button>
          </div>
        ))}
        <Button variant="outline" size="sm" onClick={addData}>
          Add Data Path
        </Button>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          checked={options.layoutCol || false}
          onCheckedChange={(checked) => handleChange(d => { d.layoutCol = checked; })}
        />
        <Label>Column Layout</Label>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          checked={options.titleTop || false}
          onCheckedChange={(checked) => handleChange(d => { d.titleTop = checked; })}
        />
        <Label>Title on Top</Label>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          checked={options.largeText || false}
          onCheckedChange={(checked) => handleChange(d => { d.largeText = checked; })}
        />
        <Label>Large Text</Label>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          checked={options.right || false}
          onCheckedChange={(checked) => handleChange(d => { d.right = checked; })}
        />
        <Label>Align Right</Label>
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
