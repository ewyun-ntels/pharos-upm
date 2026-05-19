/**
 * Append PDU Editor
 *
 * SNMP trap에 추가할 고정 PDU 편집기
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
import type { AppendPdu, ValueType } from '@pharos/shared/types/notification';

interface AppendPduEditorProps {
  value: AppendPdu[];
  onChange: (value: AppendPdu[]) => void;
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

export function AppendPduEditor({ value = [], onChange }: AppendPduEditorProps) {
  const handleAdd = () => {
    onChange([
      ...value,
      {
        oid: '',
        value_type: 'OctetString',
        value: '',
      },
    ]);
  };

  const handleRemove = (index: number) => {
    const newValue = [...value];
    newValue.splice(index, 1);
    onChange(newValue);
  };

  const handleChange = (index: number, field: keyof AppendPdu, val: any) => {
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
            Append PDUs (Fixed Values)
          </Label>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleAdd}
            className="h-8 shadow-none"
          >
            <Plus className="h-4 w-4 mr-1" />
            Add PDU
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Appends fixed PDU values to every SNMP trap. Use this to send constant information like device type, location, or environment.
          <br />
          <strong>Example:</strong> Send OID <code>1.3.6.1.4.1.12345.1.99</code> with value <code>production</code> to identify the environment.
        </p>
      </div>

      {value.length === 0 && (
        <div className="text-sm text-muted-foreground border border-dashed rounded-md p-4 text-center">
          No append PDUs. No additional fixed values will be sent.
        </div>
      )}

      {value.map((pdu, index) => (
        <div key={index} className="border rounded-md p-4 space-y-3">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">PDU #{index + 1}</span>
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
              <Label htmlFor={`append_oid_${index}`} className="text-xs">
                OID *
              </Label>
              <Input
                id={`append_oid_${index}`}
                value={pdu.oid}
                onChange={(e) => handleChange(index, 'oid', e.target.value)}
                placeholder="1.3.6.1.4.1..."
                className="h-9"
              />
            </div>

            <div>
              <Label htmlFor={`append_value_type_${index}`} className="text-xs">
                Value Type *
              </Label>
              <Select
                value={pdu.value_type}
                onValueChange={(val) => handleChange(index, 'value_type', val as ValueType)}
              >
                <SelectTrigger className="h-9">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {VALUE_TYPES.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor={`append_value_${index}`} className="text-xs">
                Value *
              </Label>
              <Input
                id={`append_value_${index}`}
                value={String(pdu.value)}
                onChange={(e) => {
                  // Convert to appropriate type based on value_type
                  const val = e.target.value;
                  if (
                    pdu.value_type === 'counter32' ||
                    pdu.value_type === 'counter64' ||
                    pdu.value_type === 'gauge32' ||
                    pdu.value_type === 'uinteger32'
                  ) {
                    handleChange(index, 'value', val ? parseInt(val, 10) : 0);
                  } else if (
                    pdu.value_type === 'opaqueFloat' ||
                    pdu.value_type === 'opaqueDouble'
                  ) {
                    handleChange(index, 'value', val ? parseFloat(val) : 0);
                  } else {
                    handleChange(index, 'value', val);
                  }
                }}
                placeholder="Fixed value"
                type={
                  pdu.value_type === 'counter32' ||
                  pdu.value_type === 'counter64' ||
                  pdu.value_type === 'gauge32' ||
                  pdu.value_type === 'uinteger32' ||
                  pdu.value_type === 'opaqueFloat' ||
                  pdu.value_type === 'opaqueDouble'
                    ? 'number'
                    : 'text'
                }
                className="h-9"
              />
            </div>
          </div>

          <p className="text-xs text-muted-foreground">
            Always sends OID <strong>{pdu.oid || '...'}</strong> with value{' '}
            <strong>{pdu.value || '...'}</strong> ({pdu.value_type})
          </p>
        </div>
      ))}
    </div>
  );
}
