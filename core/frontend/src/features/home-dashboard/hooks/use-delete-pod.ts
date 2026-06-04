import {useMutation} from '@tanstack/react-query';
import {axiosInstance} from '@lib/axios';
import type {PodInfo} from '../types';

export interface DeletePodResponse {
  message: string;
  namespace: string;
  pod: string;
}

export function useDeletePod() {
  return useMutation({
    mutationFn: async (pod: Pick<PodInfo, 'namespace' | 'name'>) => {
      const namespace = encodeURIComponent(pod.namespace);
      const name = encodeURIComponent(pod.name);
      const response = await axiosInstance.delete<DeletePodResponse>(
        `/home-upm/pods/${namespace}/${name}`,
      );
      return response.data;
    },
  });
}
