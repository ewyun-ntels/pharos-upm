import type { PasswordPolicy } from '../types/auth';

/**
 * 패스워드 정책 검증 시 출력할 에러 메시지 커스터마이징
 * 각 항목은 제약 값(n)을 받아 문자열을 반환하는 함수
 *
 * @example 한국어 메시지
 * {
 *   minLength: (n) => `비밀번호는 최소 ${n}자 이상이어야 합니다`,
 *   minUppercase: (n) => `대문자를 ${n}자 이상 포함해야 합니다`,
 * }
 */
export interface PolicyValidationMessages {
  minLength?: (n: number) => string;
  minUppercase?: (n: number) => string;
  minLowercase?: (n: number) => string;
  minDigits?: (n: number) => string;
  minSpecial?: (n: number) => string;
}

const defaultMessages: Required<PolicyValidationMessages> = {
  minLength: (n) => `Password must be at least ${n} characters`,
  minUppercase: (n) => `Password must contain at least ${n} uppercase letter(s)`,
  minLowercase: (n) => `Password must contain at least ${n} lowercase letter(s)`,
  minDigits: (n) => `Password must contain at least ${n} digit(s)`,
  minSpecial: (n) => `Password must contain at least ${n} special character(s)`,
};

/**
 * 패스워드 정책 검증 유틸리티
 * 모든 로그인 템플릿에서 공통으로 사용
 *
 * @param policy
 * @param password
 * @param messages - 에러 메시지 커스터마이징 (미전달 시 영어 기본값 사용)
 * @returns 에러 메시지 문자열, 정책을 만족하면 null
 */
export function validatePasswordPolicy(
  policy: PasswordPolicy | undefined,
  password: string,
  messages?: PolicyValidationMessages
): string | null {
  if (!policy) return null;

  const msg = { ...defaultMessages, ...messages };

  const minLen = policy.min_length ?? 0;
  if (minLen > 0 && password.length < minLen) {
    return msg.minLength(minLen);
  }

  const minUpper = policy.min_uppercase ?? 0;
  if (minUpper > 0 && (password.match(/[A-Z]/g) || []).length < minUpper) {
    return msg.minUppercase(minUpper);
  }

  const minLower = policy.min_lowercase ?? 0;
  if (minLower > 0 && (password.match(/[a-z]/g) || []).length < minLower) {
    return msg.minLowercase(minLower);
  }

  const minDigits = policy.min_digits ?? 0;
  if (minDigits > 0 && (password.match(/[0-9]/g) || []).length < minDigits) {
    return msg.minDigits(minDigits);
  }

  const minSpecial = policy.min_special ?? 0;
  if (
    minSpecial > 0 &&
    (password.match(/[!@#~$%^&*()_+\-=[\]{};':"\\|,.<>/?]/g) || []).length < minSpecial
  ) {
    return msg.minSpecial(minSpecial);
  }

  return null;
}
