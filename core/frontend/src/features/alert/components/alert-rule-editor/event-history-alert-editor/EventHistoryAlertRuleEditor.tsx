'use client';

import React, {useState} from 'react';
import {useForm} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {Button, Input, Label, Switch, Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@pharos/shared/components/ui';
import {History} from 'lucide-react';
import {EventHistoryAlertRuleFormSchema, type EventHistoryAlertRuleForm} from '@features/alert/types/form';
import parse from 'parse-duration';

interface EventHistoryAlertRuleEditorProps {
  initialData?: Partial<EventHistoryAlertRuleForm>;
  onSubmit: (data: EventHistoryAlertRuleForm) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

// Retention period options (human-readable format)
const RETENTION_OPTIONS = [
  {readable: '1h', seconds: 3600},
  {readable: '6h', seconds: 21600},
  {readable: '12h', seconds: 43200},
  {readable: '1d', seconds: 86400},
  {readable: '3d', seconds: 259200},
  {readable: '7d', seconds: 604800},
  {readable: '30d', seconds: 2592000},
] as const;

export function EventHistoryAlertRuleEditor({
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: EventHistoryAlertRuleEditorProps) {
  const [useCustomInput, setUseCustomInput] = useState(false);
  const [customInput, setCustomInput] = useState('1d');
  const [parseError, setParseError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: {errors},
  } = useForm<EventHistoryAlertRuleForm>({
    resolver: zodResolver(EventHistoryAlertRuleFormSchema),
    defaultValues: {
      alert_type: 'event-history',
      name: '',
      description: '',
      retention_period: 86400, // 1 day in seconds
      notifications: [],
      ...initialData,
    },
  });

  const retentionPeriod = watch('retention_period');

  // Handle custom input parsing
  const handleCustomInputChange = (value: string) => {
    setCustomInput(value);
    const parsed = parse(value, 's'); // Parse to seconds
    
    if (parsed && parsed > 0) {
      setValue('retention_period', parsed);
      setParseError(null);
    } else {
      setParseError('Invalid format (e.g., 2h, 30m, 7d, 1w)');
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <History className="h-5 w-5" />
          Event History Alert Configuration
        </h3>
        <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name" className="text-sm font-normal">Rule Name *</Label>
              <Input
                id="name"
                placeholder="e.g., job-history-cleanup"
                {...register('name')}
              />
              {errors.name && (
                <p className="text-sm text-destructive">{errors.name.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description" className="text-sm font-normal">Description</Label>
              <Input
                id="description"
                placeholder="Optional description"
                {...register('description')}
              />
              {errors.description && (
                <p className="text-sm text-destructive">{errors.description.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="retention_period" className="text-sm font-normal">Retention Period *</Label>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">
                    {useCustomInput ? 'Custom input' : 'Quick select'}
                  </span>
                  <Switch checked={useCustomInput} onCheckedChange={setUseCustomInput} />
                </div>
              </div>

              <div className="flex items-center gap-2">
                {useCustomInput ? (
                  <Input
                    placeholder="e.g., 2h, 30m, 7d, 1w"
                    value={customInput}
                    onChange={(e) => handleCustomInputChange(e.target.value)}
                    className="flex-1"
                  />
                ) : (
                  <Select
                    value={String(retentionPeriod)}
                    onValueChange={(value) => setValue('retention_period', Number(value))}
                  >
                    <SelectTrigger className="flex-1">
                      <SelectValue placeholder="Select retention period" />
                    </SelectTrigger>
                    <SelectContent>
                      {RETENTION_OPTIONS.map((option) => (
                        <SelectItem key={option.seconds} value={String(option.seconds)}>
                          {option.readable}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
                
                <span className="text-sm text-muted-foreground whitespace-nowrap min-w-[100px]">
                  {(retentionPeriod || 0).toLocaleString()}s
                </span>
              </div>

              {parseError && (
                <p className="text-sm text-destructive">{parseError}</p>
              )}

              {useCustomInput && !parseError && (
                <p className="text-xs text-muted-foreground">
                  Supported units: s (seconds), m (minutes), h (hours), d (days), w (weeks)
                </p>
              )}

              {/* Hidden input for form submission */}
              <input type="hidden" {...register('retention_period', {valueAsNumber: true})} />

              {errors.retention_period && (
                <p className="text-sm text-destructive">{errors.retention_period.message}</p>
              )}
            </div>

            <div className="text-sm text-muted-foreground p-4 bg-muted rounded-md">
              <p className="font-medium mb-2">ℹ️ Event History Alert</p>
              <p>
                This alert type automatically cleans up alert status records older than the
                retention period. Events can be submitted via POST /alert/event/:name endpoint.
              </p>
            </div>
        </div>
      </div>

      {/* Buttons */}
      <div className="space-x-3">
        <Button
          type="button"
          variant="outline"
          className="h-8 shadow-none px-3 py-1"
          onClick={onCancel}
        >
          Cancel
        </Button>
        <Button
          type="submit"
          variant="default"
          className="h-8 shadow-none px-3 py-1"
          disabled={isLoading}
        >
          {isLoading ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </form>
  );
}
