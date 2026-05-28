import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@lib/axios';

export interface HomeV2Config {
  datasource: string;
}

export function useHomeV2Config() {
  return useQuery({
    queryKey: ['home-v2-config'],
    queryFn: async () => {
      const response = await axiosInstance.get<HomeV2Config>('/home-v2');
      return response.data;
    },
    staleTime: Infinity,
  });
}
