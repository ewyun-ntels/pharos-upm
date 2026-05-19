import type { ReactNode } from 'react';
import type { PasswordPolicy } from '../../../types/auth';

/**
 * 모던 카드 스타일 로그인 템플릿 Props
 */
export interface LoginTemplateProps {
  /**
  * 상단 브랜드 영역을 커스텀 컴포넌트로 교체
   */
  headerSlot?: ReactNode;

  /**
   * 푸터 영역 (약관, 저작권 등)을 커스텀 컴포넌트로 교체
   */
  footerSlot?: ReactNode;

  /**
   * 기본 브랜드 설정 (headerSlot 없을 때 사용)
   */
  header?: {
    /** 로고 (ReactNode - 이미지, SVG, 커스텀 컴포넌트 모두 가능) */
    logo?: ReactNode;
    /** 브랜드명 */
    brandName?: string;
    /** 브랜드 링크 */
    linkHref?: string;
  };

  /** 템플릿 메인 컨텐츠 */
  children: ReactNode;

  /** 배경 CSS 클래스 (기본: 그라데이션) */
  backgroundColor?: string;
}

/**
 * 로그인 폼 레이블
 */
export interface LoginFormLabels {
  title?: string;
  subtitle?: string;
  username?: string;
  password?: string;
  signIn?: string;
  signingIn?: string;
  forgotPassword?: string;
  noAccount?: string;
  signUp?: string;
}

/**
 * 로그인 폼 템플릿 Props
 */
export interface LoginFormTemplateProps {
  logo?: ReactNode;
  onSubmit: (username: string, password: string) => Promise<void>;
  errorMessage?: string;
  isLoading?: boolean;
  labels?: LoginFormLabels;
}

/**
 * 비밀번호 변경 폼 레이블
 */
export interface PasswordChangeFormLabels {
  title?: string;
  currentPassword?: string;
  newPassword?: string;
  confirmPassword?: string;
  submit?: string;
  submitting?: string;
  passwordMismatch?: string;
  backToLogin?: string;
}

/**
 * 비밀번호 변경 폼 템플릿 Props
 */
export interface PasswordChangeFormTemplateProps {
  onSubmit: (newPassword: string, currentPassword?: string) => Promise<void>;
  onBack?: () => void;
  errorMessage?: string;
  guideMessage?: string;
  policyMessage?: string;
  passwordPolicy?: PasswordPolicy;
  isLoading?: boolean;
  isExpired?: boolean;
  labels?: PasswordChangeFormLabels;
}
