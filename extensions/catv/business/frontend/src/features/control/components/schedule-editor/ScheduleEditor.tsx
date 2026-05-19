'use client';

import {useState, useEffect, useMemo} from 'react';
import {useCreate, useUpdate, useList} from '@lib/data-provider';
import {Button, Calendar, Input, Label} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {TimePicker} from '@pharos/shared/components/ui-extension';
import {MultiSelect} from '@pharos/shared/components/ui-extension/multiselect';
import {toast} from '@pharos/shared/components';
import {CATV_PROVIDER_NAME, CATV_RESOURCES} from '../../../../providers';
import type {HierarchyItem, Schedule} from '../../../../providers';
import {AlertCircle} from 'lucide-react';
import {WORK_TYPE_OPTIONS} from '../constants';

const TIMEZONE_OPTIONS = [
  'UTC',
  'Asia/Seoul',
  'Asia/Tokyo',
  'Asia/Shanghai',
  'Asia/Singapore',
  'Asia/Kolkata',
  'Europe/London',
  'Europe/Berlin',
  'America/New_York',
  'America/Los_Angeles',
] as const;

function getTimezoneLabel(tz: string): string {
  try {
    const offset =
      new Intl.DateTimeFormat('en', {timeZone: tz, timeZoneName: 'shortOffset'})
        .formatToParts(new Date())
        .find((p) => p.type === 'timeZoneName')?.value ?? '';
    return `${tz} (${offset})`;
  } catch {
    return tz;
  }
}

function parseCronSpec(spec: string): {timezone: string; cron: string} {
  const match = spec.match(/^(?:CRON_)?TZ=(\S+)\s+(.+)$/);
  if (match) return {timezone: match[1], cron: match[2].trim()};
  return {timezone: 'UTC', cron: spec.trim()};
}

function formatScheduleDate(date: Date): string {
  const dateStr = date.toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'short',
  });
  const h = String(date.getHours()).padStart(2, '0');
  const m = String(date.getMinutes()).padStart(2, '0');
  return `${dateStr} ${h}:${m}`;
}

const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

interface ScheduleEditorProps {
  editingSchedule?: Schedule | null;
  onSuccess?: (name: string) => void;
  onCancel?: () => void;
}

