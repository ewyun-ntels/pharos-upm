import React from 'react';
import {useUpdate, useInvalidate, useDelete} from '@/lib/data-provider';
import {SquareCheckBig, Square, Trash2} from '@pharos/shared/components';
import {useToast} from '@hooks/use-toast';
import {ColumnConfig} from '@pharos/shared/hooks/table-columns';
import {ToggleIconButton} from '@pharos/shared/components/ui-extension';
import AlertBadge from '@components/badge/alert-badge';
import LabelBadge from '@components/badge/label-badge';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {ALERT_RESOURCES} from '@providers/alert-provider/types';

interface AlertRowData {
  alert_id: string;
  name: string;
  mask: boolean;
  severity: string;
  labels?: Record<string, any>;
  timestamp: string;
  description?: string;
}

export const useAlertTableColumns = () => {
  const {toast} = useToast();
  const {mutate: updateMutate} = useUpdate();
  const {mutate: deleteMutate} = useDelete();
  const invalidate = useInvalidate();

  const updateMask = (data: AlertRowData) => {
    const {alert_id, name, mask} = data;

    updateMutate(
      {
        resource: ALERT_RESOURCES.STATUS_MASK,
        id: `${name}/${alert_id}`,
        values: {
          mask: !mask,
        },
      },
      {
        onSuccess: () => {
          invalidate({
            resource: ALERT_RESOURCES.STATUS,
            invalidates: ['list'],
          });
          toast({description: 'Alert state updated.'});
        },
        onError: () => {
          toast({description: 'Failed to update alert state.'});
        },
      },
    );
  };

  const deleteAlert = (data: AlertRowData) => {
    const {alert_id, name} = data;

    deleteMutate(
      {
        resource: ALERT_RESOURCES.STATUS,
        id: `${name}/${alert_id}`,
      },
      {
        onSuccess: () => {
          invalidate({
            resource: ALERT_RESOURCES.STATUS,
            invalidates: ['list'],
          });
          toast({description: 'Alert deleted successfully.'});
        },
        onError: () => {
          toast({description: 'Error deleting alert.'});
        },
      },
    );
  };

  const columns: ColumnConfig[] = [
    {
      key: 'timestamp',
      properties: {
        title: 'Time',
        unit: 'local_time',
        sortingFn: 'datetime',
        size: 180,
        defaultSort: 'desc',
      },
    },
    {
      key: 'description',
      properties: {
        title: 'Name',
        size: 300,
        sortingFn: 'text',
      },
      customCell: (_value: string, data: AlertRowData) => {
        const Name = data?.labels?.name;
        return <div className="w-full flex justify-start items-center">{Name || '-'}</div>;
      },
    },
    {
      key: 'severity',
      properties: {
        title: 'Severity',
        sortingFn: 'text',
        size: 60,
        searchByFormatted: false,
      },
      customCell: (severity: string) => {
        return (
          <div className="w-full flex justify-center items-center">
            <AlertBadge id={severity} />
          </div>
        );
      },
    },
    {
      key: 'labels',
      properties: {
        title: 'Information',
        size: 450,
        sortingFn: 'text',
        enableSorting: false,
        searchByFormatted: false,
      },
      customCell: (labels: Record<string, any>) => {
        return <LabelBadge labels={labels} />;
      },
    },
    {
      key: 'mask',
      properties: {
        title: 'Actions',
        size: 60,
        enableSorting: false,
      },
      customCell: (mask: boolean, row: AlertRowData) => {
        return (
          <div className="w-full flex justify-center items-center space-x-1">
            <ToggleIconButton
              icon={mask ? <SquareCheckBig /> : <Square />}
              size="icon-xs"
              toggled={mask}
              tooltip="Acknowledge"
              onClick={(e) => {
                e.stopPropagation();
                updateMask(row);
              }}
            />
            <IconButton
              icon={<Trash2 />}
              size="icon-xs"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation();
                deleteAlert(row);
              }}
            >
              Clear
            </IconButton>
          </div>
        );
      },
    },
  ];

  return {columns, updateMask, deleteAlert};
};
