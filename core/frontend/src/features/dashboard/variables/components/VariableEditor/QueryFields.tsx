'use client';

import React, { useEffect, useState } from 'react';
import { useDatasourceList } from '@hooks/use-datasource-list';
import AceEditor from '@lib/ace-editor';
import 'ace-builds/src-noconflict/mode-text';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/theme-monokai';
import { useTheme } from '@providers/theme-provider';
import 'ace-builds/src-noconflict/ext-language_tools';
import { Label } from '@pharos/shared/components/ui';
import { Input } from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

/**
 * Query 유효성 검증 결과
 */
interface QueryValidationResult {
  isValid: boolean;
  message: string;
}

/**
 * Query 유효성 검증
 * data-driven 타입 (select with isMulti option)의 query 검증
 */
const validateQuery = (query: string): QueryValidationResult => {
  if (!query.trim()) {
    return {isValid: true, message: ''};
  }

  const queryLower = query.toLowerCase().trim();

  if (!queryLower.startsWith('select')) {
    return {
      isValid: false,
      message: 'Query must be a SELECT statement',
    };
  }

  const hasValue =
    queryLower.includes('as value') ||
    queryLower.includes(' value,') ||
    queryLower.includes(' value ');
  const hasLabel =
    queryLower.includes('as label') ||
    queryLower.includes(' label,') ||
    queryLower.includes(' label ');

  if (!hasValue || !hasLabel) {
    return {
      isValid: false,
      message:
        'Query must include "value" and "label" columns (e.g., "column_name AS value, column_name AS label")',
    };
  }

  return {isValid: true, message: ''};
};

export interface QueryFieldsProps {
  data: FilterConfig;
  onChange: React.Dispatch<React.SetStateAction<FilterConfig>>;
}

export function QueryFields({ data, onChange }: QueryFieldsProps) {
  const { resolvedTheme } = useTheme();
  const [queryValidation, setQueryValidation] = useState<QueryValidationResult>({
    isValid: true,
    message: '',
  });

  const { dataSourceList } = useDatasourceList();

  const handleChange = (field: keyof FilterConfig, value: string | boolean) => {
    if (field === 'query' && typeof value === 'string') {
      const validation = validateQuery(value);
      setQueryValidation(validation);
    }

    onChange((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  useEffect(() => {
    if (data.query) {
      const validation = validateQuery(data.query);
      setQueryValidation(validation);
    }
  }, [data.query]);

  return (
    <div className="space-y-4 mt-4">
      <div>
        <Label htmlFor="dataSource" className="text-sm font-normal">
          Data Source
        </Label>
        <Select
          value={data.datasourceName || ''}
          onValueChange={(value) => handleChange('datasourceName', value)}
        >
          <SelectTrigger className="mt-2 w-60">
            <SelectValue placeholder="Select data source">
              {data.datasourceName || 'Select data source'}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            {/* Show current value even if not in registered list */}
            {data.datasourceName && 
             !dataSourceList.some(ds => ds.value === data.datasourceName) && (
              <SelectItem key={data.datasourceName} value={data.datasourceName}>
                {data.datasourceName} (Not registered)
              </SelectItem>
            )}
            {dataSourceList.map((source) => (
              <SelectItem key={source.value} value={source.value}>
                {source.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label htmlFor="label" className="text-sm font-normal">
          Label
        </Label>
        <p className="text-xs text-muted-foreground mt-1"> Optional display name</p>
        <Input
          id="label"
          value={data.label || ''}
          onChange={(e) => handleChange('label', e.target.value)}
          placeholder="Enter label"
          className="mt-2 w-60"
        />
      </div>

      <div>
        <Label htmlFor="query" className="text-sm font-normal">
          Query <span className="text-red-500">*</span>
        </Label>
        <div className="text-xs text-muted-foreground mt-1 mb-2 space-y-2">
          <p>Write a SELECT query that returns &quot;value&quot; and &quot;label&quot; columns:</p>
          <div className="pl-2 space-y-1">
            <p>
              • <strong>value</strong>: The actual data value used for filtering (stored internally)
            </p>
            <p>
              • <strong>label</strong>: The human-readable text displayed to users in the dropdown
            </p>
          </div>
          <p>
            Example:{' '}
            <code className="bg-input px-1 rounded">
              SELECT system_id AS value, system_name AS label FROM table_name ORDER BY system_name
            </code>
          </p>
          <p className="text-muted-foreground">
            This will show &quot;system_name&quot; to users but filter by &quot;system_id&quot; when
            selected.
          </p>
        </div>
        <div
          className={`max-w-3xl mt-2 border rounded-md overflow-hidden ${!queryValidation.isValid ? 'border-red-500' : 'border-border'}`}
        >
          <AceEditor
            mode="text"
            theme={resolvedTheme === 'dark' ? 'monokai' : 'github'}
            name="variable-query-editor"
            onChange={(value) => handleChange('query', value)}
            value={data.query || ''}
            editorProps={{ $blockScrolling: true }}
            setOptions={{
              enableBasicAutocompletion: false,
              enableLiveAutocompletion: false,
              enableSnippets: false,
            }}
            width="100%"
            height="120px"
            placeholder="SELECT column_name AS value, column_name AS label FROM table_name"
            // showPrintMargin={false}
          />
        </div>
        {!queryValidation.isValid && (
          <p className="text-xs text-red-500 mt-1">{queryValidation.message}</p>
        )}
      </div>
    </div>
  );
}
