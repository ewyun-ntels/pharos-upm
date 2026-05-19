/**
 * Trap PDU Rule Editor
 *
 * Alert label을 SNMP PDU로 매핑하는 규칙 편집기
 */

'use client';

import React from 'react';
import { Button, Input, Label } from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import { Plus, Trash2 } from 'lucide-react';
import type { TrapPduRule, ValueType } from '@pharos/shared/types/notification';

interface TrapPduRuleEditorProps {
  value: TrapPduRule[];
  onChange: (value: TrapPduRule[]) => void;
}

const VALUE_TYPES: ValueType[] = [
  'OctetString',
  'counter32',
  'counter64',
  'gauge32',
  'uinteger32',
  'opaqueFloat',
  'opaqueDouble',
  'objectIdentifier',
];

export function TrapPduRuleEditor({ value = [], onChange }: TrapPduRuleEditorProps) {
  const handleAdd = () => {
    onChange([
      ...value,
      {
        label_name: '',
        oid: '',
        value_type: undefined,
      },
    ]);
  };

  const handleRemove = (index: number) => {
    const newValue = [...value];
    newValue.splice(index, 1);
    onChange(newValue);
  };

  const handleChange = (index: number, field: keyof TrapPduRule, val: any) => {
    const newValue = [...value];
    newValue[index] = {
      ...newValue[index],
      [field]: val,
    };
    onChange(newValue);
  };

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-normal">
            Trap PDU Rules (Label → OID Mapping)
          </Label>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleAdd}
            className="h-8 shadow-none"
          >
            <Plus className="h-4 w-4 mr-1" />
            Add Rule
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Maps alert labels to SNMP PDUs. When an alert is triggered, the label values will be sent as SNMP trap variables.
          <br />
          <strong>Example:</strong> Map label <code>severity</code> to OID <code>1.3.6.1.4.1.12345.1.1</code> to send alert severity level.
        </p>
      </div>

      {value.length === 0 && (
        <div className="text-sm text-muted-foreground border border-dashed rounded-md p-4 text-center">
          No trap PDU rules. Alert labels will not be mapped to SNMP PDUs.
        </div>
      )}

      {value.map((rule, index) => (
        <div key={index} className="border rounded-md p-4 space-y-3">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">Rule #{index + 1}</span>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => handleRemove(index)}
              className="h-8 w-8 p-0"
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <Label htmlFor={`label_name_${index}`} className="text-xs">
                Label Name *
              </Label>
              <Input
                id={`label_name_${index}`}
                value={rule.label_name}
                onChange={(e) => handleChange(index, 'label_name', e.target.value)}
                placeholder="e.g., severity"
                className="h-9"
              />
            </div>

            <div>
              <Label htmlFor={`oid_${index}`} className="text-xs">
                OID *
              </Label>
              <Input
                id={`oid_${index}`}
                value={rule.oid}
                onChange={(e) => handleChange(index, 'oid', e.target.value)}
                placeholder="1.3.6.1.4.1..."
                className="h-9"
              />
            </div>

            <div>
              <Label htmlFor={`value_type_${index}`} className="text-xs">
                Value Type (Optional)
              </Label>
              <Select
                value={rule.value_type || '__auto__'}
                onValueChange={(val) => handleChange(index, 'value_type', val === '__auto__' ? undefined : val)}
              >
                <SelectTrigger className="h-9">
                  <SelectValue placeholder="Auto-detect" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__auto__">Auto-detect</SelectItem>
                  {VALUE_TYPES.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <p className="text-xs text-muted-foreground">
            Maps alert label "<strong>{rule.label_name || '...'}</strong>" to OID{' '}
            <strong>{rule.oid || '...'}</strong>
          </p>
        </div>
      ))}
    </div>
  );
}
