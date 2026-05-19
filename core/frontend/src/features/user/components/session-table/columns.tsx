import React from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {Badge} from '@pharos/shared/components/ui';
import {formatRelativeTime} from '@lib/format-date';
import {UAParser} from 'ua-parser-js';

function parseUASummary(ua: string): string {
  if (!ua) return '';
  const result = new UAParser(ua).getResult();
  const browser = result.browser.name ?? '';
  const os = result.os.name ?? '';
  if (browser && os) return `${browser} / ${os}`;
  return browser || os;
}

export type Session = {
  username: string;
  client_id: string;
  client_ip?: string;
  user_agent?: string;
  blocked: boolean;
  created_at: string;
  expires_at: string;
};

export function buildSessionColumns(collectClientInfo: boolean): ColumnDef<Session>[] {
  const cols: ColumnDef<Session>[] = [
    {
      id: 'username',
      accessorKey: 'username',
      header: 'Username',
      cell: ({getValue}) => <span className="font-medium">{getValue() as string}</span>,
    },
    {
      id: 'blocked',
      ...(collectClientInfo ? {size: 100} : {meta: {flex: 1}}),
      accessorKey: 'blocked',
      header: 'Status',
      cell: ({getValue}) =>
        getValue() ? (
          <Badge variant="destructive">Blocked</Badge>
        ) : (
          <Badge variant="outline" className="text-green-600 border-green-600">
            Active
          </Badge>
        ),
    },
  ];

  if (collectClientInfo) {
    cols.push(
      {
        id: 'client_ip',
        accessorKey: 'client_ip',
        header: 'Client IP',
        size: 140,
        cell: ({getValue}) => {
          const v = getValue() as string | undefined;
          if (!v) return <span className="text-muted-foreground">-</span>;
          return <span className="font-mono text-xs">{v}</span>;
        },
      },
      {
        id: 'user_agent',
        accessorKey: 'user_agent',
        header: 'Client',
        meta: {flex: 1},
        cell: ({getValue}) => {
          const v = getValue() as string | undefined;
          if (!v) return <span className="text-muted-foreground">-</span>;
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

  cols.push(
    {
      id: 'created_at',
      accessorKey: 'created_at',
      header: 'Created at',
      size: 160,
      cell: ({getValue}) => {
        const v = getValue() as string;
        if (!v) return <span className="text-muted-foreground">-</span>;
        return (
          <Badge variant="outline" className="text-xs w-fit text-muted-foreground border-border">
            {formatRelativeTime(v)}
          </Badge>
        );
      },
    },
    {
      id: 'expires_at',
      accessorKey: 'expires_at',
      header: 'Expires at',
      size: 160,
      cell: ({getValue}) => {
        const v = getValue() as string;
        if (!v) return <span className="text-muted-foreground">-</span>;
        const isExpired = new Date(v) < new Date();
        return (
          <Badge
            variant="outline"
            className={`text-xs w-fit ${
              isExpired ? 'text-red-500 border-red-500' : 'text-muted-foreground border-border'
            }`}
          >
            {formatRelativeTime(v)}
          </Badge>
        );
      },
    },
  );

  return cols;
}

/** @deprecated Use buildSessionColumns(collectClientInfo) instead */
export const sessionColumns: ColumnDef<Session>[] = buildSessionColumns(false);
