import { cn } from '../../lib';

type Variant = 'default' | 'h2' | 'h3' | 'h4' | 'formLabel';

const VARIANT_CONFIG: Record<Variant, { tag: React.ElementType; class: string }> = {
  default:      { tag: 'h2',    class: 'text-lg font-medium' }, // 페이지 헤더, 탭·선택 템플릿 제목
  h2:           { tag: 'h2',    class: 'text-lg font-medium' },
  h3:           { tag: 'h3',    class: 'text-base font-medium pb-3' }, // 섹션, 그룹 제목
  h4:           { tag: 'h4',    class: 'text-sm font-medium' },  // 소제목, Sheet 제목
  formLabel:    { tag: 'label', class: 'block text-sm font-medium pb-1' }, // TODO : base design system 작업 시 수정 예정
};

interface TitleProps extends React.HTMLAttributes<HTMLElement> {
  variant?: Variant;
  htmlFor?: string;
}

export function Title({ className, children, variant = 'default', ...props }: TitleProps) {
  const { tag: Tag, class: variantClass } = VARIANT_CONFIG[variant];

  return (
    <Tag className={cn('title', variantClass, className)} {...props}>
      {children}
    </Tag>
  );
}