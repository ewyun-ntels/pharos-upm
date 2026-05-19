import type { ReactNode } from 'react';
import type { PasswordPolicy } from '../../../types/auth';

/**
 * 로그인 템플릿 Props
 *
 * Slot 기반으로 extension에서 모든 요소를 커스터마이징 가능
 */
export interface LoginTemplateProps {
  /**
   * 좌측 브랜딩 영역 전체를 커스텀 컴포넌트로 교체
   * 설정 시 branding prop은 무시됨
   */
  brandingSlot?: ReactNode;

  /**
   * 푸터 영역 (저작권 등)을 커스텀 컴포넌트로 교체
   */
  footerSlot?: ReactNode;

  /**
   * 기본 브랜딩 설정 (brandingSlot 없을 때 사용)
   */
  branding?: {
    /** 로고 (ReactNode - 이미지, SVG, 커스텀 컴포넌트 모두 가능) */
    logo?: ReactNode;
    /** 배경 CSS 클래스 */
    backgroundClassName?: string;
    /** 저작권 문구 또는 컴포넌트 */
    copyright?: ReactNode;
  };

  /** 우측 액션 영역에 렌더링할 컨텐츠 */
  children: ReactNode;
}

/**
 * 로그인 폼 레이블
 */
export interface LoginFormLabels {
  title?: string;
  username?: string;
  password?: string;
  signIn?: string;
  signingIn?: string;
}

/**
 * 로그인 폼 템플릿 Props
 */
export interface LoginFormTemplateProps {
  /** 폼 제출 핸들러 */
  onSubmit: (username: string, password: string) => Promise<void>;
  /** 에러 메시지 */
  errorMessage?: string;
  /** 로딩 상태 */
  isLoading?: boolean;
  /** 레이블 (직접 지정) */
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
  changePassword?: string;
  changing?: string;
  backToLogin?: string;
}

/**
 * 비밀번호 변경 폼 템플릿 Props
 */
export interface PasswordChangeFormTemplateProps {
  /** 폼 제출 핸들러 */
  onSubmit: (newPassword: string, currentPassword?: string) => Promise<void>;
  /** 뒤로가기 핸들러 */
  onBack: () => void;
  /** 에러 메시지 */
  errorMessage?: string;
  /** 안내 메시지 (만료/최초 로그인) */
  guideMessage?: string;
  /** 비밀번호 정책 설명 */
  policyMessage?: string;
  /** 비밀번호 정책 (클라이언트 사전 검증용) */
  passwordPolicy?: PasswordPolicy;
  /** 로딩 상태 */
  isLoading?: boolean;
  /** 만료 여부 (true면 현재 비밀번호 입력 숨김) */
  isExpired?: boolean;
  /** 레이블 (직접 지정) */
  labels?: PasswordChangeFormLabels;
}
