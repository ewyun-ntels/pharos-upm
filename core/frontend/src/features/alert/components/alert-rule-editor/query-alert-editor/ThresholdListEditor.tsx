/**
 * Threshold List Editor
 * 
 * React Hook Form 호환 - Threshold 배열 관리
 */

'use client';

import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {Plus} from 'lucide-react';
import type {Threshold} from '@pharos/shared/types/alert';
import {ThresholdEditor} from './ThresholdEditor';

interface ThresholdListEditorProps {
  value: Threshold[];
  onChange: (thresholds: Threshold[]) => void;
}

export function ThresholdListEditor({value, onChange}: ThresholdListEditorProps) {
  const handleAddThreshold = () => {
    const newThreshold: Threshold = {
      id: `threshold-${Date.now()}`,
      condition: 'is_above',
      start_value: 0,
      severity: 'Normal',
    };
    onChange([...value, newThreshold]);
  };

  const handleUpdateThreshold = (index: number, updates: Partial<Threshold>) => {
    const updated = [...value];
    updated[index] = {...updated[index], ...updates};
    onChange(updated);
  };

  const handleRemoveThreshold = (index: number) => {
    const updated = value.filter((_, i) => i !== index);
    onChange(updated);
  };

  return (
    <div className="space-y-4">
      {value.length === 0 && (
        <div className="text-sm text-muted-foreground text-center py-8">
          No thresholds configured. Click &#34;Add Threshold&#34; to create one.
        </div>
      )}

      {value.map((threshold, index) => (
        <ThresholdEditor
          key={threshold.id}
          threshold={threshold}
          onUpdate={(updates) => handleUpdateThreshold(index, updates)}
          onRemove={() => handleRemoveThreshold(index)}
        />
      ))}

      <Button type="button" variant="outline" onClick={handleAddThreshold} className="w-full">
        <Plus className="h-4 w-4 mr-2" />
        Add Threshold
      </Button>
    </div>
  );
}
