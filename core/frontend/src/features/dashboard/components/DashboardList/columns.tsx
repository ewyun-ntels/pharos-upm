import {ColumnDef} from '@tanstack/react-table';
import {ArrowDownIcon, ArrowUpIcon, CaretSortIcon} from '@radix-ui/react-icons';
import {Home} from '@pharos/shared/components';
import {IconButton, ToggleIconButton} from '@pharos/shared/components/ui-extension';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import {FolderInput, FolderOpen, FolderMinus, Pencil, Trash2, Folder as LucideFolder, Settings} from 'lucide-react';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import type {DashboardData, DashboardFolder} from '@pharos/shared/types/dashboard';

// ============================================================================
// Tree Row 타입
// ============================================================================

export type DashboardRow = DashboardData & {rowType: 'dashboard'};
export type FolderRow = DashboardFolder & {rowType: 'folder'; subRows: DashboardRow[]};
export type DashboardTreeRow = DashboardRow | FolderRow;

// ============================================================================

interface DashboardColumnsProps {
  router: {push: (path: string) => void};
  homeId: string;
  setHomeId: (id: string) => void;
  hasUserRole?: boolean;
  onRenameFolder?: (folder: DashboardFolder) => void;
  onDeleteFolder?: (folderId: string) => void;
  onMoveToFolder?: (dashboardId: string, folderId: string) => void;
  onRemoveFromFolder?: (dashboardId: string) => void;
  folders?: DashboardFolder[];
}

export function getDashboardColumns({
  router,
  homeId,
  setHomeId,
  hasUserRole = false,
  onRenameFolder,
  onDeleteFolder,
  onMoveToFolder,
  onRemoveFromFolder,
  folders = [],
}: DashboardColumnsProps): ColumnDef<DashboardTreeRow>[] {

  const titleCol: ColumnDef<DashboardTreeRow> = {
    id: 'title',
    minSize: 120,
    meta: {flex: 1},
    accessorFn: (row) => (row.rowType === 'folder' ? row.name : row.config.title),
    header: ({column}) => {
      const isSorted = column.getIsSorted();
      return (
        <div
          className="flex flex-1 items-center gap-2 text-[13px] p-0 -mb-px rounded-none border-primary hover:text-primary active:text-primary dark:border-primary dark:hover:text-primary dark:active:text-primary cursor-pointer py-2"
          onClick={() => {
            if (isSorted === false) column.toggleSorting(true);
            else if (isSorted === 'desc') column.toggleSorting(false);
            else if (isSorted === 'asc') column.clearSorting();
          }}
        >
          Title{' '}
          {isSorted === 'asc' ? (
            <ArrowUpIcon className="h-4 w-4 shrink-0" />
          ) : isSorted === 'desc' ? (
            <ArrowDownIcon className="h-4 w-4 shrink-0" />
          ) : (
            <CaretSortIcon className="h-4 w-4 shrink-0" />
          )}
        </div>
      );
    },
    enableSorting: true,
    sortingFn: 'text',
    cell: ({row}) => {
      if (row.original.rowType === 'folder') {
        return (
          <div className="flex items-center gap-2">
            <FolderOpen size={14} className="text-muted-foreground shrink-0" />
            <span className="font-medium text-sm">{row.original.name}</span>
          </div>
        );
      }

      const {displayName, title, panels} = row.original.config;
      const isEmpty = !panels || panels.length === 0;
      const displayText = displayName || title;

      return (
        <TooltipProvider delayDuration={100} disableHoverableContent>
          <Tooltip>
            <TooltipTrigger asChild>
              <div
                onClick={() => router.push(`/dashboards/${row.original.id}`)}
                className="shadow-none cursor-pointer text-ellipsis overflow-hidden whitespace-nowrap"
              >
                {displayText}
                {isEmpty ? ' (Empty)' : ''}
              </div>
            </TooltipTrigger>
            <TooltipContent>{title}</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
    },
  };

  const actionsCol: ColumnDef<DashboardTreeRow> = {
    id: 'Actions',
    size: 114,
    minSize: 114,
    maxSize: 114,
    enableResizing: false,
    header: '',
    cell: ({row}) => {
      if (row.original.rowType === 'folder') {
        if (!hasUserRole) return null;
        const folder = row.original as DashboardFolder;
        return (
          <div className="flex items-center gap-2">
            <IconButton
              variant="ghost"
              size="icon-xs"
              onClick={() => onRenameFolder?.(folder)}
              icon={<Pencil/>}
            >
              Rename
            </IconButton>
            <IconButton
              variant="destructive"
              size="icon-xs"
              onClick={() => onDeleteFolder?.(folder.id)}
              icon={<Trash2/>}
            >
              Delete
            </IconButton>
          </div>
        );
      }

      const dashboardData = row.original as DashboardData;
      const permission = dashboardData.permission;
      const isOwner = permission === ActionSchema.enum.owner;
      const canEdit =
        permission === ActionSchema.enum.owner || permission === ActionSchema.enum.editor;
      const isInFolder = row.depth > 0;

      return (
        <div className="flex items-center gap-2">
          {/* Edit: editor 이상 */}
          {canEdit && (
            <IconButton
            variant="ghost"
            size="icon-xs"
              onClick={() => router.push(`/dashboards/${dashboardData.id}/settings`)}
              icon={<Settings/>}
            >
              Settings
            </IconButton>
          )}
          {/* 폴더 이동: owner만 */}
          {isOwner && !isInFolder && folders.length > 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <span>
                  <IconButton variant="ghost" size="icon-xs" icon={<FolderInput size={13} />}>
                    Move to folder
                  </IconButton>
                </span>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                {folders.map((folder) => (
                  <DropdownMenuItem
                    key={folder.id}
                    className="cursor-pointer"
                    onClick={() => onMoveToFolder?.(dashboardData.id, folder.id)}
                  >
                    <FolderOpen size={13} className="mr-2" />
                    {folder.name}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          {isOwner && isInFolder && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <span>
                  <IconButton variant="ghost" size="icon-xs" icon={<LucideFolder/>}>
                    Move
                  </IconButton>
                </span>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem
                  className="cursor-pointer"
                  onClick={() => onRemoveFromFolder?.(dashboardData.id)}
                >
                  <FolderMinus size={13} className="mr-2" />
                  폴더에서 제거
                </DropdownMenuItem>
                {folders.map((folder) => (
                  <DropdownMenuItem
                    key={folder.id}
                    className="cursor-pointer"
                    onClick={() => onMoveToFolder?.(dashboardData.id, folder.id)}
                  >
                    <FolderOpen size={13} className="mr-2" />
                    {folder.name}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          {/* Home: 모든 권한 */}
          <ToggleIconButton
            icon={<Home />}
            size="icon-xs"
            tooltip="Set as home"
            toggled={homeId === dashboardData.id}
            onClick={() => setHomeId(dashboardData.id)}
          />
        </div>
      );
    },
  };

  return [titleCol, actionsCol] as ColumnDef<DashboardTreeRow>[];
}
