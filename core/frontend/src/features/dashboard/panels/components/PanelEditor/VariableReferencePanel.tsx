import React from 'react';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@pharos/shared/components/ui';
import {
  TEMPLATE_VAR_START_TIME,
  TEMPLATE_VAR_END_TIME,
  QUERY_PARAM_DESCRIPTIONS,
} from '@lib/query-params';

const BUILTIN_VARIABLES = [
  QUERY_PARAM_DESCRIPTIONS['__start_time'],
  QUERY_PARAM_DESCRIPTIONS['__end_time'],
  QUERY_PARAM_DESCRIPTIONS['__step'],
  { name: '{{필터ID}}',       desc: '단일 선택 필터',      example: 'server1' },
  { name: '{{필터ID}}',       desc: '다중 선택 (IN 절용)', example: "'tag1','tag2'" },
] as const;

const QUERY_EXAMPLES = [
  {
    label: '시간 범위',
    query: `WHERE toDateTime(ts / 1000) BETWEEN toDateTime(${TEMPLATE_VAR_START_TIME} / 1000) AND toDateTime(${TEMPLATE_VAR_END_TIME} / 1000)`,
  },
  { label: '단일 선택 필터', query: "WHERE host = '{{Host}}'" },
  { label: '다중 선택 필터', query: "WHERE tag IN ({{tags}})" },
] as const;

export const VariableReferencePanel: React.FC = () => (
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
        {BUILTIN_VARIABLES.map((v, i) => (
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
      {QUERY_EXAMPLES.map((ex) => (
        <div key={ex.label}>
          <span className="text-muted-foreground/70">{ex.label}: </span>
          <code className="font-mono text-slate-700 dark:text-slate-300 break-all">{ex.query}</code>
        </div>
      ))}
    </div>
  </div>
);
