import {useMemo, useState, useEffect, useCallback} from 'react';
import {CalendarIcon} from 'lucide-react';
import {useList} from '@lib/data-provider';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid';
import {VariableSelect, SearchInput, TimePicker} from '@pharos/shared/components/ui-extension';
import {
  Button,
  Calendar,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@pharos/shared/components/ui';
import {CATV_PROVIDER_NAME, CATV_RESOURCES} from '../../../../providers';
import type {ScheduleResultRecord, Schedule} from '../../../../providers';
import {getScheduleResultColumns, ScheduleResultRow} from './columns';
import {toScheduleResultRow} from './utils';
import {STATUS_RUNNING, STATUS_PENDING} from '../constants';
import {WORK_TYPE_OPTIONS, POLL_INTERVAL_MS, SEARCH_DEBOUNCE_MS} from '../constants';

type DateRange = {from: Date | undefined; to?: Date | undefined};

function formatDate(date: Date): string {
  const datePart = date.toLocaleDateString('sv'); // YYYY-MM-DD
  const h = date.getHours();
  const m = date.getMinutes();
  if (h === 0 && m === 0) return datePart;
  return `${datePart} ${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

function DateRangePickerContent({
  value,
  onChange,
  onClose,
}: {
  value: DateRange | undefined;
  onChange: (range: DateRange | undefined) => void;
  onClose?: () => void;
}) {
  const [pendingFrom, setPendingFrom] = useState<Date | undefined>(value?.from);
  const [pendingTo, setPendingTo] = useState<Date | undefined>(value?.to);
  const [fromTime, setFromTime] = useState(() => {
    const d = value?.from;
    if (!d) return '00:00';
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  });
  const [toTime, setToTime] = useState(() => {
    const d = value?.to;
    if (!d) return '23:59';
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  });

  // 같은 날이고 종료 시간이 시작 시간 이전인 경우 유효하지 않음
  const isTimeRangeInvalid =
    !!pendingFrom &&
    !!pendingTo &&
    pendingFrom.getFullYear() === pendingTo.getFullYear() &&
    pendingFrom.getMonth() === pendingTo.getMonth() &&
    pendingFrom.getDate() === pendingTo.getDate() &&
    pendingTo.getTime() <= pendingFrom.getTime();

  // Popover 재오픈 시 외부 value와 동기화
  useEffect(() => {
    setPendingFrom(value?.from);
    setPendingTo(value?.to);
    setFromTime(
      value?.from
        ? `${String(value.from.getHours()).padStart(2, '0')}:${String(value.from.getMinutes()).padStart(2, '0')}`
        : '00:00',
    );
    setToTime(
      value?.to
        ? `${String(value.to.getHours()).padStart(2, '0')}:${String(value.to.getMinutes()).padStart(2, '0')}`
        : '23:59',
    );
  }, [value]);

  return (
    <div className="flex flex-col">
      <div className="flex">
        <div className="border-r">
          <div className="pt-2 pb-1 text-xs font-medium text-muted-foreground text-center">
            시작 날짜
          </div>
          <Calendar
            mode="single"
            selected={pendingFrom}
            onSelect={(date) => {
              if (date) {
                const [h, m] = fromTime.split(':').map(Number);
                date.setHours(h, m, 0, 0);
                setPendingFrom(date);
              }
            }}
            disabled={(date) => date > new Date()}
          />
          <div className="border-t px-3 py-2">
            <TimePicker
              value={pendingFrom ?? new Date()}
              onChange={(date) => {
                const h = date.getHours();
                const m = date.getMinutes();
                setFromTime(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
                if (pendingFrom) {
                  const d = new Date(pendingFrom);
                  d.setHours(h, m, 0, 0);
                  setPendingFrom(d);
                }
              }}
              timePicker={{hour: true, minute: true}}
            />
          </div>
        </div>
        <div>
          <div className="pt-2 pb-1 text-xs font-medium text-muted-foreground text-center">
            종료 날짜
          </div>
          <Calendar
            mode="single"
            selected={pendingTo}
            onSelect={(date) => {
              if (date) {
                const [h, m] = toTime.split(':').map(Number);
                date.setHours(h, m, 0, 0);
                setPendingTo(date);
              }
            }}
            disabled={(date) => {
              if (date > new Date()) return true;
              if (pendingFrom) {
                const fromDay = new Date(
                  pendingFrom.getFullYear(),
                  pendingFrom.getMonth(),
                  pendingFrom.getDate(),
                );
                return date < fromDay;
              }
              return false;
            }}
          />
          <div className="border-t px-3 py-2">
            <TimePicker
              value={pendingTo ?? new Date()}
              onChange={(date) => {
                const h = date.getHours();
                const m = date.getMinutes();
                setToTime(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
                if (pendingTo) {
                  const d = new Date(pendingTo);
                  d.setHours(h, m, 0, 0);
                  setPendingTo(d);
                }
              }}
              timePicker={{hour: true, minute: true}}
            />
          </div>
        </div>
      </div>
      <div className="flex justify-end gap-2 p-2 border-t">
        {isTimeRangeInvalid && (
          <span className="text-xs text-destructive flex items-center mr-auto">
            종료 시간이 시작 시간보다 앞설 수 없습니다
          </span>
        )}
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            setPendingFrom(undefined);
            setPendingTo(undefined);
            setFromTime('00:00');
            setToTime('23:59');
            onChange(undefined);
            onClose?.();
          }}
        >
          초기화
        </Button>
        <Button
          size="sm"
          disabled={isTimeRangeInvalid}
          onClick={() => {
            onChange({from: pendingFrom, to: pendingTo});
            onClose?.();
          }}
        >
          적용
        </Button>
      </div>
    </div>
  );
}

interface ScheduleResultProps {
  onRowClick: (
    resultId: string,
    workType: string,
    scheduleName?: string,
    requestTime?: string,
  ) => void;
}

export function ScheduleResult({onRowClick}: ScheduleResultProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(100);
  const [workTypeFilter, setWorkTypeFilter] = useState('');
  const [scheduleNameInput, setScheduleNameInput] = useState('');
  const [scheduleNameFilter, setScheduleNameFilter] = useState('');
  const [dateRange, setDateRange] = useState<DateRange | undefined>();
  const [dateRangeOpen, setDateRangeOpen] = useState(false);

  // 스케줄 이름 검색어 디바운스 (IME 조합 완료 후 API 호출)
  useEffect(() => {
    const timer = setTimeout(() => {
      setScheduleNameFilter(scheduleNameInput);
      setCurrentPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [scheduleNameInput]);

  // 필터 → API filters 빌드
  const filters = useMemo(() => {
    const result: {field: string; operator: 'eq'; value: string}[] = [];
    if (workTypeFilter) result.push({field: 'work_type', operator: 'eq', value: workTypeFilter});
    if (scheduleNameFilter)
      result.push({field: 'schedule_name', operator: 'eq', value: scheduleNameFilter});
    if (dateRange?.from) {
      result.push({field: 'from', operator: 'eq', value: dateRange.from.toISOString()});
    }
    if (dateRange?.to) {
      result.push({field: 'to', operator: 'eq', value: dateRange.to.toISOString()});
    } else if (dateRange?.from) {
      const to = new Date(dateRange.from);
      to.setHours(23, 59, 59, 999);
      result.push({field: 'to', operator: 'eq', value: to.toISOString()});
    }
    return result;
  }, [workTypeFilter, scheduleNameFilter, dateRange]);

  const listQuery = useList<ScheduleResultRecord>({
    resource: CATV_RESOURCES.SCHEDULE_RESULTS,
    dataProviderName: CATV_PROVIDER_NAME,
    filters,
    pagination: {
      currentPage,
      pageSize,
      mode: 'server' as const,
    },
    queryOptions: {
      refetchOnMount: true,
    },
  });

  // 스케줄 이름 필터 옵션 제공용
  const scheduleQuery = useList<Schedule>({
    resource: CATV_RESOURCES.SCHEDULES,
    dataProviderName: CATV_PROVIDER_NAME,
    pagination: {
      currentPage: 1,
      pageSize: 1000,
      mode: 'server' as const,
    },
  });

  // 스케줄 이름 옵션
  useMemo(() => {
    const schedules = (scheduleQuery.query.data?.data as Schedule[]) || [];
    return [...schedules]
      .sort((a, b) => a.name.localeCompare(b.name))
      .map((s) => ({label: s.name, value: s.name}));
  }, [scheduleQuery.query.data]);
  const isLoading = listQuery.query.isLoading;
  const isFetching = listQuery.query.isFetching;
  const data = listQuery.query.data;
  const refetch = listQuery.query.refetch;

  // ScheduleResult API 응답 → 테이블 행 형식으로 변환
  const displayData: ScheduleResultRow[] = useMemo(() => {
    return ((data?.data as ScheduleResultRecord[]) || []).map(toScheduleResultRow);
  }, [data]);

  // 고정된 work type 필터 옵션 (데이터 변경과 무관하게 유지)
  const workTypeOptions = useMemo(() => [{label: '전체', value: ''}, ...WORK_TYPE_OPTIONS], []);

  // 진행 중인 스케줄이 있을 때 폴링 활성화
  const hasActiveSchedules = useMemo(
    () => displayData.some((row) => row.status === STATUS_RUNNING || row.status === STATUS_PENDING),
    [displayData],
  );

  useEffect(() => {
    if (!hasActiveSchedules) return;
    const interval = setInterval(() => {
      refetch();
    }, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [hasActiveSchedules, refetch]);

  useEffect(() => {
    setCurrentPage(1);
  }, [workTypeFilter, dateRange]);

  const handleScheduleNameChange = useCallback((v: string) => setScheduleNameInput(v), []);
  const renderLeftFilters = useCallback(
    () => [
      <VariableSelect
        key="work-type"
        label="제어 명령"
        placeholder="전체"
        options={workTypeOptions}
        value={workTypeFilter}
        showAllOption={false}
        onChange={(v) => setWorkTypeFilter((v as string) || '')}
      />,
      <SearchInput
        key="schedule-name"
        size="hsmall"
        placeholder="스케줄 이름 검색..."
        searchValue={scheduleNameInput}
        onSearchChange={handleScheduleNameChange}
      />,
      <Popover key="date-range" open={dateRangeOpen} onOpenChange={setDateRangeOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className="h-8 justify-start px-2.5 font-normal text-sm gap-1.5"
          >
            <CalendarIcon className="h-3.5 w-3.5" />
            <span className="font-semibold">기간</span>
            {dateRange?.from ? (
              dateRange.to ? (
                <span>
                  {formatDate(dateRange.from)} ~ {formatDate(dateRange.to)}
                </span>
              ) : (
                <span>{formatDate(dateRange.from)}</span>
              )
            ) : (
              <span className="text-muted-foreground">날짜 선택</span>
            )}
            {dateRange?.from && (
              <span
                role="button"
                onClick={(e) => {
                  e.stopPropagation();
                  setDateRange(undefined);
                }}
                className="ml-1 text-muted-foreground hover:text-foreground"
              >
                ✕
              </span>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-auto p-0">
          <DateRangePickerContent
            value={dateRange}
            onChange={(range) => setDateRange(range)}
            onClose={() => setDateRangeOpen(false)}
          />
        </PopoverContent>
      </Popover>,
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [workTypeOptions, workTypeFilter, dateRange, dateRangeOpen, handleScheduleNameChange],
  );
  const renderRightFilters = useCallback(() => {
    const items = [];
    if (hasActiveSchedules) {
      items.push(
        <span key="polling" className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span
            className={`w-1.5 h-1.5 rounded-full bg-green-500 ${isFetching ? 'animate-pulse' : ''}`}
          />
          실시간 갱신 중
        </span>,
      );
    }
    return items;
  }, [hasActiveSchedules, isFetching]);

  const columns = getScheduleResultColumns();

  return (
    <div className="h-full flex flex-col p-4">
      <DataGrid
        data={displayData}
        columns={columns}
        rowCursor={true}
        getRowId={(row: ScheduleResultRow) => row.result_id}
        onRowClick={(row: ScheduleResultRow) =>
          onRowClick(row.result_id, row.work_type, row.schedule_name, row.request_time)
        }
        isLoading={isLoading}
        enablePagination
        totalCount={data?.total}
        paginationConfig={{
          pageSize,
          pageIndex: currentPage - 1,
        }}
        pageSizes={[10, 50, 100, 500]}
        onPageChange={(pageIndex) => {
          setCurrentPage(pageIndex);
        }}
        onPageSizeChange={(newPageSize) => {
          setPageSize(newPageSize);
          setCurrentPage(1);
        }}
        leftFilters={renderLeftFilters}
        rightFilters={renderRightFilters}
      />
    </div>
  );
}
