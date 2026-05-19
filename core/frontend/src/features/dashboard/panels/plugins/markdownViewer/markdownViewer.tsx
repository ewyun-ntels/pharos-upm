import React, {useEffect, useRef, useState} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import hljs from 'highlight.js';
import 'highlight.js/styles/github-dark.css'; // 코드블록: dark테마를 기본으로 사용함
// import "highlight.js/styles/atom-one-light.css"; // 코드블록을 light테마로 사용하고 싶을때
import {Button} from '@pharos/shared/components/ui';
import {Check, Copy} from '@pharos/shared/components';
import type {PanelProps} from '@pharos/core/panel-registry';
import {MarkdownViewerPanelOptions} from './types';

/** markdown 설치
 * npm install react-markdown
 * npm install remark-gfm : 표/체크박스: GitHub-flavored markdown 지원
 * markdown 스타일은 직접 아래에 component에 직접 삽입. @tailwindcss/typography Plugin이 tailwindcss4에 설치되지 않기때문
 * npm install highlight.js : 코드블록 하이라이팅 적용
 */

type CodeProps = {
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
} & React.HTMLAttributes<HTMLElement>;

export const defaultMarkdown = `
  ## GFM 마크다운 예제
  - [x] remark-gfm 설치
  - [ ] 스타일 다듬기

  \`\`\`tsx
  <BadgeCustom variant="green">Normal</BadgeCustom>
  <BadgeCustom variant="blue">Succeeded</BadgeCustom>
  <BadgeCustom variant="red">Error</BadgeCustom>
  \`\`\`

  이건 \`인라인 코드\` 예제입니다.

  > Blockquite
  
  ## 📋 표 (Table)
  | 이름 | 역할 | 상태 |
  |------|------|------|
  | Dashboard | 📊 데이터 시각화 | ✅ 완료 |
  | Markdown  | 📄 문서 작성     | ⏳ 진행중 |
  
  ~~이 항목은 삭제되었습니다.~~
  
  [Refine 공식 문서](https://refine.dev)
  `;

export const MarkdownViewer = (props: PanelProps<MarkdownViewerPanelOptions>) => {
  const markdownRef = useRef<HTMLDivElement>(null);
  // 패널이 자신의 상태를 관리 - CardRenderer에 의존하지 않음
  const [content, setContent] = useState<string>(props.options?.content || defaultMarkdown);

  // props.options.content가 변경되면 업데이트 (대시보드에서 저장된 값 로드)
  useEffect(() => {
    if (props.options?.content) {
      setContent(props.options.content);
    }
  }, [props.options?.content]);

  const CodeBlock = ({inline, className = '', children, ...props}: CodeProps) => {
    const codeRef = useRef<HTMLElement>(null);
    const isBlock = !inline && /language-\w+/.test(className);
    const [copied, setCopied] = useState(false);

    // 코드 블록을 모두 하이라이트
    useEffect(() => {
      if (markdownRef.current) {
        const codeBlocks = markdownRef.current.querySelectorAll('pre code');
        codeBlocks.forEach((block) => {
          hljs.highlightElement(block as HTMLElement);
        });
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [content]);

    // 구형브라우저 대응, 브라우저 호환성, https 환경등 문제로 삽입함
    const fallbackCopy = (text: string) => {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.focus();
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
    };

    const handleCopy = async () => {
      const textToCopy = codeRef.current?.textContent || '';

      try {
        if (navigator.clipboard) {
          await navigator.clipboard.writeText(textToCopy);
        } else {
          fallbackCopy(textToCopy); // fallback
        }
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (err) {
        console.error('Copy failed:', err);
      }
    };

    if (isBlock) {
      return (
        <pre className="relative w-full bg-[#0d1117] border text-white p-0 my-4 rounded-md overflow-auto text-sm break-all">
          <code ref={codeRef} className={className} {...props}>
            {children}
          </code>
          <Button
            onClick={handleCopy}
            variant="outline"
            size="icon"
            className="absolute top-2 right-2 z-10 w-8 h-8"
          >
            {copied ? (
              <Check className="w-4 h-4 text-green-500" />
            ) : (
              <Copy className="w-4 h-4 text-foreground" />
            )}
          </Button>
        </pre>
      );
    }

    return (
      <code
        className="bg-gray-100 border font-mono text-[0.85em] px-[0.3em] py-[0.25em] rounded-md break-words"
        {...props}
      >
        {children}
      </code>
    );
  };

  return (
    <div
      ref={markdownRef}
      className="relative w-full h-full text-sm text-pretty p-3 overflow-auto border border-"
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({node, ...props}) => (
            <h1 className="text-2xl font-bold border-b pb-1 mt-6 mb-2" {...props} />
          ),
          h2: ({node, ...props}) => <h2 className="text-xl font-semibold mt-6 mb-2" {...props} />,
          h3: ({node, ...props}) => <h3 className="text-lg font-medium mt-5 mb-1" {...props} />,
          p: ({node, ...props}) => <p className="mt-2 mb-4 leading-relaxed" {...props} />,
          ul: ({node, ...props}) => <ul className="list-disc pl-6 space-y-1 my-3" {...props} />,
          ol: ({node, ...props}) => <ol className="list-decimal pl-6 space-y-1 my-3" {...props} />,
          li: ({node, ...props}) => <li className="mt-1 leading-snug" {...props} />,
          a: ({node, ...props}) => (
            <a className="text-blue-600 underline hover:text-blue-800" {...props} />
          ),
          code: CodeBlock,
          table: ({node, ...props}) => (
            <table className="w-full border-collapse table-auto text-left my-4" {...props} />
          ),
          thead: ({node, ...props}) => (
            <thead className="bg-accent border-b text-sm font-semibold" {...props} />
          ),
          th: ({node, ...props}) => <th className="px-3 py-2 border" {...props} />,
          td: ({node, ...props}) => <td className="px-3 py-2 border" {...props} />,
          input: ({node, ...props}) => (
            <input
              type="checkbox"
              className="rounded-sm mr-2 accent-blue-500"
              disabled
              {...props}
            />
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
