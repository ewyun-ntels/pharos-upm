import {useState, useCallback} from 'react';
import {TableTabsTemplate, Tab, Tabs} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@pharos/core/components/breadcrumb';
import {ScheduleResult} from './components/schedule-result/ScheduleResult';
import {ScheduleList} from './components/schedule-list/ScheduleList';
import {ScheduleResultDetail} from './components/schedule-result-detail/ScheduleResultDetail';
import {ScheduleEditor} from './components/schedule-editor/ScheduleEditor';
import {useHasPermission} from '@pharos/shared/features/auth';
import {CATV_PERMISSIONS} from '../../permissions';
import type {Schedule} from '../../providers';

const RESULT_TAB_NAME_MAX_LENGTH = 10;
const FORM_TAB_NAME_MAX_LENGTH = 12;

interface ResultTab {
  resultId: string;
  workType: string;
  tabName: string;
}

function buildTabName(scheduleName: string | undefined, requestTime: string): string {
  const timeStr = (() => {
    try {
      const d = new Date(requestTime);
      const M = d.getMonth() + 1;
      const D = String(d.getDate()).padStart(2, '0');
      const HH = String(d.getHours()).padStart(2, '0');
      const mm = String(d.getMinutes()).padStart(2, '0');
      return `${M}/${D} ${HH}:${mm}`;
    } catch {
      return requestTime.slice(0, 10);
    }
  })();
  if (!scheduleName) return timeStr;
  const nameLabel =
    scheduleName.length > RESULT_TAB_NAME_MAX_LENGTH
      ? scheduleName.slice(0, RESULT_TAB_NAME_MAX_LENGTH) + '\u2026'
      : scheduleName;
  return `${nameLabel} \u00b7 ${timeStr}`;
}

export default function ControlPage() {
  const canRead = useHasPermission(CATV_PERMISSIONS.Read);
  const canCreate = useHasPermission(CATV_PERMISSIONS.Create);
  const canUpdate = useHasPermission(CATV_PERMISSIONS.Update);

  const [editingSchedule, setEditingSchedule] = useState<Schedule | null>(null);
  const [formMode, setFormMode] = useState<'create' | 'edit' | null>(null);
  const [resultTabs, setResultTabs] = useState<ResultTab[]>([]);
  const [activeTab, setActiveTab] = useState<string>('스케줄 목록');

  const formTabName =
    formMode === 'edit' && editingSchedule
      ? `스케줄 수정 · ${editingSchedule.name.length > FORM_TAB_NAME_MAX_LENGTH ? editingSchedule.name.slice(0, FORM_TAB_NAME_MAX_LENGTH) + '…' : editingSchedule.name}`
      : '스케줄 등록';

  const handleRowClick = useCallback(
    (resultId: string, workType: string, scheduleName?: string, requestTime?: string) => {
      const existing = resultTabs.find((t) => t.resultId === resultId);
      if (existing) {
        setActiveTab(existing.tabName);
        return;
      }
      const tabName = buildTabName(scheduleName, requestTime ?? new Date().toISOString());
      setResultTabs((prev) => [...prev, {resultId, workType, tabName}]);
      setActiveTab(tabName);
    },
    [resultTabs],
  );

  const handleCloseTab = (tabName: string) => {
    if (tabName === formTabName && formMode !== null) {
      handleFormClose();
      return;
    }
    setResultTabs((prev) => prev.filter((t) => t.tabName !== tabName));
  };

  const handleScheduleEdit = (schedule: Schedule) => {
    setEditingSchedule(schedule);
    setFormMode('edit');
    setActiveTab(
      `스케줄 수정 · ${schedule.name.length > FORM_TAB_NAME_MAX_LENGTH ? schedule.name.slice(0, FORM_TAB_NAME_MAX_LENGTH) + '…' : schedule.name}`,
    );
  };

  const handleFormClose = () => {
    setFormMode(null);
    setEditingSchedule(null);
    setActiveTab('스케줄 목록');
  };

  const handleFormSuccess = (name: string) => {
    handleFormClose();
  };

  if (!canRead) {
    return (
      <main className="flex items-center justify-center w-full h-full text-muted-foreground text-sm">
        이 페이지에 접근할 권한이 없습니다. (extension:catv:read 필요)
      </main>
    );
  }

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate
        name="STB 제어 및 모니터링"
        showTabs={true}
        defaultTab="스케줄 목록"
        activeTab={activeTab}
        onTabChange={setActiveTab}
        onTabClose={handleCloseTab}
        breadcrumb={<PageBreadcrumb />}
      >
        <Tabs>
          <Tab name="스케줄 목록" closable={false}>
            <ScheduleList
              onCreateClick={
                canCreate
                  ? () => {
                      setFormMode('create');
                      setEditingSchedule(null);
                      setActiveTab('스케줄 등록');
                    }
                  : undefined
              }
              onScheduleEdit={canUpdate ? handleScheduleEdit : undefined}
            />
          </Tab>
          <Tab name="스케줄 결과" closable={false}>
            <ScheduleResult onRowClick={handleRowClick} />
          </Tab>
          {formMode !== null && (
            <Tab name={formTabName} closable={true}>
              <ScheduleEditor
                editingSchedule={editingSchedule}
                onSuccess={handleFormSuccess}
                onCancel={handleFormClose}
              />
            </Tab>
          )}
          {resultTabs.map(({resultId, workType, tabName}) => (
            <Tab
              key={resultId}
              name={tabName}
              label={
                <span className="flex items-baseline gap-1">
                  상세
                  <span className="text-xs font-normal text-muted-foreground">({tabName})</span>
                </span>
              }
              closable={true}
            >
              <ScheduleResultDetail resultId={resultId} workType={workType} />
            </Tab>
          ))}
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}
