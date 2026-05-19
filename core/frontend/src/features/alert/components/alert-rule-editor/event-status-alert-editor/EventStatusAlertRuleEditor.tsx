'use client';

import React from 'react';
import {useForm} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {Button, Input, Label} from '@pharos/shared/components/ui';
import {AlertCircle} from 'lucide-react';
import {EventStatusAlertRuleFormSchema, type EventStatusAlertRuleForm} from '@features/alert/types/form';

interface EventStatusAlertRuleEditorProps {
  initialData?: Partial<EventStatusAlertRuleForm>;
  onSubmit: (data: EventStatusAlertRuleForm) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

export function EventStatusAlertRuleEditor({
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: EventStatusAlertRuleEditorProps) {
  const {
    register,
    handleSubmit,
    formState: {errors},
  } = useForm<EventStatusAlertRuleForm>({
    resolver: zodResolver(EventStatusAlertRuleFormSchema),
    defaultValues: {
      alert_type: 'event-status',
      name: '',
      description: '',
      notifications: [],
      ...initialData,
    },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <AlertCircle className="h-5 w-5" />
          Event Status Alert Configuration
        </h3>
        <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name" className="text-sm font-normal">Rule Name *</Label>
              <Input
                id="name"
                placeholder="e.g., interface-status-alert"
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

            <div className="text-sm text-muted-foreground p-4 bg-muted rounded-md">
              <p className="font-medium mb-2">ℹ️ Event Status Alert</p>
              <p>
                This alert type listens for external events via POST /alert/event/:name endpoint.
                Events update the current status and trigger notifications when severity changes.
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
