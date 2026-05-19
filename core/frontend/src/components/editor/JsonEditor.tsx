import {GenericObjectType} from '@rjsf/utils';
import React from 'react';
import {useTheme} from '@providers/theme-provider';
import {cn} from '@lib/utils';
import AceEditor from '@lib/ace-editor';
import 'ace-builds/src-noconflict/mode-json';
import 'ace-builds/src-noconflict/theme-tomorrow';
import 'ace-builds/src-noconflict/theme-chaos';
import 'ace-builds/src-noconflict/ext-language_tools';
import 'ace-builds/src-noconflict/ext-inline_autocomplete';
import 'ace-builds/src-noconflict/ext-searchbox';

export type JsonEditorProps = {
  width?: string;
  height?: string;
  setValue?: string;
  readOnly?: boolean;
  onChangeValue?: (value: string) => void;
  onChange?: (value: GenericObjectType) => void;
};

export function JsonEditor({width, height, setValue, readOnly, onChangeValue, onChange}: JsonEditorProps) {
  const [parsingError, setParsingError] = React.useState('');
  const {resolvedTheme} = useTheme();

  return (
    <div className={cn('w-full h-full')}>
      <AceEditor
        mode="json"
        width={width ?? '100%'}
        height={height}
        theme={resolvedTheme === 'dark' ? 'chaos' : 'tomorrow'}
        onChange={(e) => {
          onChangeValue?.(e);
          try {
            onChange?.(JSON.parse(e));
            setParsingError('');
          } catch (err) {
            setParsingError((err as Error).message);
          }
        }}
        showPrintMargin={false}
        enableBasicAutocompletion={!readOnly}
        showGutter={true}
        highlightActiveLine={!readOnly}
        readOnly={readOnly}
        value={setValue}
        wrapEnabled={false}
        setOptions={{
          useWorker: false,
          displayIndentGuides: true,
          enableLiveAutocompletion: true,
          enableSnippets: true,
          showLineNumbers: true,
          tabSize: 4,
        }}
      />
      {parsingError && <div className="text-red-500 text-xs px-2 py-1">{parsingError}</div>}
    </div>
  );
}
