/**
 * react-ace CJS interop adapter
 *
 * Vite 8 (Rolldown)은 CJS 모듈을 `export default require_lib()`로 감싸면서
 * `__esModule: true`를 인식하지 못해 exports 객체 전체를 default로 내보낸다.
 * 이 파일은 `.default`를 직접 꺼내는 처리를 한 곳에서 담당한다.
 *
 * Vite/Rolldown에서 CJS interop이 수정되면 이 파일만 제거하면 된다.
 */
import _AceEditor from 'react-ace';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export default ((_AceEditor as any).default ?? _AceEditor) as typeof _AceEditor;
