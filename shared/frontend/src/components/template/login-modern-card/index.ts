/**
 * 모던 카드 스타일 로그인 템플릿
 *
 * 카드 기반의 깔끔하고 모던한 디자인으로 로그인 페이지를 제공합니다.
 * 중앙 정렬 단일 열 레이아웃을 사용하며, Slot 기반으로 완전히 커스터마이징 가능합니다.
 *
 * @example 기본 사용
 * ```tsx
 * import {
 *   LoginTemplate,
 *   LoginFormTemplate,
 * } from '@pharos/shared/components/template/login-modern-card';
 *
 * <LoginTemplate
 *   header={{
 *     logo: <img src="/logo.svg" alt="Logo" />,
 *     title: "Welcome Back",
 *     subtitle: "Sign in to your account",
 *   }}
 * >
 *   <LoginFormTemplate
 *     onSubmit={handleLogin}
 *     errorMessage={error}
 *     labels={loginLabels}
 *   />
 * </LoginTemplate>
 * ```
 *
 * @example 비밀번호 변경
 * ```tsx
 * <LoginTemplate
 *   header={{
 *     title: "Secure Your Account",
 *   }}
 * >
 *   <PasswordChangeFormTemplate
 *     onSubmit={handleChangePassword}
 *     errorMessage={error}
 *   />
 * </LoginTemplate>
 * ```
 */

export { LoginTemplate } from './LoginTemplate';
export { LoginFormTemplate } from './LoginFormTemplate';
export { PasswordChangeFormTemplate } from './PasswordChangeFormTemplate';
export type {
  LoginTemplateProps,
  LoginFormLabels,
  LoginFormTemplateProps,
  PasswordChangeFormLabels,
  PasswordChangeFormTemplateProps,
} from './types';
