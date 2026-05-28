
import { Suspense } from 'react';
import { Navigate } from 'react-router-dom';
import { Authenticated } from '@/lib/data-provider';

export default function IndexPage() {
  return (
    <Suspense>
      <Authenticated key="home-page">
        <Navigate to="/home-upm" replace />
      </Authenticated>
    </Suspense>
  );
}
