'use client';

import React from 'react';
import {Card, CardContent} from '@pharos/shared/components/ui';
import {Button, Input, Label} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {Trash2} from 'lucide-react';
import type {Threshold, Severity, Condition} from '@pharos/shared/types/alert';
import {SEVERITY_OPTIONS, CONDITION_OPTIONS, getSeverityColor} from './threshold-constants';

interface ThresholdEditorProps {
  threshold: Threshold;
  onUpdate: (updates: Partial<Threshold>) => void;
  onRemove: () => void;
}

export function ThresholdEditor({threshold, onUpdate, onRemove}: ThresholdEditorProps) {
  const selectedCondition = CONDITION_OPTIONS.find((opt) => opt.value === threshold.condition);
  const needsEndValue =
    threshold.condition === 'within_range' || threshold.condition === 'outside_range';

  return (
    <Card className="border-l-4" style={{borderLeftColor: getSeverityColor(threshold.severity)}}>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4 flex-1">
              <div className="space-y-1 flex-1 max-w-xs">
                <Label>Severity</Label>
                <Select
                  value={threshold.severity}
                  onValueChange={(value) => onUpdate({severity: value as Severity})}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SEVERITY_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        <span className={opt.color}>{opt.label}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1 flex-1">
                <Label>Condition</Label>
                <Select
                  value={threshold.condition}
                  onValueChange={(value) => onUpdate({condition: value as Condition})}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {CONDITION_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        <div>
                          <div className="font-medium">{opt.label}</div>
                          <div className="text-xs text-muted-foreground">{opt.description}</div>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <Button variant="ghost" size="icon" onClick={onRemove} className="text-destructive">
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <Label>{needsEndValue ? 'Start Value' : 'Threshold Value'}</Label>
              <Input
                type="number"
                value={threshold.start_value}
                onChange={(e) => onUpdate({start_value: parseFloat(e.target.value) || 0})}
                placeholder="0"
              />
            </div>

            {needsEndValue && (
              <div className="space-y-1">
                <Label>End Value</Label>
                <Input
                  type="number"
                  value={threshold.end_value || 0}
                  onChange={(e) => onUpdate({end_value: parseFloat(e.target.value) || 0})}
                  placeholder="0"
                />
              </div>
            )}
          </div>

          {needsEndValue &&
            threshold.end_value !== undefined &&
            threshold.start_value > threshold.end_value && (
              <div className="text-sm text-destructive">
                ⚠️ Start value must be less than or equal to end value
              </div>
            )}

          <div className="text-sm text-muted-foreground">
            {selectedCondition?.description}
            {': '}
            <span className="font-mono">
              {threshold.condition === 'is_above' && `value > ${threshold.start_value}`}
              {threshold.condition === 'is_below' && `value < ${threshold.start_value}`}
              {threshold.condition === 'within_range' &&
                `${threshold.start_value} ≤ value ≤ ${threshold.end_value || 0}`}
              {threshold.condition === 'outside_range' &&
                `value < ${threshold.start_value} OR value > ${threshold.end_value || 0}`}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
