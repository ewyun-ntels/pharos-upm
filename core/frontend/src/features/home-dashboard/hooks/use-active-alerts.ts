import { useCallback, useEffect, useRef, useState } from 'react';
import { axiosInstance } from '@lib/axios';

export interface AlarmItem {
  id: string;
  resource: string;
  event: string;
  environment: string;
  severity: string;
  status: string;
  service: string[];
  text: string;
  lastReceiveTime: string;
}

interface AlertaResponse {
  alerts: AlarmItem[];
}

export interface ActiveAlertsResult {
  alerts: AlarmItem[];
  criticalCount: number;
  majorCount: number;
  minorCount: number;
  isLoading: boolean;
  isError: boolean;
  refetch: () => Promise<void>;
}

export function useActiveAlerts(refetchInterval = 30_000): ActiveAlertsResult {
  const [alerts, setAlerts] = useState<AlarmItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isError, setIsError] = useState(false);
  const isMounted = useRef(true);

  useEffect(() => {
    isMounted.current = true;
    return () => { isMounted.current = false; };
  }, []);

  const fetchAlerts = useCallback(async () => {
    setIsLoading(true);
    setIsError(false);
    try {
      const resp = await axiosInstance.get<AlertaResponse>('/alarm/alerts');
      const all = resp.data.alerts ?? [];
      const open = all.filter(a => a.status === 'open');
      if (isMounted.current) setAlerts(open);
    } catch {
      if (isMounted.current) setIsError(true);
    } finally {
      if (isMounted.current) setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAlerts();
  }, [fetchAlerts]);

  useEffect(() => {
    if (refetchInterval === 0) return;
    const timer = setInterval(fetchAlerts, refetchInterval);
    return () => clearInterval(timer);
  }, [fetchAlerts, refetchInterval]);

  return {
    alerts,
    criticalCount: alerts.filter(a => a.severity.toLowerCase() === 'critical').length,
    majorCount: alerts.filter(a => a.severity.toLowerCase() === 'major').length,
    minorCount: alerts.filter(a => a.severity.toLowerCase() === 'minor').length,
    isLoading,
    isError,
    refetch: fetchAlerts,
  };
}
