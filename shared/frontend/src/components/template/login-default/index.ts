/**
 * Login Default Template
 *
 * 기본 로그인 페이지 템플릿 컴포넌트들
 *
 * @example
 * ```tsx
 * import {
 *   LoginTemplate,
 *   LoginFormTemplate,
 * } from '@pharos/shared/components/template/login-default';
 *
 * function LoginPage({ context, onComplete }: LoginActionProps) {
 *   return (
 *     <LoginTemplate branding={{
 *       logoUrl: '/ui-config/images/logo.svg',
 *       copyright: 'Copyright ⓒ NTELS',
 *     }}>
 *       <LoginFormTemplate
 *         onSubmit={handleLogin}
 *         t={context.utils.t}
 *       />
 *     </LoginTemplate>
 *   );
 * }
 * ```
 */

export { LoginTemplate } from './LoginTemplate';
export { LoginFormTemplate } from './LoginFormTemplate';
export { PasswordChangeFormTemplate } from './PasswordChangeFormTemplate';
export type {
  LoginTemplateProps,
  LoginFormTemplateProps,
  LoginFormLabels,
  PasswordChangeFormTemplateProps,
  PasswordChangeFormLabels,
} from './types';
