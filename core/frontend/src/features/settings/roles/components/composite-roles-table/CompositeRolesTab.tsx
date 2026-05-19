import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useList, useOne, useCreate, HttpError } from '@/lib/data-provider';
import { useNavigate } from 'react-router-dom';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import { ROLE_PROVIDER_NAME, ROLE_RESOURCES } from '@providers/role-provider';
import { useToast } from '@hooks/use-toast';
import { CompositeRoleActions } from './CompositeRoleActions';
import { getCompositeRoleColumns } from './columns';
import type { RoleMetadata, RoleGroup } from '../../types';

export function CompositeRolesTab() {
  const { toast } = useToast();
  const navigate = useNavigate();

  const {
    query: { data: baseData, isLoading: isLoadingBase },
  } = useList<RoleGroup>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.ALL_ROLES,
    queryOptions: { retry: false, staleTime: 0 },
    meta: { includeHidden: true },
  });

  const {
    query: { data: overlayData, isLoading: isLoadingOverlay, refetch: refetchOverlay },
  } = useOne<RoleMetadata[]>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.CONFIG,
    id: 'config',
    queryOptions: { retry: false, staleTime: 0 },
  });

  const { mutateAsync: saveConfig } = useCreate<RoleMetadata[], HttpError, { config: RoleMetadata[] }>();

  const [isSaving, setIsSaving] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [compositeRoles, setCompositeRoles] = useState<RoleMetadata[]>([]);

  const baseGroups: RoleGroup[] = useMemo(
    () => (baseData?.data as unknown as RoleGroup[]) ?? [],
    [baseData],
  );

  const allAtomicRoles: RoleMetadata[] = useMemo(
    () => baseGroups.flatMap((g) => g.roles),
    [baseGroups],
  );

  const currentOverlay: RoleMetadata[] = useMemo(() => {
    const d = overlayData?.data as unknown;
    return Array.isArray(d) ? (d as RoleMetadata[]) : [];
  }, [overlayData]);

  const atomicOverlay = useMemo(
    () => currentOverlay.filter((r) => !r.roles || Object.keys(r.roles).length === 0),
    [currentOverlay],
  );

  useEffect(() => {
    if (isLoadingOverlay || isLoadingBase) return;
    const composites = currentOverlay
      .filter((r) => r.roles && Object.keys(r.roles).length > 0)
      .sort((a, b) => (a.group ?? '').localeCompare(b.group ?? ''));
    setCompositeRoles(composites);
  }, [currentOverlay, isLoadingOverlay, isLoadingBase]);

  // ── 서버 저장 ──────────────────────────────────────────
  const saveToServer = async (composites: RoleMetadata[]): Promise<boolean> => {
    const configToSave: RoleMetadata[] = [...atomicOverlay, ...composites];
    setIsSaving(true);
    try {
      await saveConfig({
        resource: ROLE_RESOURCES.CONFIG,
        values: { config: configToSave },
        dataProviderName: ROLE_PROVIDER_NAME,
      });
      toast({ description: 'Saved successfully.' });
      refetchOverlay();
      return true;
    } catch {
      toast({ description: 'Failed to save.', variant: 'destructive' });
      return false;
    } finally {
      setIsSaving(false);
    }
  };

  // ── 핸들러 ──────────────────────────────────────────────
  const handleAdd = useCallback(() => {
    navigate('/settings/roles/composite');
  }, [navigate]);

  const handleEdit = (role: RoleMetadata) => {
    navigate('/settings/roles/composite', { state: { editTarget: role } });
  };

  const handleDelete = async (key: string) => {
    const newComposites = compositeRoles.filter((r) => r.key !== key);
    setCompositeRoles(newComposites);
    setDeleteTarget(null);
    await saveToServer(newComposites);
  };

  const isLoading = isLoadingBase || isLoadingOverlay;

  // DataGrid columns
  const columns = useMemo(
    () =>
      getCompositeRoleColumns({
        allAtomicRoles,
        isSaving,
        onEdit: handleEdit,
        onDelete: setDeleteTarget,
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [allAtomicRoles, isSaving],
  );

  const actionsItem = useMemo(
    () => <CompositeRoleActions isLoading={isLoading} isSaving={isSaving} onAdd={handleAdd} />,
    [isLoading, isSaving, handleAdd],
  );

  return (
    <>
      <div className="px-5">
        <DataGrid<RoleMetadata>
          tableKey="composite-roles"
          data={compositeRoles}
          columns={columns}
          getRowId={(row) => row.key}
          searchableColumns={['displayName', 'key', 'description']}
          description="Create composite roles by bundling multiple base roles into a single assignable role. Base roles are automatically expanded at login time."
          rightFilters={() => [actionsItem]}
          rowCursor={false}
          isLoading={isLoading}
          useTableSorting={false}
          enablePagination={false}
          emptyCustomMessage="No composite roles defined. Click 'Add composite role' to create a composite role by bundling base roles."
        />
      </div>

      <AlertDialog open={!!deleteTarget} onOpenChange={(v) => !v && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Composite Role?</AlertDialogTitle>
            <AlertDialogDescription>
              <code className="font-mono">{deleteTarget}</code> will be removed. Users assigned
              this role will lose associated permissions on next login.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteTarget && handleDelete(deleteTarget)}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
