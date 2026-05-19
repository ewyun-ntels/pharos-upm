// TODO 로그뷰 컴포넌트로 차후 수정시 참고용으로 남겨둠
// import React, {useEffect, useRef} from 'react';
// import 'ace-builds';
// import 'ace-builds/src-noconflict/mode-yaml';
// import 'ace-builds/src-noconflict/theme-chaos';
// import 'ace-builds/src-noconflict/theme-tomorrow';
// import 'ace-builds/src-noconflict/ext-inline_autocomplete';
// import {Badge} from '@pharos/shared/components/ui';
// import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@pharos/shared/components/ui';
// import {useShow} from '@/lib/data-provider';
// import {Log, LogResponse} from '@providers/kubernetes-provder/pods';
// import {IconButton} from '@pharos/shared/components/ui-extension';
// import {Download} from '@pharos/shared/components';
// import AceEditor from 'react-ace';
// import {useTheme} from 'next-themes';
//
// const heightPaddingSize = 28;
// const headerHeight = 12;
//
// const exportTxt = (fileName: string, output: string) => {
//   const element = document.createElement('a');
//   const file = new Blob([output], {
//     type: 'text/plain',
//   });
//   element.href = URL.createObjectURL(file);
//   element.download = fileName;
//   document.body.appendChild(element); // FireFox
//   element.click();
// };
//
// interface LogViewProps {
//   id: string;
//   height: number;
//   width: string;
//   data: Log;
//   selectedContainer: (tabId: string, container: string) => void;
// }
//
// const LogView: React.FC<LogViewProps> = ({id, height, width, data, selectedContainer}) => {
//   const editorHeight = height - headerHeight;
//   const {theme} = useTheme();
//   const editorRef = useRef<any>(null);
//
//   const {query} = useShow({
//     dataProviderName: 'kubernetesProvider',
//     resource: 'log',
//     id: 1,
//     queryOptions: {
//       refetchInterval: 5000,
//       structuralSharing: true,
//     },
//     meta: {
//       variables: {
//         value: data,
//       },
//     },
//   });
//
//   let viewData: string;
//   if (query.isLoading) {
//     viewData = 'Loading...';
//   } else if (query.isError) {
//     viewData = 'Error...';
//   } else {
//     viewData = LogResponse.parse(query.data).resp;
//   }
//
//   useEffect(() => {
//     const editor = editorRef.current.editor;
//     editor.gotoLine(viewData.length, 0, true);
//   }, [viewData.length]);
//
//   return (
//     <div className="bg-secondary/30 dark:bg-secondary/60">
//       <div className="bg-site h-10 flex items-center justify-between px-6">
//         <div className={'flex items-center h-8 space-x-3'}>
//           <h4 className={'text-muted-foreground text-sm'}>Namespace</h4>
//           <Badge variant={'secondary'}>default</Badge>
//           <h4 className={'text-muted-foreground text-sm'}>Pod</h4>
//           <Badge variant={'secondary'}>{data.pod}</Badge>
//           <h4 className={'text-muted-foreground text-sm'}>Logs from</h4>
//           <h4 className={'text-muted-foreground text-sm font-bold'}>
//             {new Date(query.dataUpdatedAt).toLocaleTimeString()}
//           </h4>
//           {/* TODO 로그보는데 불편할 경우 추가 필요 */}
//           {/*<h4 className="text-muted-foreground text-sm">*/}
//           {/*    Auto reload*/}
//           {/*</h4>*/}
//           {/*<Checkbox id="auto"/>*/}
//         </div>
//
//         <div className={'flex items-center h-8 space-x-3'}>
//           <h4 className={'text-muted-foreground text-sm'}>Container</h4>
//           <Select onValueChange={(value) => selectedContainer(id, value)}>
//             <SelectTrigger className="h-8 w-48">
//               <SelectValue defaultValue={data.selectContainer} placeholder={data.selectContainer} />
//             </SelectTrigger>
//             <SelectContent>
//               {data.containers.map((item, index) => (
//                 <SelectItem key={index} value={item}>
//                   {item}
//                 </SelectItem>
//               ))}
//             </SelectContent>
//           </Select>
//           <IconButton
//             icon={<Download />}
//             onClick={() => {
//               const now = new Date();
//               const formattedDate = now
//                 .toLocaleString('ko-KR', {
//                   year: 'numeric',
//                   month: '2-digit',
//                   day: '2-digit',
//                   hour: '2-digit',
//                   minute: '2-digit',
//                   second: '2-digit',
//                   hour12: false,
//                 })
//                 .replace(/\. /g, '-')
//                 .replace(/, /g, 'T')
//                 .replace(/:/g, '_');
//
//               exportTxt(`${data.pod}-${formattedDate}.log`, viewData);
//             }}
//             variant="outline"
//             size="sm"
//           >
//             Export
//           </IconButton>
//         </div>
//       </div>
//       <div style={{height: editorHeight - heightPaddingSize}} className="bg-gray-800 dark:bg-black">
//         <AceEditor
//           ref={editorRef}
//           height={(editorHeight - heightPaddingSize).toString() + 'px'}
//           width={width}
//           theme={theme === 'dark' ? 'chaos' : 'tomorrow'}
//           fontSize={12}
//           showPrintMargin={false}
//           enableBasicAutocompletion={true}
//           showGutter={true}
//           highlightActiveLine={true}
//           value={viewData}
//           wrapEnabled={false}
//           setOptions={{
//             useWorker: false,
//             enableSnippets: false,
//             showLineNumbers: false,
//             tabSize: 4,
//             readOnly: true,
//           }}
//         />
//       </div>
//     </div>
//   );
// };
//
// export {LogView};
