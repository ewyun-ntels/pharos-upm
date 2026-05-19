// TODO KubernetesProvider를 사용하는 Editor 컴포넌트로 차후 수정 필요
// import {Button} from '@pharos/shared/components/ui';
// import React, {useEffect, useRef} from 'react';
// import AceEditor from 'react-ace';
// import 'ace-builds';
// import {useTheme} from 'next-themes';
// import 'ace-builds/src-noconflict/mode-yaml';
// import 'ace-builds/src-noconflict/theme-tomorrow';
// import 'ace-builds/src-noconflict/theme-chaos';
// import 'ace-builds/src-noconflict/ext-inline_autocomplete';
// import 'ace-builds/src-noconflict/ext-searchbox';
// import {Template} from './template';
// import {useCreate} from '@/lib/data-provider';
// import {useToast} from '@pharos/shared/components/ui';
// import {ManifestErrorResponse, ManifestResponse} from '@providers/kubernetes-provder/manifest';
// import {Spinner} from '@pharos/shared/components/ui-extension';
// import {TabEditorData, TabType} from '@components/controlBar/controlTabStore';
//
// const heightPaddingSize = 28;
//
// interface EditorProps {
//   id: string;
//   tabType: TabType;
//   language: string;
//   height: number;
//   width: string;
//   line?: number;
//   column?: number;
//   data: string;
//   onCloseHandler?: (id: string) => void;
//   onSaveData: (id: string, data: TabEditorData) => void;
//   onSaveLine: (id: string, line: number, column: number) => void;
// }
//
// const Editor: React.FC<EditorProps> = ({
//   id,
//   tabType,
//   language,
//   height,
//   width,
//   line,
//   column,
//   data,
//   onCloseHandler,
//   onSaveData,
//   onSaveLine,
// }) => {
//   const headerHeight = 12;
//   const {theme} = useTheme();
//   const editorHeight = height - headerHeight;
//   const createButtonName = tabType === 'editor' ? 'Create' : 'Save';
//
//   const editorRef = useRef<any>(null);
//   const [editData, setEditData] = React.useState<string>(data);
//   const [loading, setLoading] = React.useState<boolean>(false);
//
//   const {mutate} = useCreate();
//   const {toast} = useToast();
//
//   useEffect(() => {
//     if (editorRef.current) {
//       const editor = editorRef.current.editor;
//       editor.gotoLine(line, column, true);
//       editor.focus();
//     }
//     // eslint-disable-next-line react-hooks/exhaustive-deps
//   }, []);
//
//   const saveHandler = (close: boolean) => {
//     setLoading(true);
//     mutate(
//       {
//         dataProviderName: 'kubernetesProvider',
//         resource: 'manifest',
//         values: editData,
//       },
//       {
//         onSuccess: (data) => {
//           const resp = ManifestResponse.parse(data);
//           setLoading(false);
//           toast({
//             title: 'Create Success',
//             description: resp.resp.message,
//           });
//           if (close) {
//             onCloseHandler && onCloseHandler(id);
//           }
//         },
//         onError: (err) => {
//           const resp = ManifestErrorResponse.parse(err);
//           setLoading(false);
//           toast({
//             variant: 'destructive',
//             title: resp.response.statusText,
//             description: resp.response.data.error,
//           });
//         },
//       },
//     );
//   };
//
//   return (
//     <div className="bg-site">
//       <div className="flex justify-between p-1 px-6 border-b">
//         {/* TODO justify-bitween 설정으로 인하여 editor가 아닐 <button>을 오른쪽에 설정하기 위해 <div></div>를 사용, 추후 좋은 방법이 있으면 수정 필요 */}
//         {tabType === 'editor' ? (
//           <Template
//             onSelect={(data) => {
//               if (data) {
//                 onSaveData(id, {data: data, showProps: undefined});
//                 setEditData(data);
//               }
//             }}
//           />
//         ) : (
//           <div></div>
//         )}
//         <div className={'flex space-x-3'}>
//           {loading && <Spinner className={'h-6 w-6'} />}
//           <Button className="h-8" variant="default" onClick={() => saveHandler(false)}>
//             {createButtonName}
//           </Button>
//           {tabType === 'modifyEditor' && (
//             <Button className="h-8" variant="default" onClick={() => saveHandler(true)}>
//               Save & Close
//             </Button>
//           )}
//         </div>
//       </div>
//       <div style={{height: editorHeight - heightPaddingSize}} className="bg-gray-800 dark:bg-black">
//         <AceEditor
//           ref={editorRef}
//           mode={language}
//           height={(editorHeight - heightPaddingSize).toString() + 'px'}
//           width={width}
//           theme={theme === 'dark' ? 'chaos' : 'tomorrow'}
//           onChange={(e) => {
//             onSaveData(id, {data: e, showProps: undefined});
//             setEditData(e);
//           }}
//           onCursorChange={(e) => {
//             onSaveLine(id, e.cursor.row + 1, e.cursor.column);
//           }}
//           showPrintMargin={false}
//           enableBasicAutocompletion={true}
//           showGutter={true}
//           highlightActiveLine={true}
//           value={editData}
//           wrapEnabled={false}
//           setOptions={{
//             useWorker: false,
//             displayIndentGuides: true,
//             enableLiveAutocompletion: true,
//             enableSnippets: true,
//             showLineNumbers: true,
//             tabSize: 4,
//           }}
//         />
//       </div>
//     </div>
//   );
// };
//
// export {Editor};
