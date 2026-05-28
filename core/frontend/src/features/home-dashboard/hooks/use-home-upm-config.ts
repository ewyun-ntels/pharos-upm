import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@lib/axios';

export interface HomeUPMConfig {
  datasource: string;
}

export function useHomeUPMConfig() {
  return useQuery({
    queryKey: ['home-upm-config'],
    queryFn: async () => {
      const response = await axiosInstance.get<HomeUPMConfig>('/home-upm');
      return response.data;
    },
    staleTime: Infinity,
  });
}
