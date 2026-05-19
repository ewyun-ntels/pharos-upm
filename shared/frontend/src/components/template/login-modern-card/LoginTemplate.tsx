import type { LoginTemplateProps } from './types';

/**
 * 모던 카드 스타일 로그인 페이지 템플릿
 *
 * login-03 패턴처럼 상단 브랜드와 하단 폼을 단일 컬럼으로 구성합니다.
 * Slot 기반으로 모든 영역을 커스터마이징 가능합니다.
 *
 * @example 기본 사용
 * ```tsx
 * <LoginTemplate
 *   header={{
 *     logo: <img src="/logo.svg" alt="Logo" />,
 *     brandName: "Pharos",
 *   }}
 * >
 *   <LoginFormTemplate {...props} />
 * </LoginTemplate>
 * ```
 *
 * @example 완전 커스텀
 * ```tsx
 * <LoginTemplate
 *   headerSlot={<MyCustomHeader />}
 *   footerSlot={<MyCustomFooter />}
 * >
 *   <LoginFormTemplate {...props} />
 * </LoginTemplate>
 * ```
 */
export function LoginTemplate({
  headerSlot,
  footerSlot,
  header,
  children,
  backgroundColor = 'bg-[var(--site)]',
}: LoginTemplateProps) {
  return (
    <div
      className={`flex min-h-svh flex-col items-center justify-center gap-6 ${backgroundColor} bg-cover bg-no-repeat bg-center p-6 md:p-10`}
    >
      <div className="flex w-full max-w-[480px] flex-col gap-6">
        {headerSlot ? (
          headerSlot
        ) : (
          (header?.logo || header?.brandName) && (
            <a
              href={header?.linkHref || '#'}
              className="flex items-center gap-2 self-center font-medium text-foreground"
            >
              {header?.logo && (
                <div className="flex size-8 items-center justify-center rounded-xl bg-card shadow-sm">
                  {header.logo}
                </div>
              )}
              {header?.brandName && <span className="text-sm font-semibold text-foreground">{header.brandName}</span>}
            </a>
          )
        )}

        {children}

        {footerSlot}
      </div>
    </div>
  );
}
