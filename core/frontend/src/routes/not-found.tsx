import { Authenticated } from '@/lib/data-provider';
import { Suspense } from 'react';

import { Button } from '@pharos/shared/components/ui';
import { useNavigate } from 'react-router-dom';

export default function NotFound() {
  const navigate = useNavigate();
  return (
    <Suspense>
      <Authenticated key="not-found">
        <div className="flex flex-col items-center justify-center min-h-screen bg-background text-foreground">
          <h1 className="text-5xl font-bold mb-4">404</h1>
          <p className="text-xl mb-8">Sorry, the page you visited does not exist.</p>
          <Button
            onClick={() => navigate('/')}
            className="bg-primary hover:bg-primary-dark text-white"
          >
            Back Home
          </Button>
        </div>
      </Authenticated>
    </Suspense>
  );
}
