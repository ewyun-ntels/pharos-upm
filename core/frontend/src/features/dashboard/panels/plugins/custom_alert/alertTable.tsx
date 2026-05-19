import React, { useState, useMemo, useEffect, useRef } from 'react';
import { Play, Pause } from '@pharos/shared/components';
import SeverityTabs from './severityTabs';
import {
  Switch,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@pharos/shared/components/ui';
import { DataGrid, IconButton, SearchInput } from '@pharos/shared/components/ui-extension';
import { useList } from '@/lib/data-provider';
import AlertDetailsDialog from './alertDetailsDialog';
import { useAlertTableColumns } from './columns';
import { PanelProps } from '@features/dashboard/panels/registry';
import { CustomAlertPanelOptions } from './types';
import { ALERT_RESOURCES } from '@providers/alert-provider/types';
import { Download } from '@pharos/shared/components';
import { DASHBOARD_PROVIDER_NAME } from '@providers/dashboard-provider';
import { useTableColumns } from '@pharos/shared/hooks/table-columns';
import { Info } from 'lucide-react';

export default function AlertTable(props: PanelProps<CustomAlertPanelOptions>) {
  const { title, description } = props;
  const { columns: alertColumns } = useAlertTableColumns();

  // ColumnConfig → ColumnDef 변환
  const { columns: tableColumns } = useTableColumns({
    columnConfigs: alertColumns,
  });

  const cardRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const [tableCalcHeight, setTableCalcHeight] = useState<string>('100%');

  const [tableData, setTableData] = useState<any[]>([]);
  const [statusTab, setStatusTab] = useState<string>('');
  const [confirmToggle, setConfirmToggle] = useState<boolean>(false);
  // Alert 테이블은 항상 1초마다 갱신 (기본값 1000ms)
  const [interval, setInterval] = useState<number | false>(1000);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedRow, setSelectedRow] = useState<any>(null);

  // ============================================================================
  // Layout - Height Calculation
  // ============================================================================

  useEffect(() => {
    const updateHeights = () => {
      const cardHeight = cardRef.current?.getBoundingClientRect().height || 0;
      const headerHeight = headerRef.current?.getBoundingClientRect().height || 0;
      const calculated = cardHeight - headerHeight;
      setTableCalcHeight(`${calculated}px`);
    };

    const resizeObserver = new ResizeObserver(() => {
      updateHeights();
    });

    if (cardRef.current) {
      resizeObserver.observe(cardRef.current);
    }

    window.addEventListener('resize', updateHeights);
    updateHeights();

    return () => {
      resizeObserver.disconnect();
      window.removeEventListener('resize', updateHeights);
    };
  }, []);

  const columnFilters = useMemo(() => {
    const filters: { id: string; value: boolean | string }[] = [
      {
        id: 'mask',
        value: confirmToggle,
      },
    ];

    if (statusTab) {
      filters.push({ id: 'severity', value: statusTab });
    }

    return filters;
  }, [statusTab, confirmToggle]);

  const filteredTableData = useMemo(() => {
    if (!tableData || tableData.length === 0) return tableData;

    return tableData.filter((row) => {
      return row.mask === confirmToggle;
    });
  }, [tableData, confirmToggle]);

  // ============================================================================
  // Filters
  // ============================================================================

  const renderLeftFilters = (table: any) => [
    <SearchInput key="search" size="hsmall" autoComplete="off" table={table} />,
    <SeverityTabs
      key="status"
      data={filteredTableData}
      statusTab={statusTab}
      setStatusTab={setStatusTab}
    />,
  ];

  const renderRightFilters = () => [
    <div key="confirm" className="flex items-center space-between space-x-2">
      <small className="text-[12px] text-muted-foreground font-right w-fit text-nowrap leading-none">
        Hidden
      </small>
      <Switch checked={confirmToggle} onCheckedChange={(val) => setConfirmToggle(val)} />
    </div>,
    <IconButton
      key="pauseBtn"
      variant="outline"
      size="icon"
      onClick={() => {
        if (typeof interval === 'number') {
          setInterval(false);
        } else {
          setInterval(1000);
        }
      }}
      icon={typeof interval === 'number' ? <Pause /> : <Play />}
    >
      {typeof interval === 'number' ? 'Pause' : 'Resume'}
    </IconButton>,
    <IconButton
      key="export"
      icon={<Download className="size-4" />}
      variant="outline"
      size="icon"
      onClick={() => { }}
    />,
  ];

  // ============================================================================
  // Meta & Data Fetch Hook
  // ============================================================================

  const meta = useMemo(() => {
    return {
      queryObject: {
        id: props.id ?? '',
        kind: props.kind,
        dashboardId: props.dashboardId ?? '',
        args: props.args,
        query: '',
        datasourceName: '',
      },
    };
  }, [props.id, props.kind, props.dashboardId, props.args]);

  const queryKey = useMemo(() => {
    const queryObject = meta?.queryObject;
    const argsArray = queryObject?.args
      ? Array.from(queryObject.args.entries()).sort()
      : [];

    return [
      DASHBOARD_PROVIDER_NAME,
      ALERT_RESOURCES.STATUS,
      'list',
      {
        id: queryObject?.id,
        dashboardId: queryObject?.dashboardId,
        args: argsArray,
      }
    ] as const;
  }, [meta]);

  const { query } = useList<any>({
    resource: ALERT_RESOURCES.STATUS,
    dataProviderName: DASHBOARD_PROVIDER_NAME,
    meta,
    pagination: { mode: 'off' },
    queryOptions: {
      queryKey,
      retry: false,
      refetchOnWindowFocus: false,
      throwOnError: false,
    },
  });

  const alertTableData = query.data?.data || [];

  return (
    <>
      <div ref={cardRef} className="relative h-full">
        <Card className="py-0 gap-0 rounded border-none shadow-none overflow-hidden">
          <div ref={headerRef} className="w-full">
            <CardHeader className="text-sm p-4 pb-2">
              <CardTitle className="flex items-center gap-2">
                {title}
                {description && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="whitespace-pre-line">{description}</p>
                    </TooltipContent>
                  </Tooltip>
                )}
              </CardTitle>
            </CardHeader>
          </div>
          <CardContent className="px-4" style={{ height: tableCalcHeight }}>
            <DataGrid<any>
              tableKey={`dashboard-${props.dashboardId}-panel-${props.id}`}
              persistState={true}
              data={alertTableData}
              isLoading={query.isLoading}
              columns={tableColumns}
              leftFilters={renderLeftFilters}
              rightFilters={renderRightFilters}
              tableHeight={tableCalcHeight}
              columnFilters={columnFilters}
              onDataLoaded={setTableData}
              onRowClick={(row: any) => {
                setSelectedRow(row);
                setDialogOpen(true);
              }}
              title={title}
              displayName={title}
              exportConfig={{
                fileName: `Alert_status`,
              }}
              searchableColumns={['description', 'timestamp']}
            />
          </CardContent>
        </Card>
      </div>
      {selectedRow && (
        <AlertDetailsDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          dialogTitle="Alert Details"
          alertData={selectedRow}
        />
      )}
    </>
  );
}
