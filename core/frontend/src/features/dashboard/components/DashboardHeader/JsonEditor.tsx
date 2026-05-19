import {useState, useEffect} from 'react';
import AceEditor from '@lib/ace-editor';
import 'ace-builds/src-noconflict/mode-json';
import 'ace-builds/src-noconflict/theme-tomorrow';
import 'ace-builds/src-noconflict/theme-chaos';
import 'ace-builds/src-noconflict/ext-language_tools';
import {useTheme} from '@providers/theme-provider';

interface JsonEditorProps {
  initialData?: string | object;
  onChange: (value: string) => void;
}

export default function JsonEditor({initialData, onChange}: JsonEditorProps) {
  const {resolvedTheme} = useTheme();
  const [jsonData, setJsonData] = useState('');
  const [isValid, setIsValid] = useState(true);

  useEffect(() => {
    if (initialData === undefined) {
      setJsonData('');
    } else if (typeof initialData === 'string') {
      setJsonData(initialData);
    } else {
      setJsonData(JSON.stringify(initialData, null, 2));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex flex-col w-full mx-auto">
      <div className="w-full">
        <AceEditor
          value={jsonData}
          onChange={(value) => {
            setJsonData(value);
            try {
              JSON.parse(value);
              setIsValid(true);
              onChange(value);
            } catch (err) {
              setIsValid(false);
              onChange(value);
            }
          }}
          setOptions={{
            useWorker: false,
            displayIndentGuides: true,
            enableLiveAutocompletion: false,
            enableSnippets: false,
            showLineNumbers: true,
            tabSize: 4,
            foldStyle: 'markbegin',
            showFoldWidgets: true,
          }}
          width="100%"
          mode="json"
          theme={resolvedTheme === 'dark' ? 'chaos' : 'tomorrow'}
          className="h-120 border border-border resize-none focus:outline-none rounded-[var(--radius)] overflow-hidden"
          enableBasicAutocompletion={false}
          wrapEnabled={true}
          showGutter={true}
          highlightActiveLine={true}
          showPrintMargin={false}
        />
      </div>
      <div className="flex justify-between items-center">
        <div className="text-sm pt-2">
          {!isValid && jsonData && <span className="text-red-600">유효하지 않은 JSON입니다.</span>}
        </div>
      </div>
    </div>
  );
}
