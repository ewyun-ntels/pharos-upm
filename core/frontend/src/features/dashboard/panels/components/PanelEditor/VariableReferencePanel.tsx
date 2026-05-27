import React from 'react';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@pharos/shared/components/ui';
import {
  TEMPLATE_VAR_START_TIME,
  TEMPLATE_VAR_END_TIME,
  TEMPLATE_VAR_STEP,
} from '@lib/query-params';

interface VariableReferencePanelProps {
  grafanaStyle?: boolean;
}

const PONGO2_VARIABLES = [
  { name: TEMPLATE_VAR_START_TIME, desc: '시작 시간', example: '1772686800' },
  { name: TEMPLATE_VAR_END_TIME, desc: '종료 시간', example: '1772690400' },
  { name: TEMPLATE_VAR_STEP, desc: '시간 간격(초)', example: '300' },
  { name: '{{filterId}}', desc: '단일 선택 필터', example: 'server1' },
  { name: '{{filterId}}', desc: '다중 선택 필터', example: "'tag1','tag2'" },
] as const;

const GRAFANA_VARIABLES = [
  { name: '$__from', desc: '시작 시간(ms)', example: '1772686800000' },
  { name: '$__to', desc: '종료 시간(ms)', example: '1772690400000' },
  { name: '${__from:date:seconds}', desc: '시작 시간(초)', example: '1772686800' },
  { name: '${__to:date:seconds}', desc: '종료 시간(초)', example: '1772690400' },
  { name: '$__interval', desc: '시간 간격', example: '300s' },
  { name: '$__interval_ms', desc: '시간 간격(ms)', example: '300000' },
  { name: '$filterId', desc: '선택 필터', example: 'server1' },
  { name: '${filterId:regex}', desc: '정규식 필터', example: '(server1|server2)' },
] as const;

const PONGO2_EXAMPLES = [
  {
    label: '시간 범위',
    query: `WHERE toDateTime(ts / 1000) BETWEEN toDateTime(${TEMPLATE_VAR_START_TIME} / 1000) AND toDateTime(${TEMPLATE_VAR_END_TIME} / 1000)`,
  },
  { label: '단일 선택 필터', query: "WHERE host = '{{Host}}'" },
  { label: '다중 선택 필터', query: 'WHERE tag IN ({{tags}})' },
] as const;

const GRAFANA_EXAMPLES = [
  { label: 'PromQL 필터', query: 'up{job=~"$job", pod=~"$pod"}' },
  { label: '정규식 필터', query: 'up{job=~"${job:regex}"}' },
  { label: '시간 간격', query: 'rate(http_requests_total[$__interval])' },
] as const;

export const VariableReferencePanel: React.FC<VariableReferencePanelProps> = ({grafanaStyle = false}) => {
  const variables = grafanaStyle ? GRAFANA_VARIABLES : PONGO2_VARIABLES;
  const examples = grafanaStyle ? GRAFANA_EXAMPLES : PONGO2_EXAMPLES;

  return (
    <div className="mt-2 rounded border bg-slate-50 dark:bg-slate-900/80 p-3 space-y-3 text-xs">
      <Table className="text-xs">
        <TableHeader>
          <TableRow>
            <TableHead className="h-7 text-xs">변수</TableHead>
            <TableHead className="h-7 text-xs">설명</TableHead>
            <TableHead className="h-7 text-xs">예시 값</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {variables.map((v, i) => (
            <TableRow key={i}>
              <TableCell className="py-1 font-mono text-blue-600 dark:text-blue-400">{v.name}</TableCell>
              <TableCell className="py-1 text-muted-foreground">{v.desc}</TableCell>
              <TableCell className="py-1 font-mono text-slate-600 dark:text-slate-400">{v.example}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <div className="space-y-1 border-t pt-2">
        <p className="font-medium text-muted-foreground">쿼리 예제</p>
        {examples.map((ex) => (
          <div key={ex.label}>
            <span className="text-muted-foreground/70">{ex.label}: </span>
            <code className="font-mono text-slate-700 dark:text-slate-300 break-all">{ex.query}</code>
          </div>
        ))}
      </div>
    </div>
  );
};
