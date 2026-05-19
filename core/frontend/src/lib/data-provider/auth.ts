import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Navigate } from 'react-router-dom';
import React from 'react';
import { getAuthProvider } from './registry';
import { useAuthStore } from '@pharos/shared/features/auth';

// ============================================================
// useGetIdentity
// ============================================================

export function useGetIdentity<T = any>() {
  const query = useQuery<T>({
    queryKey: ['identity'],
    queryFn: () => getAuthProvider().getIdentity() as Promise<T>,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
  });

  // Match refine's return shape: { data }
  return { data: query.data, isLoading: query.isLoading, query };
}

// ============================================================
// useLogout
// ============================================================

export function useLogout() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => getAuthProvider().logout(),
    onSuccess: () => {
      queryClient.clear();
    },
  });

  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    mutation,
    isPending: mutation.isPending,
  };
}

// ============================================================
// Authenticated component
// ============================================================

interface AuthenticatedProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  key?: string;
  loading?: React.ReactNode;
}

export const Authenticated: React.FC<AuthenticatedProps> = ({ children, fallback }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    return React.createElement(React.Fragment, null, fallback ?? React.createElement(Navigate, { to: '/login', replace: true }));
  }

  return React.createElement(React.Fragment, null, children);
};
