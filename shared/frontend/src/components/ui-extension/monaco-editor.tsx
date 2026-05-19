// 현재 사용하지 않음
// "@monaco-editor/react": "^4.6.0" 설치 필요
// import React, {useEffect, useRef} from 'react';
// import Editor, {EditorProps, Monaco} from '@monaco-editor/react';
// import {useTheme} from "next-themes";
//
// interface MonacoEditorProps extends EditorProps {
//     restore?: any
//     editorOnChange?: (e: any) => void
//     onSave?: (data: any) => void;
// }
//
// const MonacoEditor: React.FC<MonacoEditorProps> = ({options, restore, editorOnChange, onSave, ...props}) => {
//     const editorRef = useRef<any>(null);
//     const handleEditorDidMount = (editor: any, monaco: Monaco) => {
//         editorRef.current = editor;
//         restore && editorRef.current.restoreViewState(restore)
//     };
//
//     const {theme} = useTheme();
//
//     return (
//         <Editor
//             // className={"bg-black"}
//             // height={height}
//             // width={width}
//             defaultLanguage={"yaml"}
//             // defaultValue="name: Monaco Editor\nlanguage: YAML"
//             options={options ? options : {
//                 fontSize: 13,
//                 theme: theme === "dark" ? "vs-dark" : "light",
//                 minimap: {enabled: true},
//                 scrollbar: {
//                     vertical: 'auto',
//                     horizontal: 'auto'
//                 },
//                 cursorSmoothCaretAnimation: 'on',
//                 automaticLayout: true,
//                 formatOnPaste: true,
//                 formatOnType: true,
//                 wordWrap: 'on'
//             }}
//             // beforeMount={(e)=> {
//             //     console.log("beforeMount");
//             //     console.log(e)}}
//             onMount={handleEditorDidMount}
//             {...props}
//             // onChange={handleEditorChange}
//             // onMount={handleEditorDidMount}
//         />
//     );
// };
//
// export type {MonacoEditorProps};
// export default MonacoEditor;
