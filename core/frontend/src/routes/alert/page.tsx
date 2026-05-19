
import React from 'react';
import { AlertTableTabs } from '@features/alert';

/**
 * Alert Management Page
 * 
 * Alert Rules와 Alert History를 탭으로 구분하여 관리하는 페이지
 * 
 * Features:
 * - Rules 탭: Alert Rule 목록 조회/관리
 * - History 탭: Alert History 조회 (백엔드 페이징 지원 대기 중)
 * 
 * @see {@link @features/alert/README.md} for more details
 */
export default function AlertPage() {
  return (
    <main className="flex flex-col w-full h-full">
      <AlertTableTabs />
    </main>
  );
}
