import {useMemo, useCallback} from 'react';
import {useList, useDelete} from '@lib/data-provider';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid';
import {Button} from '@pharos/shared/components/ui';
import {toast} from '@pharos/shared/components';
import {useSheetStore} from '@components/sheet/sheetStore';
import {useHasPermission} from '@pharos/shared/features/auth';
import {CATV_PROVIDER_NAME, CATV_RESOURCES, Schedule} from '../../../../providers';
import {CATV_PERMISSIONS} from '../../../../permissions';
import {getScheduleListColumns} from './columns';
import {ScheduleDetailSheet} from './ScheduleDetailSheet';

interface ScheduleListProps {
  onCreateClick?: () => void;
  onScheduleEdit?: (schedule: Schedule) => void;
}

export function ScheduleList({onCreateClick, onScheduleEdit}: ScheduleListProps) {
  const {setSheet} = useSheetStore();

  const canUpdate = useHasPermission(CATV_PERMISSIONS.Update);
  const canDelete = useHasPermission(CATV_PERMISSIONS.Delete);

  const scheduleQuery = useList<Schedule>({
    resource: CATV_RESOURCES.SCHEDULES,
    dataProviderName: CATV_PROVIDER_NAME,
    pagination: {
      currentPage: 1,
      pageSize: 1000,
      mode: 'server' as const,
    },
    queryOptions: {
      refetchOnMount: true,
    },
  });

  const sortedData = useMemo(() => {
    const data = (scheduleQuery.query.data?.data as Schedule[]) || [];
    return [...data].sort((a, b) => {
      const ta = a.create_time ?? '';
      const tb = b.create_time ?? '';
      return tb.localeCompare(ta);
    });
  }, [scheduleQuery.query.data]);

  const {mutate: deleteSchedule} = useDelete();

  const handleDelete = (name: string) => {
    deleteSchedule(
      {
        resource: CATV_RESOURCES.SCHEDULES,
        id: name,
        dataProviderName: CATV_PROVIDER_NAME,
      },
      {
        onSuccess: () => {
          toast.success('스케줄이 삭제되었습니다.');
          scheduleQuery.query.refetch();
        },
        onError: (error: any) => {
          const message =
            error?.response?.data?.message ||
            error?.message ||
            (typeof error === 'string' ? error : '알 수 없는 오류');
          toast.error(`스케줄 삭제 실패: ${message}`);
        },
      },
    );
  };

  const columns = getScheduleListColumns(
    (schedule) => {
      if (onScheduleEdit) {
        onScheduleEdit(schedule);
      }
    },
    handleDelete,
    {canUpdate, canDelete},
  );

  const handleRowClick = useCallback(
    (schedule: Schedule) => {
      setSheet(<ScheduleDetailSheet schedule={schedule} />);
    },
    [setSheet],
  );

  return (
    <div className="h-full flex flex-col p-4">
      <DataGrid
        data={sortedData}
        columns={columns}
        rowCursor={true}
        getRowId={(row: Schedule) => row.name}
        onRowClick={handleRowClick}
        isLoading={scheduleQuery.query.isLoading}
        rightFilters={() => {
          const items = [];
          if (onCreateClick) {
            items.push(
              <Button key="create" onClick={onCreateClick}>
                스케줄 등록
              </Button>,
            );
          }
          return items;
        }}
      />
    </div>
  );
}
