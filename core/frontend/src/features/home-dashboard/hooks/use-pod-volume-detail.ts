import { useQuery } from '@tanstack/react-query';
import type { PodInfo } from '../types';
import { runBatchQuery, escapePromLabelValue } from './use-kubernetes-data';

export interface PVCVolumeDetail {
  pvcName: string;
  usedBytes: number;
  freeBytes: number;
  usedPercent: number;
}

function getVal(row: Record<string, unknown>): number {
  const raw = row['Value'] ?? row['value'];
  const n = parseFloat(String(raw ?? 0));
  return isFinite(n) ? n : 0;
}

export function usePodVolumeDetail(pod: PodInfo | null, datasourceName: string) {
  return useQuery({
    queryKey: ['pod-volume-detail', pod?.namespace, pod?.name, datasourceName],
    enabled: !!pod,
    queryFn: async () => {
      const ns = escapePromLabelValue(pod!.namespace);
      const podName = escapePromLabelValue(pod!.name);

      // kubelet_volume_stats_* 에는 pod 레이블이 없으므로
      // kube_pod_spec_volumes_persistentvolumeclaims_info 와 join 하여 pod 기준 필터링.
      // 양쪽 모두 max by (namespace, persistentvolumeclaim) 로 집계하여
      // uid, volume 등 extra 레이블로 인한 다중 series 문제를 방지하고 1:1 join 보장.
      const kubeletSel = `job="kubelet", metrics_path="/metrics", namespace="${ns}"`;
      const podPvcRight =
        `max by (namespace, persistentvolumeclaim) (` +
        `kube_pod_spec_volumes_persistentvolumeclaims_info{namespace="${ns}", pod="${podName}"})`;
      const join = `* on(namespace, persistentvolumeclaim) group_left() ${podPvcRight}`;

      const queries = {
        capacityBytes:
          `max by (namespace, persistentvolumeclaim) ` +
          `(kubelet_volume_stats_capacity_bytes{${kubeletSel}}) ${join}`,
        availableBytes:
          `max by (namespace, persistentvolumeclaim) ` +
          `(kubelet_volume_stats_available_bytes{${kubeletSel}}) ${join}`,
      };

      const data = await runBatchQuery(datasourceName, queries);

      const capacityMap = new Map<string, number>();
      const availableMap = new Map<string, number>();
      const getPvc = (row: Record<string, unknown>) =>
        String(row['persistentvolumeclaim'] ?? '');

      for (const row of data.capacityBytes ?? []) {
        const pvc = getPvc(row);
        if (!pvc) continue;
        capacityMap.set(pvc, getVal(row));
      }
      for (const row of data.availableBytes ?? []) {
        const pvc = getPvc(row);
        if (!pvc) continue;
        availableMap.set(pvc, getVal(row));
      }

      const results: PVCVolumeDetail[] = [];
      for (const [pvc, capacity] of capacityMap.entries()) {
        const available = availableMap.get(pvc) ?? 0;
        const used = Math.max(0, capacity - available);
        results.push({
          pvcName: pvc,
          usedBytes: used,
          freeBytes: available,
          usedPercent: capacity > 0 ? (used / capacity) * 100 : 0,
        });
      }
      return results;
    },
    staleTime: 30_000,
  });
}
