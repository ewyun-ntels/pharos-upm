
import React from 'react';
import { AlertRuleEditor, useAlertRule } from '@features/alert';
import {PageHeader} from '@components/page-header';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useCreate, useUpdate } from '@/lib/data-provider';
import { useToast } from '@hooks/use-toast';
import { ALERT_RESOURCES, ALERT_PROVIDER_NAME } from '@providers/alert-provider/types';
import { ScrollArea } from '@pharos/shared/components/ui/scroll-area';

/**
 * Alert Rule Edit Page
 * 
 * Create new or edit existing alert rule
 * - URL: /alert/edit (create new)
 * - URL: /alert/edit?id=xxx (edit existing)
 */
export default function AlertRuleEdit() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { toast } = useToast();
  const ruleId = searchParams?.get('id');

  // Fetch existing alert rule with validation (only when id exists)
  // Custom hook이 자동으로:
  // 1. API 호출 (ruleId 있을 때만)
  // 2. Zod schema로 데이터 검증
  // 3. 타입 안정성 보장 (AlertRule 타입)
  const {
    data: alertRule,
    isLoading: isFetching,
    validationError,
  } = useAlertRule({
    id: ruleId || null,
    enabled: !!ruleId,
  });

  const { mutate: createRule, mutation: createMutation } = useCreate();
  const { mutate: updateRule, mutation: updateMutation } = useUpdate();
  const isCreating = createMutation.isPending;
  const isUpdating = updateMutation.isPending;

  const alertType = alertRule?.alert_type;

  // Validation Error 처리
  React.useEffect(() => {
    if (validationError) {
      toast({
        variant: 'destructive',
        title: 'Data Validation Error',
        description: 'The alert rule data format is invalid. Please contact support.',
      });
    }
  }, [validationError, toast]);

  const handleCancel = () => {
    navigate('/alert');
  };

  const handleSubmit = (data: any) => {
    const alertType = data.alert_type;

    // Common validation
    if (!data.name) {
      toast({ variant: 'destructive', description: 'Rule name is required' });
      return;
    }

    // Type-specific validation
    if (alertType === 'query') {
      if (!data.datasource) {
        toast({ variant: 'destructive', description: 'Datasource is required' });
        return;
      }
      if (!data.datasource_query?.query) {
        toast({ variant: 'destructive', description: 'Query is required' });
        return;
      }
      if (!data.threshold || data.threshold.length === 0) {
        toast({ variant: 'destructive', description: 'At least one threshold is required' });
        return;
      }
    } else if (alertType === 'event-history') {
      if (!data.retention_period || data.retention_period <= 0) {
        toast({ variant: 'destructive', description: 'Retention period must be greater than 0' });
        return;
      }
    }

    const payload = {
      resource: ALERT_RESOURCES.RULE,
      values: {
        alert_type: alertType,
        rule: data,
      },
      meta: {
        dataProviderName: ALERT_PROVIDER_NAME,
      },
    };

    // Create or Update based on ruleId
    if (ruleId) {
      // Update existing rule
      updateRule(
        {
          ...payload,
          id: ruleId,
        },
        {
          onSuccess: () => {
            toast({ description: 'Alert rule updated successfully' });
            navigate('/alert');
          },
          onError: () => {
            toast({ variant: 'destructive', description: 'Failed to update alert rule' });
          },
        },
      );
    } else {
      // Create new rule
      createRule(
        payload,
        {
          onSuccess: () => {
            toast({ description: 'Alert rule created successfully' });
            navigate('/alert');
          },
          onError: () => {
            toast({ variant: 'destructive', description: 'Failed to create alert rule' });
          },
        },
      );
    }
  };

  // ✅ 조건부 렌더링 (Hooks 이후)
  // Show loading only when fetching existing rule
  if (ruleId && isFetching) {
    return (
      <ScrollArea className="w-full h-full">
        <PageHeader title={ruleId ? 'Edit Alert Rule' : 'Create Alert Rule'} />
        <section className="flex flex-col w-full px-5 pt-5 pb-5">
          <div className="flex items-center justify-center py-8">
            <p>Loading alert rule...</p>
          </div>
        </section>
      </ScrollArea>
    );
  }

  // Show error only when rule ID exists but not found
  if (ruleId && !isFetching && !alertRule) {
    return (
      <ScrollArea className="w-full h-full">
        <PageHeader title={ruleId ? 'Edit Alert Rule' : 'Create Alert Rule'} />
        <section className="flex flex-col w-full px-5 pt-5 pb-5">
          <div className="flex items-center justify-center py-8">
            <p>Alert rule not found</p>
          </div>
        </section>
      </ScrollArea>
    );
  }

  const isLoading = isCreating || isUpdating;

  return (
    <ScrollArea className="w-full h-full">
      <PageHeader title={ruleId ? 'Edit Alert Rule' : 'Create Alert Rule'} />
      <section className="flex flex-col w-full px-5 pt-5 pb-5">
        <AlertRuleEditor
          initialType={alertType}
          initialData={alertRule?.rule}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          isLoading={isLoading}
        />
      </section>
    </ScrollArea>
  );
}
