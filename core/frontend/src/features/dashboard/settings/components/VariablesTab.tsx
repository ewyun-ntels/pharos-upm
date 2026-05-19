import React, { useState, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus } from '@pharos/shared/components';
import { Button } from '@pharos/shared/components/ui';
import VariablesTable, { ToggleTableRow } from '@/features/dashboard/variables/components/VariablesTable/VariablesTable';
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
import { useDashboardData, useDashboardActions } from '@features/dashboard/hooks/use-dashboard-store';
import { FilterConfig } from '@features/dashboard/components/VariableBar/types';

interface VariablesTabProps {
  dashboardId: string;
}

export function VariablesTab({ dashboardId }: VariablesTabProps) {
  const navigate = useNavigate();
  const { filters } = useDashboardData();
  const { setFilters } = useDashboardActions();

  const [showDeleteAlert, setShowDeleteAlert] = useState(false);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  const convertVariablesForTable = useCallback((variables: FilterConfig[] | undefined, position: string) => {
    if (!variables) return [];

    return variables.map((variable) => ({
      ...variable,
      id: variable.id,
      type: variable.type,
      description: variable.options?.description || '',
      position,
    }));
  }, []);

  const handleRowsChange = useCallback((position: string, newRows: ToggleTableRow[]) => {
    const reorderedFilters = newRows.reduce<FilterConfig[]>((acc, row) => {
      const match = filters.find(f => f.id === row.id);
      if (match) {
        acc.push(match);
      }
      return acc;
    }, []);

    const otherPositionFilters = filters.filter(f => f.kind !== position);

    setFilters([...otherPositionFilters, ...reorderedFilters]);
  }, [filters, setFilters]);

  const handleDeleteRow = useCallback((rowId: string) => {
    setDeleteTargetId(rowId);
    setShowDeleteAlert(true);
  }, []);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteTargetId) return;

    setFilters(filters.filter(f => f.id !== deleteTargetId));
    setDeleteTargetId(null);
    setShowDeleteAlert(false);
  }, [deleteTargetId, filters, setFilters]);

  const handleNewVariable = useCallback(() => {
    const params = new URLSearchParams(window.location.search);
    navigate(`/dashboards/${dashboardId}/variables/new?${params.toString()}`);
  }, [dashboardId, navigate]);

  const headerLRows = useMemo(() => 
    convertVariablesForTable(filters?.filter(f => f.kind === 'headerL'), 'headerL'),
    [filters, convertVariablesForTable]
  );
  
  const headerRRows = useMemo(() => 
    convertVariablesForTable(filters?.filter(f => f.kind === 'headerR'), 'headerR'),
    [filters, convertVariablesForTable]
  );
  


  return (
    <div className="px-5 py-3">
      <div className="max-w-7xl space-y-4">
        <div className="flex flex-row justify-end mb-1">
          <Button
            variant="default"
            size="default"
            onClick={handleNewVariable}
          >
            <Plus/>
            New variable
          </Button>
        </div>

        <div className="flex flex-col gap-6">
          <VariablesTable
            title="Top left variables"
            rows={headerLRows}
            onRowsChange={(newRows) => handleRowsChange('headerL', newRows)}
            onDeleteRow={handleDeleteRow}
            dashboardId={dashboardId}
          />

          <VariablesTable
            title="Top right variables"
            rows={headerRRows}
            onRowsChange={(newRows) => handleRowsChange('headerR', newRows)}
            onDeleteRow={handleDeleteRow}
            dashboardId={dashboardId}
          />

        </div>
      </div>

      <AlertDialog open={showDeleteAlert} onOpenChange={setShowDeleteAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Variable</AlertDialogTitle>
            <AlertDialogDescription>Are you sure you want to delete this variable?</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleConfirmDelete}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
