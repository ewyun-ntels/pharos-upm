import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import iconUrl from './icon.svg';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { TableCardChart } from '@features/dashboard/panels/plugins/table/TableCard';
import type { TablePanelOptions } from './types';

export const tablePlugin: PanelPlugin<TablePanelOptions> = {
  info: {
    id: 'table',
    label: 'Table',
    icon: React.createElement('img', { src: iconUrl, alt: 'Table' }),
    description: 'Data table display',
    category: 'table',
  },
  component: TableCardChart,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
  actionHandlers: {
    // 콜럼 리사이즈 완료: columnConfigs[i].properties.size 업데이트
    // columnConfigs에 없는 컬럼은 새 entry 추가
    'resize-column': (payload, api) => {
      const sizing = payload as Record<string, number>;
      if (Object.keys(sizing).length === 0) return;

      const opts = api.getOptions() ?? {};
      const existingConfigs = opts.columnConfigs ?? [];

      // 기존 configs 업데이트 (key 매핑)
      const updated = existingConfigs.map((col) => {
        const newSize = sizing[String(col.key)];
        if (newSize === undefined) return col;
        return {
          ...col,
          properties: { ...(col.properties ?? {}), size: newSize },
        };
      });

      // 기존 configs에 없는 컬럼은 sizing에 있으면 새로 추가
      const existingKeys = new Set(existingConfigs.map((c) => String(c.key)));
      for (const [colKey, colSize] of Object.entries(sizing)) {
        if (!existingKeys.has(colKey)) {
          updated.push({ key: colKey, properties: { size: colSize } });
        }
      }

      api.updateOptions({ columnConfigs: updated });
    },
  },
};

// 🔥 직접 등록
panelPluginRegistry.register(tablePlugin);
