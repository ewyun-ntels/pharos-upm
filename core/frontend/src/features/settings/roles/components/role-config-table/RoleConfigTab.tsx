import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import { useList, useOne, useCreate, HttpError } from '@/lib/data-provider';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { ROLE_PROVIDER_NAME, ROLE_RESOURCES } from '@providers/role-provider';
import { useToast } from '@hooks/use-toast';
import { getRoleConfigColumns, type TableRow } from './columns';
import { RoleConfigActions } from './RoleConfigActions';
import type { RoleMetadata, RoleGroup, AtomicEditState } from '../../types';

export function RoleConfigTab() {
  const { toast } = useToast();

  const {
    query: { data: baseData, isLoading: isLoadingBase },
  } = useList<RoleGroup>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.BASE_ROLES,
    queryOptions: { retry: false, staleTime: 0 },
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
  const [isDirty, setIsDirty] = useState(false);
  const [atomicEdits, setAtomicEdits] = useState<Record<string, AtomicEditState>>({});
  const shouldResetRef = useRef(true);

  // raw base: overlay 적용 전 순수 base (비교 기준)
  const baseGroups: RoleGroup[] = useMemo(
    () => (baseData?.data as unknown as RoleGroup[]) ?? [],
    [baseData],
  );

  // raw base를 key → RoleMetadata 맵으로 변환 (저장 시 비교 기준)
  const rawBaseMap = useMemo(() => {
    const m: Record<string, RoleMetadata> = {};
    baseGroups.forEach((g) => g.roles.forEach((r) => { m[r.key] = r; }));
    return m;
  }, [baseGroups]);

  const currentOverlay: RoleMetadata[] = useMemo(() => {
    const d = overlayData?.data as unknown;
    return Array.isArray(d) ? (d as RoleMetadata[]) : [];
  }, [overlayData]);

  const compositeOverlay = useMemo(
    () => currentOverlay.filter((r) => r.roles && Object.keys(r.roles).length > 0),
    [currentOverlay],
  );

  // tableData용: raw base + overlay 적용 결과 (display name 등 overlay 값 반영)
  const mergedGroups: RoleGroup[] = useMemo(() => {
    const overlayMap: Record<string, RoleMetadata> = {};
    currentOverlay.forEach((r) => { overlayMap[r.key] = r; });
    return baseGroups.map((g) => ({
      ...g,
      roles: g.roles.map((r) => (overlayMap[r.key] ? { ...r, ...overlayMap[r.key] } : r)),
    }));
  }, [baseGroups, currentOverlay]);

  useEffect(() => {
    if (isLoadingOverlay || isLoadingBase) return;
    if (!shouldResetRef.current) return;

    const atomicOverlayMap: Record<string, RoleMetadata> = {};
    currentOverlay.forEach((r) => {
      if (!r.roles || Object.keys(r.roles).length === 0) atomicOverlayMap[r.key] = r;
    });

    const initial: Record<string, AtomicEditState> = {};
    baseGroups.forEach((group) => {
      group.roles.forEach((role) => {
        if (role.roles && Object.keys(role.roles).length > 0) return;
        const ov = atomicOverlayMap[role.key];
        initial[role.key] = {
          displayName: ov?.displayName ?? role.displayName,
          description: ov?.description ?? role.description ?? '',
          hide: ov?.hide ?? role.hide ?? false,
        };
      });
    });

    setAtomicEdits(initial);
    setIsDirty(false);
    shouldResetRef.current = false;
  }, [currentOverlay, baseGroups, isLoadingOverlay, isLoadingBase]);

  // 편집 상태를 row 데이터에 포함 → data prop 변경으로 셀 리렌더링 (columns 재생성 없음)
  // group → key 순으로 정렬하여 저장/새로고침 시 행 순서가 변하지 않도록 함
  const tableData: TableRow[] = useMemo(
    () =>
      mergedGroups
        .flatMap((g) =>
          g.roles
            .filter((r) => !r.roles || Object.keys(r.roles).length === 0)
            .map((r) => ({
              ...r,
              editState: atomicEdits[r.key] ?? {
                displayName: r.displayName,
                description: r.description ?? '',
                hide: r.hide ?? false,
              },
            })),
        )
        .sort((a, b) => {
          const groupCmp = (a.group ?? '').localeCompare(b.group ?? '');
          return groupCmp !== 0 ? groupCmp : a.key.localeCompare(b.key);
        }),
    [mergedGroups, atomicEdits],
  );

  const handleAtomicEdit = useCallback(
    (key: string, field: keyof AtomicEditState, value: string | boolean) => {
      setAtomicEdits((prev) => ({ ...prev, [key]: { ...(prev[key] ?? {}), [field]: value } }));
      setIsDirty(true);
    },
    [],
  );

  const handleSave = useCallback(async () => {
    // raw base와 비교해 변경된 항목만 overlay로 저장
    const configToSave: RoleMetadata[] = [];
    baseGroups.forEach((group) => {
      group.roles.forEach((role) => {
        if (role.roles && Object.keys(role.roles).length > 0) return;
        const edit = atomicEdits[role.key];
        if (!edit) return;
        const base = rawBaseMap[role.key];
        const changed =
          edit.displayName !== (base?.displayName ?? role.displayName) ||
          edit.description !== (base?.description ?? role.description ?? '') ||
          edit.hide !== (base?.hide ?? role.hide ?? false);
        if (changed) {
          configToSave.push({
            ...role,
            displayName: edit.displayName,
            description: edit.description,
            hide: edit.hide,
          });
        }
      });
    });
    compositeOverlay.forEach((cr) => configToSave.push(cr));

    if (configToSave.length === 0) {
      toast({ description: 'No changes to save.' });
      setIsDirty(false);
      return;
    }

    setIsSaving(true);
    try {
      await saveConfig({
        resource: ROLE_RESOURCES.CONFIG,
        values: { config: configToSave },
        dataProviderName: ROLE_PROVIDER_NAME,
      });
      toast({ description: 'Configuration saved successfully.' });
      shouldResetRef.current = true;
      refetchOverlay();
    } catch (err: unknown) {
      const detail =
        (err as { response?: { data?: { details?: string; error?: string } } })?.response?.data?.details ??
        (err as { response?: { data?: { details?: string; error?: string } } })?.response?.data?.error ??
        (err as { message?: string })?.message ??
        'Unknown error';
      toast({ description: `Failed to save configuration: ${detail}`, variant: 'destructive' });
    } finally {
      setIsSaving(false);
    }
  }, [baseGroups, rawBaseMap, atomicEdits, compositeOverlay, saveConfig, toast, refetchOverlay]);


  // columns는 handleAtomicEdit만 의존 → 타이핑 시 재생성되지 않아 포커스 유지
  const columns = useMemo(
    () => getRoleConfigColumns({ handleAtomicEdit }),
    [handleAtomicEdit],
  );

  const isLoading = isLoadingBase || isLoadingOverlay;

  const actionsItem = useMemo(
    () => (
      <RoleConfigActions
        isDirty={isDirty}
        isSaving={isSaving}
        isLoading={isLoading}
        onSave={handleSave}
      />
    ),
    [isDirty, isSaving, isLoading, handleSave],
  );

  return (
    <>
      <div className="px-5">
        <DataGrid<TableRow>
          tableKey="role-config"
          data={tableData}
          columns={columns}
          getRowId={(row) => row.key}
          searchableColumns={['key', 'displayName', 'description']}
          description="Configure display names, descriptions, and visibility of base roles. Base roles are atomic permissions that can be assigned directly or combined into composite roles."
          rightFilters={() => [actionsItem]}
          rowCursor={false}
          isLoading={isLoading}
          useTableSorting={false}
          enablePagination={false}
          excludeRowClickColumns={['displayName', 'description', 'visible']}
        />
      </div>
    </>
  );
}
