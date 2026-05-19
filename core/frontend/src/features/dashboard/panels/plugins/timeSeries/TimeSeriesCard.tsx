'use client';

import React, { useCallback, useState, useEffect } from 'react';
import { TimeSeriesChart, TimeSeriesPanelOptions } from '@pharos/shared/components/charts';
import { withCardChartPanel } from '../common/withCardChartPanel';
import { PanelContentProps } from '../common/withChartPanel';
import { useTranslation } from '@/lib/data-provider';
import { convertUnit, UnitType, formatLocalTime24 } from '@pharos/shared/lib/unitUtils';
import { Button } from '@pharos/shared/components/ui';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@pharos/shared/components/ui';
import type { Annotation } from '@pharos/shared/types/dashboard';
import { v4 as uuidv4 } from 'uuid';

const EMPTY_ANNOTATIONS: Annotation[] = [];

// ─── helpers ──────────────────────────────────────────────────────────────────

function toDatetimeLocal(ms: number): string {
  const d = new Date(ms);
  d.setSeconds(0, 0);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// ─── AnnotationModal ──────────────────────────────────────────────────────────

interface AnnotationModalProps {
  open: boolean;
  onClose: () => void;
  initialTimeMs: number;
  panelId: string;
  onSave?: (annotation: Annotation) => void | Promise<void>;
}

function AnnotationModal({ open, onClose, initialTimeMs, panelId, onSave }: AnnotationModalProps) {
  const [text, setText] = useState('');
  const [timeInput, setTimeInput] = useState(() => toDatetimeLocal(initialTimeMs));
  const [scope, setScope] = useState<'panel' | 'global'>('panel');
  const [color, setColor] = useState('#ef4444');

  // 모달이 열릴 때마다 폼 초기화 + 클릭된 시각으로 세팅
  useEffect(() => {
    if (open) {
      setText('');
      setTimeInput(toDatetimeLocal(initialTimeMs));
      setScope('panel');
      setColor('#ef4444');
    }
  }, [open, initialTimeMs]);

  const handleSave = useCallback(() => {
    if (!text.trim()) return;
    const annotation: Annotation = {
      id: uuidv4(),
      time: new Date(timeInput).getTime(),
      text: text.trim(),
      color,
      scope,
      ...(scope === 'panel' ? { panel_id: panelId } : {}),
    };
    onSave?.(annotation);
    onClose();
  }, [text, timeInput, color, scope, panelId, onSave, onClose]);

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Add Annotation</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3 py-2">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium">Time</label>
            <input
              type="datetime-local"
              value={timeInput}
              onChange={(e) => setTimeInput(e.target.value)}
              className="border rounded px-2 py-1 text-sm bg-background"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium">Label</label>
            <input
              type="text"
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Annotation text"
              className="border rounded px-2 py-1 text-sm bg-background"
              autoFocus
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium">Color</label>
            <input
              type="color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              className="h-8 w-16 rounded border cursor-pointer bg-background"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium">Scope</label>
            <div className="flex gap-4 text-sm">
              <label className="flex items-center gap-1.5 cursor-pointer">
                <input type="radio" value="panel" checked={scope === 'panel'} onChange={() => setScope('panel')} />
                This panel
              </label>
              <label className="flex items-center gap-1.5 cursor-pointer">
                <input type="radio" value="global" checked={scope === 'global'} onChange={() => setScope('global')} />
                All panels
              </label>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" size="sm" onClick={onClose}>Cancel</Button>
          <Button size="sm" onClick={handleSave} disabled={!text.trim()}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ─── TimeSeriesContent ────────────────────────────────────────────────────────

function TimeSeriesContent({
  chartData,
  options,
  timeRangeCallback,
  id,
  chartWidth,
  chartHeight,
  annotations = EMPTY_ANNOTATIONS,
  globalAnnotations,
  onAnnotationCreate,
  onAnnotationDelete,
}: PanelContentProps<TimeSeriesPanelOptions>) {
  const { translate } = useTranslation();

  const [modalState, setModalState] = useState<{ open: boolean; timeMs: number }>({
    open: false,
    timeMs: Date.now(),
  });

  const handleTimeRangeChange = useCallback(
    (startTime: number, endTime: number) => {
      if (timeRangeCallback) {
        const startMs = startTime > 10000000000 ? startTime : startTime * 1000;
        const endMs = endTime > 10000000000 ? endTime : endTime * 1000;
        timeRangeCallback(
          { type: 'absolute', absoluteValue: new Date(startMs) },
          { type: 'absolute', absoluteValue: new Date(endMs) },
        );
      }
    },
    [timeRangeCallback],
  );

  const handleChartClick = useCallback((timeMs: number) => {
    setModalState({ open: true, timeMs });
  }, []);

  // Stable convertUnit to prevent React.memo re-render on TimeSeriesChart
  const convertUnitFn = useCallback(
    (value: number, unit?: string, decimals?: number) =>
      convertUnit(value, (unit ?? 'short') as UnitType, decimals ?? options?.decimals ?? 2).toString(),
    [options?.decimals],
  );

  if (chartHeight === 0) return null;

  return (
    <div className="relative w-full h-full">
      <TimeSeriesChart
        data={chartData}
        loading={false}
        options={options}
        chartWidth={chartWidth}
        chartHeight={chartHeight}
        onTimeRangeChange={handleTimeRangeChange}
        formatLocalTime={formatLocalTime24}
        convertUnit={convertUnitFn}
        translate={translate}
        id={id}
        annotations={annotations}
        globalAnnotations={globalAnnotations ?? EMPTY_ANNOTATIONS}
        onChartClick={handleChartClick}
        onDeleteAnnotation={onAnnotationDelete}
      />
      {id && (
        <AnnotationModal
          open={modalState.open}
          onClose={() => setModalState((s) => ({ ...s, open: false }))}
          initialTimeMs={modalState.timeMs}
          panelId={id}
          onSave={onAnnotationCreate}
        />
      )}
    </div>
  );
}

export const TimeSeriesCardChart = withCardChartPanel(TimeSeriesContent);

