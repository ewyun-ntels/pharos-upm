import React from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {Badge, Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from '@pharos/shared/components/ui';
import {formatDateTime} from '@lib/format-date';
import {UAParser} from 'ua-parser-js';

export type LoginHistoryRecord = {
  session_id: string;
  user_id: string;
  client_ip: string;
  user_agent: string;
  started_at: string;
  ended_at: string | null;
  end_reason: string;
};

function EndReasonBadge({endedAt, endReason}: {endedAt: string | null; endReason: string}) {
  if (!endedAt && !endReason) {
    return <Badge variant="outline" className="text-green-600 border-green-600">Active</Badge>;
  }
  if (endReason === 'logout') {
    return <Badge variant="outline" className="text-orange-500 border-orange-500">Logout</Badge>;
  }
  if (endReason === 'expired') {
    return <Badge variant="outline" className="text-muted-foreground">Expired</Badge>;
  }
  if (endReason === 'revoked') {
    return <Badge variant="outline" className="text-blue-500 border-blue-500">Revoked</Badge>;
  }
  if (endReason.startsWith('failed:')) {
    const desc = endReason.slice('failed:'.length).trim();
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge variant="outline" className="text-red-500 border-red-500 cursor-help">Failed</Badge>
          </TooltipTrigger>
          {desc && <TooltipContent side="top" className="max-w-xs text-xs">{desc}</TooltipContent>}
        </Tooltip>
      </TooltipProvider>
    );
  }
  return <Badge variant="outline">{endReason}</Badge>;
}

function parseUASummary(ua: string): string {
  if (!ua) return '';
  const result = new UAParser(ua).getResult();
  const browser = result.browser.name ?? '';
  const os = result.os.name ?? '';
  if (browser && os) return `${browser} / ${os}`;
  return browser || os;
}

function formatDuration(startedAt: string, endedAt: string | null): string {
  const end = endedAt ? new Date(endedAt) : new Date();
  const diffMs = end.getTime() - new Date(startedAt).getTime();
  if (diffMs < 0) return '-';

  const totalSec = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSec / 86400);
  const hours = Math.floor((totalSec % 86400) / 3600);
  const minutes = Math.floor((totalSec % 3600) / 60);
  const seconds = totalSec % 60;

  if (days > 0) return `${days}d ${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}

export function buildHistoryColumns(collectClientInfo: boolean): ColumnDef<LoginHistoryRecord>[] {
  const cols: ColumnDef<LoginHistoryRecord>[] = [
    {
      id: 'started_at',
    accessorKey: 'started_at',
    header: 'Started at',
    size: 200,
    cell: ({getValue}) => {
      const v = getValue() as string;
      return formatDateTime(v);
    },
  },
  {
    id: 'ended_at',
    accessorKey: 'ended_at',
    header: 'Ended at',
    size: 200,
    cell: ({getValue}) => {
      const v = getValue() as string | null;
      return v ? formatDateTime(v) : <span className="text-muted-foreground">-</span>;
    },
  },
  {
    id: 'duration',
    header: 'Duration',
    size: 112,
    cell: ({row}) => {
      const {started_at, ended_at} = row.original;
      const label = formatDuration(started_at, ended_at);
      if (!ended_at) {
        return <span className="text-green-600 font-medium">{label}</span>;
      }
      return <span className="text-muted-foreground">{label}</span>;
    },
  },
  {
    id: 'user_id',
    accessorKey: 'user_id',
    ...(collectClientInfo ? {size: 100} : {meta: {flex: 1}}),
    header: 'Username',
    cell: ({getValue}) => <span className="font-medium">{getValue() as string}</span>,
  },
    {
      id: 'event',
      header: 'Event',
      size: 120,
      cell: ({row}) => (
        <EndReasonBadge endedAt={row.original.ended_at} endReason={row.original.end_reason} />
      ),
    },
  ];

  if (collectClientInfo) {
    cols.splice(
      4,
      0,
      {
        id: 'client_ip',
        accessorKey: 'client_ip',
        header: 'Client IP',
        size: 140,
        cell: ({getValue}) => {
          const v = getValue() as string;
          return v ? (
            <span className="font-mono text-xs">{v}</span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
      {
        id: 'user_agent',
        accessorKey: 'user_agent',
        header: 'Client',
        meta: {flex: 1},
        cell: ({getValue}) => {
          const v = getValue() as string;
          const summary = parseUASummary(v);
          return summary ? (
            <span className="text-xs text-muted-foreground">{summary}</span>
          ) : (
            <span className="text-muted-foreground">-</span>
          );
        },
      },
    );
  }

  return cols;
}
