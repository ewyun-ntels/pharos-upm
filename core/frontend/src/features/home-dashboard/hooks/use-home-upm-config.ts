import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@lib/axios';

export interface HomeUPMConfig {
  datasource: string;
  restartWindow: string;
  warningRestarts: number;
  errorRestarts: number;
}

export const DEFAULT_HOME_UPM_CONFIG: HomeUPMConfig = {
  datasource: 'prometheus-metric',
  restartWindow: '1h',
  warningRestarts: 1,
  errorRestarts: 3,
};

const PROM_DURATION_PATTERN = /^[1-9][0-9]*(ms|s|m|h|d|w|y)$/;

function toPositiveNumber(value: unknown, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function normalizeHomeUPMConfig(config: Partial<HomeUPMConfig>): HomeUPMConfig {
  const warningRestarts = toPositiveNumber(config.warningRestarts, DEFAULT_HOME_UPM_CONFIG.warningRestarts);
  const errorRestarts = toPositiveNumber(config.errorRestarts, DEFAULT_HOME_UPM_CONFIG.errorRestarts);

  return {
    datasource: config.datasource || DEFAULT_HOME_UPM_CONFIG.datasource,
    restartWindow: config.restartWindow && PROM_DURATION_PATTERN.test(config.restartWindow)
      ? config.restartWindow
      : DEFAULT_HOME_UPM_CONFIG.restartWindow,
    warningRestarts: Math.min(warningRestarts, errorRestarts),
    errorRestarts,
  };
}

export function useHomeUPMConfig() {
  return useQuery({
    queryKey: ['home-upm-config'],
    queryFn: async () => {
      const response = await axiosInstance.get<Partial<HomeUPMConfig>>('/home-upm');
      return normalizeHomeUPMConfig(response.data ?? {});
    },
    staleTime: Infinity,
  });
}
