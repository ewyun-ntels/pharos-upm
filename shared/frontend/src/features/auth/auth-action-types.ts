import React from 'react';
import type { User } from '../../types/user';
import type { PasswordPolicy } from '../../types/auth';

/**
 * Auth Action Plugin Info
 */
export interface AuthActionPluginInfo {
  /** 액션 ID - 백엔드 prepare 값과 매칭 */
  id: string;

  /** 표시 이름 */
  label: string;

  /** 아이콘 (선택) */
  icon?: React.ReactNode;

  /** 설명 (선택) */
  description?: string;
}

/**
 * Auth Action Context
 *
 * 외부 Extension에 전달되는 연동 인터페이스
 */

/**
 * 로그인 결과
 */
export interface LoginResult {
  success: boolean;
  error?: string;
}

/**
 * 비밀번호 변경 결과
 */
export interface ChangePasswordResult {
  success: boolean;
  error?: string;
}

export interface AuthActionContext {
  /**
   * 로그인 (Core에서 구현)
   * Extension은 이 메서드만 호출하면 됨
   */
  login: (username: string, password: string) => Promise<LoginResult>;

  /**
   * 비밀번호 변경 (Core에서 구현)
   * Extension은 이 메서드만 호출하면 됨
   */
  changePassword: (newPassword: string, currentPassword?: string) => Promise<ChangePasswordResult>;

  /** DataProvider (Refine 호환) */
  dataProvider: {
    getOne: <T = unknown>(params: { resource: string; id?: string }) => Promise<T>;
    create: <T = unknown>(params: { resource: string; values: unknown }) => Promise<T>;
    update: <T = unknown>(params: {
      resource: string;
      id?: string;
      values: unknown;
    }) => Promise<T>;
    /** 비표준 API 호출용 */
    custom: <T = unknown>(params: {
      url: string;
      method: string;
      payload?: unknown;
    }) => Promise<T>;
  };

  /** 유틸리티 */
  utils: {
    /** i18n 번역 */
    t: (key: string, options?: Record<string, unknown>) => string;
    /** 토스트 알림 */
    toast: (message: string, type?: 'success' | 'error' | 'info') => void;
    /** 라우팅 */
    navigate: (path: string) => void;
  };

  /** 설정 정보 (백엔드에서 조회) */
  config?: {
    passwordPolicy?: PasswordPolicy;
    [key: string]: unknown;
  };
}

/**
 * Auth Action Base Props
 *
 * 모든 AuthAction 컴포넌트가 받는 기본 Props
 */
export interface AuthActionBaseProps {
  /** 현재 사용자 정보 (로그인 후) */
  user: User | null;

  /** 액션 완료 시 호출 */
  onComplete: () => void;

  /** 액션 취소 시 호출 (선택) */
  onCancel?: () => void;

  /** 외부 Extension용 연동 인터페이스 */
  context: AuthActionContext;
}

/**
 * Login Action Props
 */
export interface LoginActionProps extends AuthActionBaseProps {
  /** 이전 로그인 실패 메시지 */
  errorMessage?: string;
}

/**
 * Password Change Action Props
 */
export interface PasswordChangeActionProps extends AuthActionBaseProps {
  /** 비밀번호 정책 */
  passwordPolicy: PasswordPolicy;
  /** 만료로 인한 변경 여부 */
  isExpired: boolean;
  /** 만료일 (User 스키마와 일치: Date | null | undefined) */
  expiresAt?: Date | null;
}

/**
 * Setup Profile Action Props
 */
export interface SetupProfileActionProps extends AuthActionBaseProps {
  /** 필수 입력 필드 목록 */
  requiredFields: string[];
}

/**
 * Setup 2FA Action Props
 */
export interface Setup2FAActionProps extends AuthActionBaseProps {
  /** QR 코드 URL */
  qrCodeUrl: string;
  /** 시크릿 키 */
  secret: string;
}

/**
 * Accept Terms Action Props
 */
export interface AcceptTermsActionProps extends AuthActionBaseProps {
  /** 약관 내용 */
  termsContent: string;
  /** 약관 버전 */
  termsVersion: string;
}

/**
 * Auth Action Plugin
 */
export interface AuthActionPlugin<T extends AuthActionBaseProps = AuthActionBaseProps> {
  /** 플러그인 메타데이터 */
  info: AuthActionPluginInfo;

  /** 컴포넌트 (직접 또는 lazy loading) */
  component:
    | React.ComponentType<T>
    | (() => Promise<{ default: React.ComponentType<T> }>);

  /**
   * 액션 실행 전 추가 Props 조회
   * 백엔드에서 필요한 데이터 fetch
   */
  getProps?: (user: User | null, context?: AuthActionContext) => Promise<Partial<T>>;
}
