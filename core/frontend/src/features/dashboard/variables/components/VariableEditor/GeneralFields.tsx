'use client';

import React from 'react';
import {Label} from '@pharos/shared/components/ui';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import type {FilterConfig, Kind} from '@pharos/shared/types/dashboard';
import { Title } from "@pharos/shared/components/ui-extension";

/**
 * Variable placement 선택 옵션
 */
const VARIABLE_KINDS: Array<{value: Kind; label: string}> = [
  {value: 'headerL', label: 'Top left'},
  {value: 'headerR', label: 'Top right'},
];

export interface GeneralFieldsProps {
  data: FilterConfig;
  onChange: React.Dispatch<React.SetStateAction<FilterConfig>>;
}

export function GeneralFields({data, onChange}: GeneralFieldsProps) {
  const handleChange = (field: keyof FilterConfig, value: string) => {
    onChange((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleKindChange = (value: Kind) => {
    onChange((prev) => ({
      ...prev,
      kind: value,
    }));
  };

  const handleDescriptionChange = (value: string) => {
    onChange((prev) => ({
      ...prev,
      options: {
        ...prev.options,
        description: value,
      },
    }));
  };

  return (
    <div className="space-y-4 mt-8 mb-0">
      <div className="max-w-3xl">
        <Title variant="h3" >General</Title>
        <Label htmlFor="id" className="text-sm font-normal">
          ID <span className="text-red-500">*</span>
        </Label>
        <p className="text-xs text-muted-foreground mt-1 mb-2">
          When Variable Type is &quot;Query&quot;, this ID will be used as a filter reference in
          dashboard queries.
          <br />
          Example: Use ID &quot;node&quot; to reference as{' '}
          <code className="bg-input px-1 rounded">{'{{node}}'}</code> in your queries.
        </p>
        <Input
          id="id"
          value={data.id || ''}
          onChange={(e) => handleChange('id', e.target.value)}
          placeholder="Enter ID (e.g., node, region)"

        />
      </div>

      <div>
        <Label htmlFor="kind" className="text-sm font-normal">
          Variable placement
        </Label>
        <Select value={data.kind} onValueChange={handleKindChange}>
          <SelectTrigger className="mt-2 w-60">
            <SelectValue placeholder="Select placement" />
          </SelectTrigger>
          <SelectContent>
            {VARIABLE_KINDS.map((position) => (
              <SelectItem key={position.value} value={position.value}>
                {position.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex flex-col">
        <Label htmlFor="description" className="text-sm font-normal">
          Description
        </Label>
        <Input
          id="description"
          value={data.options?.description || ''}
          onChange={(e) => handleDescriptionChange(e.target.value)}
          placeholder="Enter description"
          className="mt-2 max-w-3xl"
        />
      </div>
    </div>
  );
}

export default GeneralFields;
