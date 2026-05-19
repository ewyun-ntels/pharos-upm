'use client';

import React, {useState, useEffect} from 'react';
import {Button} from '@pharos/shared/components/ui';
import {variablePluginRegistry} from '@features/dashboard/variables/registry/VariablePluginRegistry';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

interface VariablePreviewProps {
  data: FilterConfig;
  dashboardId?: string;
}

export function VariablePreview({data, dashboardId}: VariablePreviewProps) {
  const [showPreview, setShowPreview] = useState(false);
  const [previewKey, setPreviewKey] = useState(0);

  // 타입 변경 시 미리보기 숨김
  useEffect(() => {
    setShowPreview(false);
  }, [data.type]);

  const generatePreviewComponent = () => {
    // Registry에서 Preview 컴포넌트 가져오기
    const PreviewComponent = variablePluginRegistry.getPreview(data.type);

    if (!PreviewComponent) {
      // Preview가 없으면 기본 메시지 표시
      return (
        <div className="p-4 text-center text-muted-foreground">
          Preview not available for this variable type.
        </div>
      );
    }

    // Preview 컴포넌트 렌더링
    return <PreviewComponent data={data} dashboardId={dashboardId} />;
  };

  return (
    <div className="space-y-4 w-1/4">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold">Preview</div>
        <Button
          variant="secondary"
          className="h-8 shadow-none px-3 py-1"
          onClick={() => {
            setPreviewKey((prev) => prev + 1);
            setShowPreview(true);
          }}
        >
          Run
        </Button>
      </div>

      {showPreview && (
        <div className="min-h-[80px] flex items-center justify-center border border-border rounded-md bg-muted/10">
          <div key={previewKey}>{generatePreviewComponent()}</div>
        </div>
      )}

      {!showPreview && (
        <div className="border border-dashed border-border rounded-md bg-muted/10 min-h-[80px] flex items-center justify-center text-muted-foreground">
          Click &quot;Run&quot; to preview.
        </div>
      )}
    </div>
  );
}

export default VariablePreview;
