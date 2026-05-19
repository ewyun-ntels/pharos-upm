
import React from 'react';
import { NotificationRuleEditor, useNotificationRule } from '@features/notification';
import {PageHeader} from '@components/page-header';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useCreate, useUpdate } from '@/lib/data-provider';
import { useToast } from '@hooks/use-toast';
import { NOTIFICATION_RESOURCES, NOTIFICATION_PROVIDER_NAME } from '@providers/notification-provider/types';

/**
 * Notification Rule Edit Page
 *
 * Create new or edit existing notification rule
 * - URL: /notification/edit (create new)
 * - URL: /notification/edit?id=xxx (edit existing)
 */
export default function NotificationRuleEdit() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { toast } = useToast();
  const ruleId = searchParams?.get('id');

  // Fetch existing notification rule (only when id exists)
  const {
    query: {
      data: notificationRuleResponse,
      isLoading: isFetching,
    },
  } = useNotificationRule(ruleId || undefined);

  const notificationRule = notificationRuleResponse?.data;

  const { mutate: createRule, mutation: createMutation } = useCreate();
  const { mutate: updateRule, mutation: updateMutation } = useUpdate();
  const isCreating = createMutation.isPending;
  const isUpdating = updateMutation.isPending;

  const handleCancel = () => {
    navigate('/notification');
  };

  const handleSubmit = (data: any) => {
    // Validation is handled by Zod schema in the form
    const payload = {
      resource: NOTIFICATION_RESOURCES.RULE,
      values: {
        notification_type: 'snmp',
        rule: data,
      },
      meta: {
        dataProviderName: NOTIFICATION_PROVIDER_NAME,
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
            toast({ description: 'Notification rule updated successfully' });
            navigate('/notification');
          },
          onError: () => {
            toast({ variant: 'destructive', description: 'Failed to update notification rule' });
          },
        },
      );
    } else {
      // Create new rule
      createRule(
        payload,
        {
          onSuccess: () => {
            toast({ description: 'Notification rule created successfully' });
            navigate('/notification');
          },
          onError: () => {
            toast({ variant: 'destructive', description: 'Failed to create notification rule' });
          },
        },
      );
    }
  };

  // Show loading only when fetching existing rule
  if (ruleId && isFetching) {
    return (
      <main className="w-full h-full overflow-y-auto">
        <PageHeader title={ruleId ? 'Edit Notification Rule' : 'Create Notification Rule'} />
        <section className="flex flex-col w-full px-5 pt-5 pb-5">
          <div className="flex items-center justify-center py-8">
            <p>Loading notification rule...</p>
          </div>
        </section>
      </main>
    );
  }

  // Show error only when rule ID exists but not found
  if (ruleId && !isFetching && !notificationRule) {
    return (
      <main className="w-full h-full overflow-y-auto">
        <PageHeader title={ruleId ? 'Edit Notification Rule' : 'Create Notification Rule'} />
        <section className="flex flex-col w-full px-5 pt-5 pb-5">
          <div className="flex items-center justify-center py-8">
            <p>Notification rule not found</p>
          </div>
        </section>
      </main>
    );
  }

  const isLoading = isCreating || isUpdating;

  // Determine notification type and rule data
  const notificationType = notificationRule?.notification_type || 'snmp';
  const rule = notificationRule?.rule as any;

  return (
    <main className="w-full h-full overflow-y-auto">
      <PageHeader title={ruleId ? 'Edit Notification Rule' : 'Create Notification Rule'} />
      <section className="flex flex-col w-full px-5 pt-5 pb-5">
        <NotificationRuleEditor
          initialType={notificationType as any}
          initialData={rule}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          isLoading={isLoading}
        />
      </section>
    </main>
  );
}
