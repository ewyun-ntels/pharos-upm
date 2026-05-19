import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {Routes, Route, useNavigate} from 'react-router-dom';
import {Button} from '@pharos/shared/components/ui';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {DashboardDialog} from '@features/dashboard/components/DashboardHeader/dashboardDialog';
import {useToast} from '@hooks/use-toast';

// Lazy load dashboard sub-pages
const DashboardShowPage = React.lazy(() => import('@/routes/dashboards/[dashboardId]/page'));
const DashboardEditPage = React.lazy(
  () => import('@/routes/dashboards/[dashboardId]/edit/[panelId]/page'),
);
const DashboardSettingsPage = React.lazy(
  () => import('@/routes/dashboards/[dashboardId]/settings/page'),
);
const DashboardVariableNewPage = React.lazy(
  () => import('@/routes/dashboards/[dashboardId]/variables/new/page'),
);
const DashboardVariableEditPage = React.lazy(
  () => import('@/routes/dashboards/[dashboardId]/variables/edit/page'),
);
import {TableTabsTemplate, Tabs, Tab} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import {handleSelectOne} from '@pharos/shared/hooks/table-columns';
import {
  getDashboardColumns,
  type DashboardTreeRow,
  type DashboardRow,
} from '@features/dashboard/components/DashboardList/columns';
import {useDownloadJSON} from '@features/dashboard/hooks/use-download-json';
import {useDashboardActions} from '@features/dashboard/hooks/use-dashboard-actions';
import {useFolders} from '@features/dashboard/hooks/use-folders';
import {FolderDialog} from '@features/dashboard/components/DashboardList/FolderDialog';
import {Plus} from '@pharos/shared/components';
import {SearchInput} from '@pharos/shared/components/ui-extension';
import {Download, FolderOpen, Pencil} from 'lucide-react';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import type {DashboardData, DashboardFolder} from '@pharos/shared/types/dashboard';
import {useHasPermission} from '@pharos/shared/features/auth/auth-store';
import {PermissionKeysSchema} from '@pharos/shared/types/role';
import {DashboardHistoryTab} from '@features/dashboard/components/history-table/DashboardHistoryTab';
import {dashboardProvider, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';

function DashboardList() {
  const navigate = useNavigate();

  const hasUserRole = useHasPermission(PermissionKeysSchema.enum['role:dashboard_create']);

  const [cardFormatData, setCardFormatData] = useState<DashboardData[]>([]);
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [dropOpenState, setDropOpenState] = useState(false);
  const [showAlert, setShowAlert] = useState(false);

  const [folderDialogOpen, setFolderDialogOpen] = useState(false);
  const [editingFolder, setEditingFolder] = useState<DashboardFolder | null>(null);
  const [editMode, setEditMode] = useState(false);

  const downloadJSON = useDownloadJSON();
  const {
    handleNewDashboard,
    handleDelete,
    handleSetHomeId,
    homeId,
    dashboardsData,
    folders,
    refetchDashboards,
  } = useDashboardActions();

  const {toast} = useToast();
  const {createFolder, renameFolder, deleteFolder, moveDashboard} = useFolders();

  // editor 또는 owner 권한인 대시보드가 하나라도 있으면 Edit 버튼 표시
  const hasEditAccess = useMemo(
    () => hasUserRole || cardFormatData.some((d) => d.permission !== ActionSchema.enum.viewer),
    [hasUserRole, cardFormatData],
  );

  useEffect(() => {
    if (dashboardsData?.data) {
      const sorted = [...dashboardsData.data].sort((a, b) =>
        (a.config?.title ?? '').localeCompare(b.config?.title ?? ''),
      );
      setCardFormatData(sorted);
    }
  }, [dashboardsData]);

  // 트리 데이터 구성: 폴더(부모) → 폴더 내 대시보드(자식), 루트 대시보드
  const folderIds = useMemo(() => new Set(folders.map((f) => f.id)), [folders]);

  const treeData = useMemo((): DashboardTreeRow[] => {
    const grouped = new Map<string, DashboardRow[]>();
    for (const d of cardFormatData) {
      if (d.folderId && folderIds.has(d.folderId)) {
        if (!grouped.has(d.folderId)) grouped.set(d.folderId, []);
        grouped.get(d.folderId)!.push({...d, rowType: 'dashboard'});
      }
    }

    const folderRows: DashboardTreeRow[] = folders.map((folder) => ({
      ...folder,
      rowType: 'folder',
      subRows: grouped.get(folder.id) ?? [],
    }));

    const rootRows: DashboardTreeRow[] = cardFormatData
      .filter((d) => !d.folderId || !folderIds.has(d.folderId))
      .map((d) => ({...d, rowType: 'dashboard'}));

    return [...folderRows, ...rootRows];
  }, [folders, folderIds, cardFormatData]);

  const handleDialogClose = useCallback(() => {
    setDropOpenState(false);
    setDialogOpen(false);
  }, []);

  const handleRenameFolder = useCallback((folder: DashboardFolder) => {
    setEditingFolder(folder);
    setFolderDialogOpen(true);
  }, []);

  const handleDeleteFolder = useCallback(
    async (folderId: string) => {
      const folder = folders.find((f) => f.id === folderId);
      if (
        window.confirm(
          `"${folder?.name ?? folderId}" 폴더를 삭제하시겠습니까? 폴더 내 대시보드는 루트로 이동됩니다.`,
        )
      ) {
        await deleteFolder(folderId);
        refetchDashboards();
      }
    },
    [folders, deleteFolder, refetchDashboards],
  );


  const handleMultiExport = useCallback(async () => {
    const selected = cardFormatData.filter((d) => selectedItems.includes(d.id));

    for (let i = 0; i < selected.length; i++) {
      const d = selected[i];
      try {
        // dashboardProvider.getOne으로 전체 데이터를 다시 불러옴
        const response = await dashboardProvider.getOne({
          resource: DASHBOARD_RESOURCES.DASHBOARD,
          id: d.id,
        });
        const fullData = response.data;

        // DashboardConfig만 export (서버 메타데이터 제외)
        const configToExport = fullData.config || fullData;

        setTimeout(() => {
          downloadJSON(configToExport, `${d.config.displayName || d.config.title}.json`);
        }, i * 150);
      } catch (error) {
        console.error(`Failed to export dashboard ${d.id}:`, error);
        toast({ description: `Failed to export ${d.config.title}`, variant: 'destructive' });
      }
    }
  }, [cardFormatData, selectedItems, downloadJSON, toast]);

  const handleRemoveFromFolder = useCallback(
    async (dashboardId: string) => {
      await moveDashboard(dashboardId, null);
      refetchDashboards();
    },
    [moveDashboard, refetchDashboards],
  );

  const handleMoveToFolder = useCallback(
    async (dashboardId: string, folderId: string) => {
      await moveDashboard(dashboardId, folderId);
      refetchDashboards();
    },
    [moveDashboard, refetchDashboards],
  );

  const checkboxConfig = useMemo(() => {
    if (!hasUserRole || !editMode) return undefined;
    return {
      checkboxKey: 'id' as const,
      selectedItems,
      isSelectable: (row: DashboardTreeRow) => row.rowType === 'dashboard' && row.permission === ActionSchema.enum.owner,
      handleSelectAll: (_checked: boolean, items: DashboardTreeRow[]) => {
        const ids = items
          .filter((r): r is DashboardRow => r.rowType === 'dashboard')
          .map((d) => d.id);
        setSelectedItems(_checked ? ids : []);
      },
      handleSelectOne: (item: DashboardTreeRow, checked: boolean) => {
        if (item.rowType !== 'dashboard') return;
        handleSelectOne({
          item,
          checked,
          selectedItems,
          keyName: 'id',
          setSelected: setSelectedItems,
        });
      },
    };
  }, [hasUserRole, editMode, selectedItems]);

  const columns = useMemo(
    () =>
      getDashboardColumns({
        router: {push: (path: string) => navigate(path)},
        homeId,
        setHomeId: handleSetHomeId,
        hasUserRole,
        onRenameFolder: handleRenameFolder,
        onDeleteFolder: handleDeleteFolder,
        onRemoveFromFolder: handleRemoveFromFolder,
        onMoveToFolder: handleMoveToFolder,
        folders,
      }),
    [
      navigate,
      homeId,
      handleSetHomeId,
      hasUserRole,
      handleRenameFolder,
      handleDeleteFolder,
      handleRemoveFromFolder,
      handleMoveToFolder,
      folders,
    ],
  );

  const deleteButton = (
    <Button
      size="default"
      variant="destructive"
      className="h-8 shadow-none px-3 py-1 cursor-pointer disabled:opacity-0"
      onClick={() => setShowAlert(true)}
      disabled={selectedItems.length === 0}
    >
      Delete {selectedItems.length > 0 ? `${selectedItems.length}` : ''}
    </Button>
  );

  const newDashboardButton = (
    <DropdownMenu
      onOpenChange={(e) => {
        if (e) setDropOpenState(false);
        else setDropOpenState(true);
      }}
    >
      <DropdownMenuTrigger asChild>
        <Button
          variant="default"
          size="default"
          className="h-8 shadow-none px-3 py-1 cursor-pointer"
        >
          <Plus />
          New dashboard
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="min-w-30 max-w-80 -translate-x-4">
        <DropdownMenuItem className="cursor-pointer">
          <span onClick={handleNewDashboard} className="h-6 w-full inline-block cursor-pointer">
            New dashboard
          </span>
        </DropdownMenuItem>
        <DropdownMenuItem className="cursor-pointer">
          <span
            onClick={() => setDialogOpen(true)}
            className="h-6 w-full inline-block cursor-pointer"
          >
            Import
          </span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const newFolderButton = hasUserRole ? (
    <Button
      variant="outline"
      size="default"
      className="h-8 shadow-none px-3 py-1 cursor-pointer"
      onClick={() => {
        setEditingFolder(null);
        setFolderDialogOpen(true);
      }}
    >
      <FolderOpen size={14} className="mr-1" />
      New folder
    </Button>
  ) : null;

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate name="Dashboard Management" breadcrumb={<PageBreadcrumb />}>
        <Tabs>
          <Tab name="Dashboards">
            <div className="px-5">
              <DataGrid
                data={treeData}
                columns={columns}
                getSubRows={(row) => (row.rowType === 'folder' ? row.subRows : undefined)}
                treeColumnId="title"
                checkboxConfig={checkboxConfig}
                visibilityState={{Actions: editMode}}
                leftFilters={(table) => [
                  <SearchInput
                    key="search"
                    size="hsmall"
                    placeholder="Search dashboards..."
                    table={table}
                  />,
                ]}
                rightFilters={() => [
                  <>
                    {hasEditAccess && (
                      <div className="flex items-center gap-2" key="actions">
                        {editMode && hasUserRole && deleteButton}
                        {editMode && hasUserRole && (
                          <Button
                            size="default"
                            variant="secondary"
                            className="h-8 shadow-none px-3 py-1"
                            onClick={handleMultiExport}
                            disabled={selectedItems.length === 0}
                          >
                            <Download size={14} className="mr-1" />
                            Export {selectedItems.length > 0 ? selectedItems.length : ''}
                          </Button>
                        )}
                        {editMode && hasUserRole && newFolderButton}
                        <Button
                          variant={editMode ? 'secondary' : 'outline'}
                          size="default"
                          className="h-8 shadow-none px-3 py-1 cursor-pointer"
                          onClick={() => {
                            setEditMode((v) => {
                              if (v) setSelectedItems([]);
                              return !v;
                            });
                          }}
                        >
                          <Pencil size={14} className="mr-1" />
                          {editMode ? 'Done' : 'Edit'}
                        </Button>
                        {hasUserRole && newDashboardButton}
                      </div>
                    )}
                    {!hasEditAccess && hasUserRole && newDashboardButton}
                  </>,
                ]}
                tableKey="dashboards"
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="History">
            <DashboardHistoryTab />
          </Tab>
        </Tabs>
      </TableTabsTemplate>

      {dropOpenState && (
        <DashboardDialog
          open={dialogOpen}
          onOpenChange={handleDialogClose}
          data={null}
          onSuccess={refetchDashboards}
        />
      )}
      <AlertDialog open={showAlert} onOpenChange={setShowAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure you want to delete?</AlertDialogTitle>
            <AlertDialogDescription>
              The selected item(s) will be deleted. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => handleDelete(selectedItems, setSelectedItems)}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <FolderDialog
        open={folderDialogOpen}
        onOpenChange={setFolderDialogOpen}
        initialName={editingFolder?.name ?? ''}
        onSave={async (name) => {
          if (editingFolder) {
            await renameFolder(editingFolder.id, name);
          } else {
            await createFolder(name);
          }
          setEditingFolder(null);
          refetchDashboards();
        }}
      />
    </main>
  );
}

export default function DashboardsRouter() {
  return (
    <React.Suspense fallback={<LoadingIndicator className="h-screen" />}>
      <Routes>
        <Route index element={<DashboardList />} />
        <Route path=":dashboardId" element={<DashboardShowPage />} />
        <Route path=":dashboardId/edit/:panelId" element={<DashboardEditPage />} />
        <Route path=":dashboardId/settings" element={<DashboardSettingsPage />} />
        <Route path=":dashboardId/variables/new" element={<DashboardVariableNewPage />} />
        <Route path=":dashboardId/variables/edit" element={<DashboardVariableEditPage />} />
      </Routes>
    </React.Suspense>
  );
}
