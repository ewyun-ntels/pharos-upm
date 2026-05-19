import React, { useEffect, useState, useMemo, useCallback } from 'react';
import { SelectBox } from '@pharos/shared/components/ui-extension/select/box';
import { Title } from '@pharos/shared/components/ui-extension';
import {
  Button,
  Checkbox,
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import { useToast } from '@hooks/use-toast';
import { useDelete, useList, useUpdate } from '@/lib/data-provider';
import { Plus } from '@pharos/shared/components';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import { cn } from '@pharos/shared/lib';
import { AddPermissionForm } from './AddPermissionForm';
import { getPermissionColumns, PermissionData } from './columns';

interface PermissionTabProps {
  dashboardId: string;
}

const PERMISSION_OPTIONS = [
  { value: ActionSchema.enum.viewer, label: 'Viewer' },
  { value: ActionSchema.enum.editor, label: 'Editor' },
  { value: ActionSchema.enum.owner, label: 'Owner' },
];

export function PermissionTab({ dashboardId }: PermissionTabProps) {
  const { toast } = useToast();
  const [publicChecked, setPublicChecked] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [showDeleteAlert, setShowDeleteAlert] = useState(false);
  const [deletingSubject, setDeletingSubject] = useState<string>('');

  const {
    query: { data, refetch },
  } = useList({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.DASHBOARD_PERMISSION.replace('{id}', dashboardId),
    queryOptions: {
      retry: false,
      enabled: !!dashboardId,
    },
  });

  const { mutate: addMutate, mutation: addMutation } = useUpdate({
    dataProviderName: USER_PROVIDER_NAME,
    resource: DASHBOARD_RESOURCES.DASHBOARD,
    id: dashboardId + '/permission',
    meta: {
      method: 'put',
    },
  });

  const { mutate: DeleteMutate, mutation: deleteMutation } = useDelete({
    dataProviderName: USER_PROVIDER_NAME,
  });

  const permissionList = useMemo<PermissionData[]>(() => {
    if (!data?.data) return [];
    return (data.data as PermissionData[])
      .filter((item) => item.subject !== '__public')
      .sort((a, b) => a.subject.localeCompare(b.subject));
  }, [data]);

  const publicPermission = useMemo(() => {
    if (!data?.data) return ActionSchema.enum.viewer;
    const publicData = (data.data as PermissionData[]).find((item) => item.subject === '__public');
    return publicData?.action ?? ActionSchema.enum.viewer;
  }, [data]);

  useEffect(() => {
    if (!data?.data) return;
    const hasPublic = (data.data as PermissionData[]).some((item) => item.subject === '__public');
    setPublicChecked(hasPublic);
  }, [data]);


  const callPublicUpdate = useCallback(({
    checked = publicChecked,
    action = publicPermission,
  }: {
    checked?: boolean;
    action?: string;
  }) => {
    if (!checked) return;
    const validAction = action || ActionSchema.enum.viewer;
    addMutate(
      {
        values: {
          subject: '__public',
          object: dashboardId,
          action: validAction,
        },
      },
      {
        onSuccess: () => {
          toast({ description: 'Successfully saved.' });
          refetch();
        },
        onError: (error) => {
          toast({ description: 'Error saving permission' });
          console.error('Error saving permission:', error);
          refetch();
        },
      },
    );
  }, [publicChecked, publicPermission, dashboardId, addMutate, toast, refetch]);

  const handleAddPermission = useCallback((formData: PermissionData) => {
    addMutate(
      {
        values: {
          ...formData,
          object: dashboardId,
        },
      },
      {
        onSuccess: () => {
          toast({ description: 'Successfully saved.' });
          refetch();
        },
        onError: () => {
          toast({ description: 'Error saving permission' });
        },
      },
    );
  }, [addMutate, dashboardId, toast, refetch]);

  const handleDeletePublic = useCallback(() => {
    if (!dashboardId) return;
    const encodedURI = encodeURI(dashboardId + `/permission/__public`);
    DeleteMutate(
      {
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: encodedURI,
      },
      {
        onSuccess: () => {
          toast({ description: 'Public access disabled' });
          refetch();
        },
        onError: () => {
          toast({ description: `Error disabling public access` });
          refetch();
        },
      },
    );
  }, [dashboardId, DeleteMutate, toast]);

  const handleDelete = useCallback((subject: string) => {
    setDeletingSubject(subject);
    setShowDeleteAlert(true);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!dashboardId || !deletingSubject) return;
    const encodedURI = encodeURI(dashboardId + `/permission/${deletingSubject}`);
    DeleteMutate(
      {
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: encodedURI,
      },
      {
        onSuccess: () => {
          toast({ description: 'Permission deleted successfully' });
          setShowDeleteAlert(false);
          setDeletingSubject('');
          refetch();
        },
        onError: () => {
          toast({ description: `Error deleting permission` });
          setShowDeleteAlert(false);
          setDeletingSubject('');
        },
      },
    );
  }, [dashboardId, deletingSubject, DeleteMutate, toast, refetch]);

  const handleActionChange = useCallback((subject: string, action: string) => {
    addMutate(
      {
        values: { subject, object: dashboardId, action },
      },
      {
        onSuccess: () => {
          toast({ description: 'Successfully saved.' });
          refetch();
        },
        onError: () => {
          toast({ description: 'Error saving permission' });
          refetch();
        },
      },
    );
  }, [addMutate, dashboardId, toast, refetch]);

  const columns = useMemo(
    () => getPermissionColumns({
      dashboardId,
      isDeletePending: deleteMutation.isPending,
      onActionChange: handleActionChange,
      onDelete: handleDelete,
    }),
    [dashboardId, deleteMutation.isPending, handleActionChange, handleDelete],
  );

  return (
    <div className="px-5 py-3">
      <div className="max-w-7xl space-y-8">
        <div>
          <Title variant="h3">Current permissions</Title>
          <DataGrid<PermissionData>
            tableKey="dashboard-permissions"
            data={permissionList}
            columns={columns}
            enablePagination={false}
            useTableSorting={false}
            tableHeight={permissionList.length === 0 ? 140 : Math.min(32 + permissionList.length * 41, 319)}
            className='mb-4'
          />
          <div className="mb-3">
            <Button
              variant="default"
              size="default"
              onClick={() => setShowAddForm(!showAddForm)}
            >
              <Plus className="w-4 h-4" />
              Add permission
            </Button>
          </div>

          {showAddForm && (
            <AddPermissionForm
              permissionList={permissionList}
              isPending={addMutation.isPending}
              onAdd={handleAddPermission}
              onCancel={() => setShowAddForm(false)}
            />
          )}
          
        </div>
        <div>
          <Title variant="h3">Public access</Title>
          <div className="space-y-3">
            <div className="flex-col">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="public-access"
                  checked={publicChecked}
                  onCheckedChange={(checked) => {
                    const isChecked = checked === true;
                    setPublicChecked(isChecked);
                    if (isChecked) callPublicUpdate({ checked: isChecked, action: publicPermission });
                    else handleDeletePublic();
                  }}
                />
                <label htmlFor="public-access" className="text-sm font-medium cursor-pointer">
                  Public
                </label>
              </div>
              <p className="pl-6 text-sm text-muted-foreground">
                Everyone can access this dashboard with the selected role.
              </p>
            </div>
            <div className="ml-6 space-y-2">
              <div className={cn("w-[200px]", !publicChecked && "opacity-50 pointer-events-none")}>
                <SelectBox
                  size="full"
                  options={PERMISSION_OPTIONS}
                  onChange={(e) => {
                    if (!publicChecked) return;
                    callPublicUpdate({ action: e });
                  }}
                  value={publicPermission}
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <AlertDialog open={showDeleteAlert} onOpenChange={setShowDeleteAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure you want to delete?</AlertDialogTitle>
            <AlertDialogDescription>
              <strong>{deletingSubject}</strong> will be deleted. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDeleteConfirm}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
