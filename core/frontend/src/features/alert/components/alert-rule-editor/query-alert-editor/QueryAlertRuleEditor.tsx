/**
 * Query Alert Rule Editor - React Hook Form 버전
 * 
 * - React Hook Form + zodResolver
 * - 타입 안전성 보장
 * - 자동 검증
 */

'use client';

import React from 'react';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {Button, Input, Label, Textarea} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {Database, Clock, AlertTriangle} from 'lucide-react';
import type {QueryAlertRule} from '@pharos/shared/types/alert';
import {QueryAlertRuleFormSchema, QueryAlertRuleForm} from '@/features/alert/types/form';
import {ThresholdListEditor} from './ThresholdListEditor';
import {QueryPreview} from './QueryPreview';
import {useDatasourceList} from '@/hooks/use-datasource-list';

interface QueryAlertRuleEditorProps {
  initialData?: Partial<QueryAlertRule>;
  onSubmit: (data: QueryAlertRuleForm) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

export function QueryAlertRuleEditor({
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: QueryAlertRuleEditorProps) {
  const {dataSourceList, isLoading: loadingDatasources} = useDatasourceList();

  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: {errors},
  } = useForm<QueryAlertRuleForm>({
    resolver: zodResolver(QueryAlertRuleFormSchema),
    defaultValues: {
      alert_type: 'query',
      name: initialData?.name || '',
      datasource: initialData?.datasource || '',
      datasource_query: initialData?.datasource_query || {
        query: '',
        time_label: 'time',
        variable_label: 'value',
      },
      evaluation_interval: initialData?.evaluation_interval || '@every 30s',
      check_type: initialData?.check_type || 'last',
      threshold: initialData?.threshold || [],
      notifications: initialData?.notifications || [],
      description: initialData?.description || '',
    },
  });

  // Watch datasource and query for preview
  const datasource = watch('datasource');
  const query = watch('datasource_query.query');

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Basic Information */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <AlertTriangle className="h-5 w-5" />
          Basic Information
        </h3>
        <div className="space-y-4">
          <div>
            <Label htmlFor="name" className="text-sm font-normal">Alert Name *</Label>
            <Input
              id="name"
              {...register('name')}
              placeholder="e.g., High CPU Usage Alert"
              className={errors.name ? 'border-red-500' : ''}
            />
            {errors.name && <p className="text-sm text-red-500 mt-1">{errors.name.message}</p>}
          </div>

          <div>
            <Label htmlFor="description" className="text-sm font-normal">Description</Label>
            <Textarea
              id="description"
              {...register('description')}
              placeholder="Describe when this alert should trigger..."
              rows={3}
            />
            {errors.description && (
              <p className="text-sm text-red-500 mt-1">{errors.description.message}</p>
            )}
          </div>
        </div>
      </div>

      {/* Datasource Configuration */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <Database className="h-5 w-5" />
          Data Source
        </h3>
        <div className="space-y-4">
          <div>
            <Label htmlFor="datasource" className="text-sm font-normal">Datasource *</Label>
            <Controller
              name="datasource"
              control={control}
              render={({field}) => (
                <Select value={field.value} onValueChange={field.onChange} disabled={loadingDatasources}>
                  <SelectTrigger className={errors.datasource ? 'border-red-500' : ''}>
                    <SelectValue placeholder={loadingDatasources ? 'Loading...' : 'Select datasource'} />
                  </SelectTrigger>
                  <SelectContent>
                    {dataSourceList.map((ds) => (
                      <SelectItem key={ds.value} value={ds.value}>
                        {ds.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            {errors.datasource && (
              <p className="text-sm text-red-500 mt-1">{errors.datasource.message}</p>
            )}
          </div>

          <div>
            <Label htmlFor="query" className="text-sm font-normal">Query *</Label>
            <Textarea
              id="query"
              {...register('datasource_query.query')}
              placeholder="e.g., up{job='api'} == 0"
              rows={4}
              className={errors.datasource_query?.query ? 'border-red-500' : ''}
            />
            {errors.datasource_query?.query && (
              <p className="text-sm text-red-500 mt-1">
                {errors.datasource_query.query.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="time_label" className="text-sm font-normal">Time Label *</Label>
              <Input
                id="time_label"
                {...register('datasource_query.time_label')}
                placeholder="time"
              />
              {errors.datasource_query?.time_label && (
                <p className="text-sm text-red-500 mt-1">
                  {errors.datasource_query.time_label.message}
                </p>
              )}
            </div>

            <div>
              <Label htmlFor="variable_label" className="text-sm font-normal">Variable Label *</Label>
              <Input
                id="variable_label"
                {...register('datasource_query.variable_label')}
                placeholder="value"
              />
              {errors.datasource_query?.variable_label && (
                <p className="text-sm text-red-500 mt-1">
                  {errors.datasource_query.variable_label.message}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Query Preview - Datasource와 Query 사이에 배치 */}
      <div className="space-y-4">
        <QueryPreview
          datasource={datasource}
          query={query}
          disabled={!datasource || !query}
        />
      </div>

      {/* Evaluation Settings */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <Clock className="h-5 w-5" />
          Evaluation Settings
        </h3>
        <div className="space-y-4">
          <div>
            <Label htmlFor="evaluation_interval" className="text-sm font-normal">Evaluation Interval *</Label>
            <Controller
              name="evaluation_interval"
              control={control}
              render={({field}) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select interval" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="@every 10s">Every 10 seconds</SelectItem>
                    <SelectItem value="@every 30s">Every 30 seconds</SelectItem>
                    <SelectItem value="@every 1m">Every 1 minute</SelectItem>
                    <SelectItem value="@every 5m">Every 5 minutes</SelectItem>
                    <SelectItem value="@every 10m">Every 10 minutes</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
            {errors.evaluation_interval && (
              <p className="text-sm text-red-500 mt-1">{errors.evaluation_interval.message}</p>
            )}
          </div>

          <div>
            <Label htmlFor="check_type" className="text-sm font-normal">Check Type</Label>
            <Controller
              name="check_type"
              control={control}
              render={({field}) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="last">Last Value</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
          </div>
        </div>
      </div>

      {/* Thresholds */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Thresholds *</h3>
        <div>
          <Controller
            name="threshold"
            control={control}
            render={({field}) => (
              <ThresholdListEditor value={field.value} onChange={field.onChange} />
            )}
          />
          {errors.threshold && (
            <p className="text-sm text-red-500 mt-2">{errors.threshold.message}</p>
          )}
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