export function ScheduleEditor({editingSchedule, onSuccess, onCancel}: ScheduleEditorProps) {
  // ── 스케줄 정보 ──────────────────────────────────────────────────────────────
  const [name, setName] = useState('');
  const [scheduleType, setScheduleType] = useState('immediately');
  const [scheduleDate, setScheduleDate] = useState<Date | undefined>(undefined);
  const [cronExpression, setCronExpression] = useState('0 0 0 * * *');
  const [cronTimezone, setCronTimezone] = useState(browserTimezone);

  // ── 대상 선택 ────────────────────────────────────────────────────────────────
  const [so, setSo] = useState('');
  const [l3, setL3] = useState<string[]>([]);
  const [cell, setCell] = useState<string[]>([]);

  // ── 제어 명령 ────────────────────────────────────────────────────────────────
  const [workType, setWorkType] = useState('stb_request_info');

  // ── 제출 오류 ────────────────────────────────────────────────────────────────
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {mutate: createSchedule, mutation: createMutation} = useCreate();
  const {mutate: updateSchedule, mutation: updateMutation} = useUpdate();
  const mutation = editingSchedule ? updateMutation : createMutation;

  // ── 계층 구조 조회 ───────────────────────────────────────────────────────────
  const soQuery = useList<HierarchyItem>({
    resource: CATV_RESOURCES.SOS,
    dataProviderName: CATV_PROVIDER_NAME,
    pagination: {currentPage: 1, pageSize: 1000, mode: 'server' as const},
  });

  const l3Query = useList<HierarchyItem>({
    resource: CATV_RESOURCES.L3S,
    dataProviderName: CATV_PROVIDER_NAME,
    meta: {soId: so},
    pagination: {currentPage: 1, pageSize: 1000, mode: 'server' as const},
    queryOptions: {enabled: !!so},
  });

  const cellQuery = useList<HierarchyItem>({
    resource: CATV_RESOURCES.CELLS,
    dataProviderName: CATV_PROVIDER_NAME,
    meta: {soId: so, l3Id: l3[0]},
    pagination: {currentPage: 1, pageSize: 1000, mode: 'server' as const},
    queryOptions: {enabled: !!so && l3.length > 0},
  });

  const soOptions =
    soQuery.query.data?.data
      ?.filter((item) => item.id && item.id.trim() !== '' && item.id !== '0')
      .map((item) => ({label: item.name, value: item.id})) || [];

  const l3OptionsWithMissing = useMemo(() => {
    const baseOptions =
      l3Query.query.data?.data
        ?.filter((item) => item.id && item.id.trim() !== '' && item.id !== '0')
        .map((item) => ({label: item.name, value: item.id})) || [];
    if (!editingSchedule?.l3s?.length) return {options: baseOptions, missingItems: []};
    const existingValues = new Set(baseOptions.map((opt) => opt.value));
    const missingItems = editingSchedule.l3s.filter((id) => !existingValues.has(id));
    return {
      options: [
        ...baseOptions,
        ...missingItems.map((id) => ({label: `${id} (삭제됨)`, value: id})),
      ],
      missingItems,
    };
  }, [l3Query.query.data, editingSchedule]);

  const cellOptionsWithMissing = useMemo(() => {
    const baseOptions =
      cellQuery.query.data?.data
        ?.filter((item) => item.id && item.id.trim() !== '' && item.id !== '0')
        .map((item) => ({label: item.name, value: item.id})) || [];
    if (!editingSchedule?.cells?.length) return {options: baseOptions, missingItems: []};
    const existingValues = new Set(baseOptions.map((opt) => opt.value));
    const missingItems = editingSchedule.cells.filter((id) => !existingValues.has(id));
    return {
      options: [
        ...baseOptions,
        ...missingItems.map((id) => ({label: `${id} (삭제됨)`, value: id})),
      ],
      missingItems,
    };
  }, [cellQuery.query.data, editingSchedule]);

  const loadingSo = soQuery.query.isLoading;
  const loadingL3 = l3Query.query.isLoading;
  const loadingCell = cellQuery.query.isLoading;

  // ── 초기화 ───────────────────────────────────────────────────────────────────
  useEffect(() => {
    if (editingSchedule) {
      setName(editingSchedule.name);
      setScheduleType(editingSchedule.schedule_type);
      if (editingSchedule.schedule_spec_once) {
        setScheduleDate(new Date(editingSchedule.schedule_spec_once));
      } else {
        setScheduleDate(undefined);
      }
      const parsed = parseCronSpec(editingSchedule.schedule_spec_repeat ?? '');
      setCronExpression(parsed.cron || '0 0 0 * * *');
      setCronTimezone(parsed.timezone || browserTimezone);
      setSo((editingSchedule.sos && editingSchedule.sos[0]) || '');
      setL3(editingSchedule.l3s || []);
      setCell(editingSchedule.cells || []);
      setWorkType(editingSchedule.work_type);
    } else {
      setName('');
      setScheduleType('immediately');
      setScheduleDate(undefined);
      setCronExpression('0 0 0 * * *');
      setCronTimezone(browserTimezone);
      setSo('');
      setL3([]);
      setCell([]);
      setWorkType('stb_request_info');
    }
  }, [editingSchedule]);

  useEffect(() => {
    if (!so) setL3([]);
  }, [so]);

  useEffect(() => {
    if (l3.length === 0) setCell([]);
  }, [l3]);

  // ── 유효성 검사 ──────────────────────────────────────────────────────────────
  const determineControlTarget = (): {type: string; ids: string[]} | null => {
    if (cell.length > 0 && cell.length < cellOptionsWithMissing.options.length) {
      return {type: 'cell', ids: cell};
    }
    if (l3.length > 0 && l3.length < l3OptionsWithMissing.options.length) {
      return {type: 'l3', ids: l3};
    }
    if (
      (cell.length === cellOptionsWithMissing.options.length ||
        cellOptionsWithMissing.options.length === 0) &&
      (l3.length === l3OptionsWithMissing.options.length ||
        l3OptionsWithMissing.options.length === 0)
    ) {
      return {type: 'so', ids: [so]};
    }
    return null;
  };

  const isFormValid = () => {
    if (!name.trim()) return false;
    if (scheduleType === 'once' && !scheduleDate) return false;
    if (scheduleType === 'repeat' && !cronExpression.trim()) return false;
    if (!so) return false;
    return determineControlTarget() !== null;
  };

  // ── 제출 ─────────────────────────────────────────────────────────────────────
  const handleSubmit = () => {
    setSubmitError(null);
    const target = determineControlTarget();
    if (!target) {
      toast.warning('전송 대상을 선택해주세요.');
      return;
    }

    const scheduleData: any = {
      name,
      schedule_type: scheduleType,
      sos: so ? [so] : [],
      l3s: l3.filter((id) => id && id.trim() !== '' && id !== '0'),
      cells: cell.filter((id) => id && id.trim() !== '' && id !== '0'),
      settopboxes: [],
      area_type: target.type,
      area_ids: target.ids,
      work_type: workType,
      work_value: null,
    };

    if (scheduleType === 'once' && scheduleDate) {
      scheduleData.schedule_spec_once = scheduleDate.toISOString();
    } else if (scheduleType === 'repeat') {
      scheduleData.schedule_spec_repeat = `TZ=${cronTimezone} ${cronExpression}`;
    }

    const mutate = editingSchedule ? updateSchedule : createSchedule;
    const successMessage = editingSchedule ? '수정되었습니다' : '등록되었습니다';

    mutate(
      {
        resource: CATV_RESOURCES.SCHEDULES,
        dataProviderName: CATV_PROVIDER_NAME,
        ...(editingSchedule ? {id: editingSchedule.name} : {}),
        values: scheduleData,
      },
      {
        onSuccess: (data) => {
          toast.success(`스케줄이 성공적으로 ${successMessage}.`);
          onSuccess?.(data.data?.name || name);
        },
        onError: (error: any) => {
          const message =
            error?.response?.data?.message ||
            error?.message ||
            error?.statusText ||
            (typeof error === 'string' ? error : '알 수 없는 오류');
          const label = editingSchedule ? '수정' : '등록';
          toast.error(`스케줄 ${label} 실패: ${message}`);
          setSubmitError(`스케줄 ${label}에 실패했습니다. ${message}`);
        },
      },
    );
  };

  const controlTarget = determineControlTarget();

  return (
    <div className="h-full overflow-y-auto">
      <section className="flex flex-col w-full px-5 pt-5 pb-5">
        <div className="max-w-3xl space-y-6">
          {/* ── 스케줄 이름 ──────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">
              스케줄 이름 <span className="text-red-500">*</span>
            </Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="스케줄 이름을 입력하세요 (예: daily-info-collection)"
              disabled={!!editingSchedule}
            />
          </div>

          {/* ── 실행 타입 ────────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">
              실행 타입 <span className="text-red-500">*</span>
            </Label>
            <Select value={scheduleType} onValueChange={setScheduleType}>
              <SelectTrigger id="scheduleType" className="h-8 shadow-none">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="z-[150]">
                <SelectItem value="immediately">즉시 실행</SelectItem>
                <SelectItem value="once">한번 실행</SelectItem>
                <SelectItem value="repeat">반복 실행</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* ── 즉시 실행 안내 ───────────────────────────────────────────── */}
          {scheduleType === 'immediately' && (
            <div className="p-3 bg-muted rounded-md text-sm text-muted-foreground">
              스케줄 등록과 동시에 즉시 실행됩니다.
            </div>
          )}

          {/* ── 한번 실행 — 달력 + 시간 ─────────────────────────────────── */}
          {scheduleType === 'once' && (
            <div>
              <Label className="text-sm font-medium leading-none">
                실행 일시 <span className="text-red-500">*</span>
              </Label>
              <div className="flex items-start gap-4">
                <div className="w-fit border rounded-md">
                  <Calendar
                    mode="single"
                    selected={scheduleDate}
                    onSelect={(date) => {
                      if (date) {
                        const prev = scheduleDate ?? new Date();
                        date.setHours(prev.getHours(), prev.getMinutes(), 0, 0);
                        setScheduleDate(date);
                      }
                    }}
                    disabled={(date) => date < new Date(new Date().setHours(0, 0, 0, 0))}
                  />
                  <div className="border-t px-3 py-2">
                    <TimePicker
                      value={scheduleDate ?? new Date()}
                      onChange={(date) => {
                        const base = scheduleDate ? new Date(scheduleDate) : new Date();
                        base.setHours(date.getHours(), date.getMinutes(), 0, 0);
                        setScheduleDate(base);
                      }}
                      timePicker={{hour: true, minute: true}}
                    />
                  </div>
                </div>
                <div className="pt-2 space-y-1 min-w-0">
                  {scheduleDate ? (
                    <>
                      <p className="text-sm font-semibold">{formatScheduleDate(scheduleDate)}</p>
                      <p className="text-xs text-muted-foreground">{browserTimezone}</p>
                    </>
                  ) : (
                    <p className="text-sm text-muted-foreground">날짜를 선택하세요</p>
                  )}
                </div>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                지정한 시간에 한 번만 실행됩니다. 브라우저 현지 시간 기준.
              </p>
            </div>
          )}

          {/* ── 반복 실행 — 타임존 + Cron ──────────────────────────────── */}
          {scheduleType === 'repeat' && (
            <div className="space-y-4">
              <div>
                <Label className="text-sm font-medium leading-none">
                  타임존 <span className="text-red-500">*</span>
                </Label>
                <Select value={cronTimezone} onValueChange={setCronTimezone}>
                  <SelectTrigger id="cronTimezone">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="z-[150]">
                    {TIMEZONE_OPTIONS.map((tz) => (
                      <SelectItem key={tz} value={tz}>
                        {getTimezoneLabel(tz)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label className="text-sm font-medium leading-none">
                  Cron 표현식 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="cronExpression"
                  value={cronExpression}
                  onChange={(e) => setCronExpression(e.target.value)}
                  placeholder="0 0 0 * * * (매일 자정)"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  초 분 시 일 월 요일 형식 (예: 0 0 9 * * * = 매일 오전 9시)
                </p>
              </div>
            </div>
          )}

          {/* ── SO 선택 ──────────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">
              SO <span className="text-red-500">*</span>
            </Label>
            <Select value={so} onValueChange={setSo} disabled={loadingSo}>
              <SelectTrigger id="so">
                <SelectValue placeholder={loadingSo ? '로딩 중...' : 'SO를 선택하세요'} />
              </SelectTrigger>
              <SelectContent className="z-[150]">
                {soOptions.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* ── L3 선택 ──────────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">L3 (기본: All)</Label>
            {l3OptionsWithMissing.missingItems.length > 0 && (
              <div className="flex items-center gap-2 p-2 mb-2 text-sm text-red-600 bg-red-50 rounded-md border border-red-200">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>
                  선택된 L3 중 {l3OptionsWithMissing.missingItems.length}개 항목이 현재 존재하지
                  않습니다 (삭제됨 표시)
                </span>
              </div>
            )}
            <MultiSelect
              key={`l3-${so}-${editingSchedule?.name || 'new'}`}
              options={l3OptionsWithMissing.options}
              onValueChange={setL3}
              defaultSelectAll={true}
              useSelectAll={true}
              initValue={l3}
              placeholder={
                !so
                  ? '먼저 SO를 선택하세요'
                  : loadingL3
                    ? '로딩 중...'
                    : l3OptionsWithMissing.options.length === 0
                      ? 'L3 장비가 없습니다'
                      : 'L3 장비를 선택하세요'
              }
            />
          </div>

          {/* ── Cell 선택 ────────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">Cell (기본: All)</Label>
            {cellOptionsWithMissing.missingItems.length > 0 && (
              <div className="flex items-center gap-2 p-2 mb-2 text-sm text-red-600 bg-red-50 rounded-md border border-red-200">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>
                  선택된 Cell 중 {cellOptionsWithMissing.missingItems.length}개 항목이 현재 존재하지
                  않습니다 (삭제됨 표시)
                </span>
              </div>
            )}
            <MultiSelect
              key={`cell-${so}-${l3.join(',')}-${editingSchedule?.name || 'new'}`}
              options={cellOptionsWithMissing.options}
              onValueChange={setCell}
              defaultSelectAll={true}
              useSelectAll={true}
              initValue={cell}
              placeholder={
                !so
                  ? '먼저 SO를 선택하세요'
                  : l3.length === 0
                    ? '먼저 L3를 선택하세요'
                    : loadingCell
                      ? '로딩 중...'
                      : cellOptionsWithMissing.options.length === 0
                        ? 'Cell이 없습니다'
                        : 'Cell을 선택하세요'
              }
            />
          </div>

          {/* ── 제어 명령 ────────────────────────────────────────────────── */}
          <div>
            <Label className="text-sm font-medium leading-none">제어 명령</Label>
            <Select value={workType} onValueChange={setWorkType}>
              <SelectTrigger id="workType">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="z-[150]">
                {WORK_TYPE_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* ── 설정 요약 ────────────────────────────────────────────────── */}
          <div className="rounded-md border p-4 bg-muted/30 text-sm space-y-1">
            <p className="font-semibold mb-2 text-muted-foreground">설정 요약</p>
            <p>
              <span className="font-medium">이름:</span> {name || '—'}
            </p>
            <p>
              <span className="font-medium">실행 타입:</span>{' '}
              {scheduleType === 'immediately'
                ? '즉시 실행'
                : scheduleType === 'once'
                  ? '한번 실행'
                  : '반복 실행'}
            </p>
            {scheduleType === 'once' && scheduleDate && (
              <p>
                <span className="font-medium">실행 일시:</span> {formatScheduleDate(scheduleDate)}
                <span className="text-muted-foreground ml-1">({browserTimezone})</span>
              </p>
            )}
            {scheduleType === 'repeat' && (
              <p>
                <span className="font-medium">Cron:</span> {cronExpression}{' '}
                <span className="text-muted-foreground">({cronTimezone})</span>
              </p>
            )}
            <p className="pt-2 border-t mt-2">
              <span className="font-medium">SO:</span>{' '}
              {soOptions.find((o) => o.value === so)?.label || '—'}
            </p>
            <p>
              <span className="font-medium">L3:</span>{' '}
              {l3.length === 0
                ? '—'
                : l3.length === l3OptionsWithMissing.options.length
                  ? 'All'
                  : `${l3.length}개 선택`}
              {l3OptionsWithMissing.missingItems.length > 0 && (
                <span className="text-red-600 ml-2">
                  (삭제됨: {l3OptionsWithMissing.missingItems.length}개)
                </span>
              )}
            </p>
            <p>
              <span className="font-medium">Cell:</span>{' '}
              {cell.length === 0
                ? '—'
                : cell.length === cellOptionsWithMissing.options.length
                  ? 'All'
                  : `${cell.length}개 선택`}
              {cellOptionsWithMissing.missingItems.length > 0 && (
                <span className="text-red-600 ml-2">
                  (삭제됨: {cellOptionsWithMissing.missingItems.length}개)
                </span>
              )}
            </p>
            <p>
              <span className="font-medium">전송 타입:</span> {controlTarget?.type || '—'}
            </p>
            <p className="pt-2 border-t mt-2">
              <span className="font-medium">제어 명령:</span>{' '}
              {WORK_TYPE_OPTIONS.find((o) => o.value === workType)?.label || workType}
            </p>
          </div>

          {/* ── 제출 오류 ──────────────────────────────────────────────── */}
          {submitError && (
            <div className="flex items-start gap-2 p-3 text-sm text-red-600 bg-red-50 rounded-md border border-red-200">
              <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>{submitError}</span>
            </div>
          )}

          {/* ── 버튼 ─────────────────────────────────────────────────────── */}
          <div className="flex justify-start gap-2 mt-6">
            <Button variant="outline" onClick={onCancel} type="button">
              취소
            </Button>
            <Button
              type="button"
              disabled={!isFormValid() || mutation.isPending}
              onClick={handleSubmit}
            >
              {mutation.isPending
                ? `스케줄 ${editingSchedule ? '수정' : '등록'} 중...`
                : `스케줄 ${editingSchedule ? '수정' : '등록'}`}
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}
