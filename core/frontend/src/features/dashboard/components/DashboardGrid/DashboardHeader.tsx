import React from 'react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import { IconButton } from '@pharos/shared/components/ui-extension';
import { Button } from '@pharos/shared/components/ui';
import { Settings, Save, Plus } from '@pharos/shared/components';
import { FilterConfig } from '@features/dashboard/components/VariableBar/types';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';

interface DashboardHeaderProps {
  header: FilterConfig[];
  permission: string;
  dropOpenState: boolean;
  setDropOpenState: (state: boolean) => void;
  setOpenDialog: (dialog: string | null) => void;
  setShowAlert: (show: boolean) => void;
  router: {
    push: (href: string) => void;
  };
  dashboardId?: string;
}


export function DashboardHeader({
  header,
  permission,
  dropOpenState,
  setDropOpenState,
  setOpenDialog,
  setShowAlert,
  router,
  dashboardId,
}: DashboardHeaderProps): FilterConfig[] {
  const additionalHeader: FilterConfig[] = [
    {
      id: 'jsonBtn',
      type: 'custom' as const,
      kind: 'headerR' as const,
      options: {
        customComponent: (
          <DropdownMenu open={dropOpenState} onOpenChange={setDropOpenState}>
            <DropdownMenuTrigger asChild>
              <div className="h-8 w-8 flex items-center">
                <IconButton variant={'outline'} icon={<Settings />}>
                  Settings
                </IconButton>
              </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="min-w-30 max-w-80 -translate-x-4">
              <DropdownMenuItem className="cursor-pointer">
                <span
                  onClick={() => {
                    setDropOpenState(false);
                    setOpenDialog('generalSetting');
                  }}
                  className="h-6 w-full inline-block cursor-pointer"
                >
                  General
                </span>
              </DropdownMenuItem>
              {permission === ActionSchema.enum.owner && (
                <DropdownMenuItem className="cursor-pointer">
                  <span
                    onClick={() => {
                      setDropOpenState(false);
                      setOpenDialog('permission');
                    }}
                    className="h-6 w-full inline-block cursor-pointer"
                  >
                    Permission
                  </span>
                </DropdownMenuItem>
              )}
              <DropdownMenuItem className="cursor-pointer">
                <span
                  onClick={() => {
                    setDropOpenState(false);
                    setOpenDialog('dashboard');
                  }}
                  className="h-6 w-full inline-block cursor-pointer"
                >
                  JSON editor
                </span>
              </DropdownMenuItem>
              <DropdownMenuItem className="cursor-pointer">
                <span
                  onClick={() => {
                    if (!dashboardId) {
                      console.error('[DashboardHeader] dashboardId is required');
                      return;
                    }
                    router.push(`/dashboards/${dashboardId}/settings${window.location.search}#variables`);
                  }}
                  className="h-6 w-full inline-block cursor-pointer"
                >
                  Variables
                </span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    },
    {
      id: 'custom',
      type: 'custom' as const,
      kind: 'headerR' as const,
      options: {
        customComponent: (
          <div className="h-8 w-8 flex items-center">
            <IconButton variant={'outline'} icon={<Save />} onClick={() => setShowAlert(true)}>
              Save dashboard
            </IconButton>
          </div>
        ),
      },
    },
  ];

  // ✅ Add Panel 버튼은 viewer가 아닐 때만 추가
  if (permission !== ActionSchema.enum.viewer) {
    additionalHeader.push({
      id: 'addButton',
      type: 'custom' as const,
      kind: 'headerR' as const,
      options: {
        customComponent: (
          <Button
            variant="default"
            className="h-8 shadow-none px-3 py-1"
            onClick={() => {
              const store = useDashboardStore.getState();

              // ✅ 1. Store에서 새 패널 생성하고 ID 받기
              const newPanelId = store.createPanel();

              // ✅ 2. dashboardId 확인
              if (!dashboardId) {
                console.error('[DashboardHeader] dashboardId is required');
                return;
              }

              // ✅ 3. 에디터로 이동 (path parameter 사용)
              const targetUrl = `/dashboards/${dashboardId}/edit/${newPanelId}`;
              router.push(targetUrl);
            }}
          >
            <span className="[&>*]:w-4 &>*]:h-4">
              <Plus className="stroke-[2]" />
            </span>
            Add panel
          </Button>
        ),
      },
    });
  }

  return [...header, ...additionalHeader];
}
