import React, {useState} from 'react';
import {Textarea} from '@pharos/shared/components/ui';
import type {PanelProps} from '@pharos/core/panel-registry';
import {MarkdownViewerPanelOptions} from './types';

type MarkdownEditorProps = {
  value: string;
  onChange: (value: string) => void;
};

export const MarkdownEditor = ({value, onChange}: MarkdownEditorProps) => {
  return (
    <div className="h-full w-full p-4">
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full h-full inset-shadow-sm bg-gray-100 resize-none"
        placeholder="마크다운 입력"
      />
    </div>
  );
};

/**
 * MarkdownEditor Card Wrapper for Panel System
 * Wraps MarkdownEditor with PanelProps interface
 */
export function MarkdownEditorCard(props: PanelProps<MarkdownViewerPanelOptions>) {
  const [content, setContent] = useState(props.options?.content || '');

  return <MarkdownEditor value={content} onChange={setContent} />;
}
