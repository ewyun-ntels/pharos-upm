
import { Suspense } from 'react';
import { Navigate } from 'react-router-dom';
import { Authenticated } from '@/lib/data-provider';
import { HOME_PATH } from '@pharos/meta/site-config';

export default function IndexPage() {
  return (
    <Suspense>
      <Authenticated key="home-page">
        <Navigate to={HOME_PATH} replace />
      </Authenticated>
    </Suspense>
  );
}
