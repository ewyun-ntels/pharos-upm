
import React from 'react';
import { NotificationTableTabs } from '@features/notification';

/**
 * Notification Management Page
 *
 * Notification Rules를 관리하는 페이지
 *
 * Features:
 * - Rules 탭: Notification Rule 목록 조회/관리 (SNMP, Slack, etc.)
 */
export default function NotificationPage() {
  return (
    <main className="flex flex-col w-full h-full">
      <NotificationTableTabs />
    </main>
  );
}
