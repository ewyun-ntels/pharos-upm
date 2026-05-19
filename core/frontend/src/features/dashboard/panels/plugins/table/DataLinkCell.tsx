import React, {useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {Button} from '@pharos/shared/components/ui';
import {Popover, PopoverContent, PopoverTrigger} from '@pharos/shared/components/ui';
import {replaceVariables} from '@features/dashboard/utils/variable-parser';
import type {FilterMeta} from '@features/dashboard/hooks/slices/types';
import {DataLink} from './types';

interface DataLinkCellProps {
  value: unknown;
  row: Record<string, unknown>;
  links: DataLink[];
  filterMetas?: Map<string, FilterMeta>;
  cellContent: React.ReactNode;
}

/** {{value}}, {{row.컬럼명}} 치환 후 대시보드 변수도 치환 */
function resolveUrl(
  urlTemplate: string,
  value: unknown,
  row: Record<string, unknown>,
  filterMetas?: Map<string, FilterMeta>,
): string {
  let url = urlTemplate;

  url = url.replace(/\{\{\s*value\s*}}/g, encodeURIComponent(String(value ?? '')));
  // row.컬럼명: 공백 포함 컬럼명 지원 (e.g. {{row.Cell ID}})
  url = url.replace(/\{\{\s*row\.([^}]+?)\s*}}/g, (_, col: string) =>
    encodeURIComponent(String(row[col.trim()] ?? '')),
  );

  // 나머지 {{변수}}는 대시보드 변수로 치환 (없는 변수는 빈 문자열로)
  url = replaceVariables(url, filterMetas);

  // 치환되지 않고 남아있는 {{변수}} 제거 (대시보드 변수에 없는 경우)
  url = url.replace(/\{\{[^}]*}}/g, '');

  return url;
}

/**
 * react-router-dom의 BrowserRouter basename="/ui" 환경에서
 * navigate()는 basename을 자동으로 처리함.
 * - "/ui/..."로 시작하면 "/ui"를 제거하여 router 내부 경로로 변환
 * - "/"로 시작하지 않으면 절대 경로로 보정
 * - "http(s)://"로 시작하면 외부 URL로 처리
 */
function normalizeUrl(url: string): {internal: boolean; path: string} {
  if (/^https?:\/\//.test(url)) {
    return {internal: false, path: url};
  }

  let path = url;

  // /ui prefix 제거 (basename이 자동으로 추가하므로 중복 방지)
  if (path.startsWith('/ui/')) {
    path = path.slice(3); // "/ui" 제거
  } else if (path === '/ui') {
    path = '/';
  }

  // 상대 경로 방지: /로 시작하도록 보정
  if (!path.startsWith('/')) {
    path = '/' + path;
  }

  return {internal: true, path};
}

export function DataLinkCell({value, row, links, filterMetas, cellContent}: DataLinkCellProps) {
  const navigate = useNavigate();
  const [popoverOpen, setPopoverOpen] = useState(false);

  const openLink = (rawUrl: string, targetBlank: boolean) => {
    const {internal, path} = normalizeUrl(rawUrl);
    if (!internal || targetBlank) {
      window.open(internal ? window.location.origin + '/ui' + path : path, '_blank', 'noopener,noreferrer');
    } else {
      navigate(path);
    }
  };

  const handleClick = () => {
    if (links.length === 1) {
      const url = resolveUrl(links[0].url, value, row, filterMetas);
      openLink(url, links[0].targetBlank ?? false);
    }
  };

  if (links.length === 1) {
    return (
      <Button
        variant="link"
        className="h-auto p-0 w-full justify-start font-normal"
        title={resolveUrl(links[0].url, value, row, filterMetas)}
        onClick={handleClick}
      >
        {cellContent}
      </Button>
    );
  }

  return (
    <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="link"
          className="h-auto p-0 w-full justify-start font-normal"
          title={`${links.length}개 링크 (클릭하여 선택)`}
        >
          {cellContent}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-1" align="start">
        {links.map((link, i) => {
          const resolvedUrl = resolveUrl(link.url, value, row, filterMetas);
          return (
            <Button
              key={i}
              variant="ghost"
              className="w-full justify-start flex-col items-start h-auto py-2 px-3 gap-0.5"
              onClick={() => {
                setPopoverOpen(false);
                openLink(resolvedUrl, link.targetBlank ?? false);
              }}
            >
              <span className="font-medium text-sm">{link.title || `Link ${i + 1}`}</span>
              <span className="text-[11px] text-muted-foreground truncate max-w-[220px]">{resolvedUrl}</span>
            </Button>
          );
        })}
      </PopoverContent>
    </Popover>
  );
}
