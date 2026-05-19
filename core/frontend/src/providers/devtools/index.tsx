/// <reference types="vite/client" />
'use client';

import React from 'react';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

export const DevtoolsProvider = (props: React.PropsWithChildren) => {
  return (
    <>
      {props.children}
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </>
  );
};
