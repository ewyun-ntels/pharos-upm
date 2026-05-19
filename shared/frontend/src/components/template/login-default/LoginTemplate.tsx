import type { LoginTemplateProps } from './types';

/**
 * 기본 로그인 페이지 템플릿
 *
 * Slot 기반으로 모든 영역을 커스터마이징 가능
 *
 * @example 기본 사용
 * ```tsx
 * <LoginTemplate
 *   branding={{
 *     logo: <img src="/logo.svg" alt="Logo" />,
 *     copyright: "Copyright ⓒ Company",
 *   }}
 * >
 *   <LoginForm />
 * </LoginTemplate>
 * ```
 *
 * @example 완전 커스텀
 * ```tsx
 * <LoginTemplate
 *   brandingSlot={<MyCustomBranding />}
 *   footerSlot={<MyCustomFooter />}
 * >
 *   <LoginForm />
 * </LoginTemplate>
 * ```
 */
export function LoginTemplate({
  brandingSlot,
  footerSlot,
  branding,
  children,
}: LoginTemplateProps) {
  const backgroundClassName =
    branding?.backgroundClassName ||
    'bg-primary bg-[linear-gradient(to_bottom,rgba(0,0,0,0),rgba(0,0,0,0.7))]';

  return (
    <div className="min-h-screen w-full h-full leading-normal flex flex-col md:flex-row">
      {/* 좌측 브랜딩 영역 */}
      {brandingSlot ? (
        brandingSlot
      ) : (
        <div
          className={`w-full md:w-1/2 flex items-center ${backgroundClassName} bg-cover bg-no-repeat bg-center`}
        >
          <div className="relative flex flex-col items-center w-60 h-24 mx-auto my-7">
            {branding?.logo}
          </div>
        </div>
      )}

      {/* 우측 액션 영역 */}
      <div className="w-full md:w-1/2 bg-background flex flex-col justify-center items-center">
        {children}
      </div>

      {/* 푸터 영역 */}
      {footerSlot ? (
        footerSlot
      ) : (
        branding?.copyright && (
          <div className="items-center absolute bottom-5 left-1/4 transform -translate-x-1/2 text-xs text-gray-300">
            {branding.copyright}
          </div>
        )
      )}
    </div>
  );
}
