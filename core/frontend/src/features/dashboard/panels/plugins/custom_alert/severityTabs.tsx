import {IconButton} from '@pharos/shared/components/ui-extension';
import {Tabs, TabsList, TabsTrigger} from '@pharos/shared/components/ui-extension';
import {X} from '@pharos/shared/components';
import React, {useMemo} from 'react';

interface SeverityTabsProps {
  data: any[];
  statusTab: string;
  setStatusTab: (tab: string) => void;
}

export default function SeverityTabs({data, statusTab, setStatusTab}: SeverityTabsProps) {
  const severityLevelsOrders = ['critical', 'major', 'minor', 'low'];

  const severityCounts = useMemo(() => {
    const counts: Record<string, number> = {};

    data.forEach((item: any) => {
      const statusArray = Array.isArray(item.status) ? item.status : [item];

      if (statusArray.length === 0) {
        counts['normal'] = (counts['normal'] || 0) + 1;
      } else {
        statusArray.forEach((statusItem: any) => {
          const sev = (statusItem.severity || '').toLowerCase();
          if (!sev) return;
          counts[sev] = (counts[sev] || 0) + 1;
        });
      }
    });
    return counts;
  }, [data]);

  const severityDataList = useMemo(() => {
    const extraList = Object.entries(severityCounts).map(([id, count]) => ({id, count}));
    return [...extraList];
  }, [severityCounts]);

  const sortedList = [...severityDataList].sort((a, b) => {
    const aIndex = severityLevelsOrders.indexOf(a.id.toLowerCase());
    const bIndex = severityLevelsOrders.indexOf(b.id.toLowerCase());
    const aRank = aIndex === -1 ? 999 : aIndex;
    const bRank = bIndex === -1 ? 999 : bIndex;
    return aRank - bRank;
  });

  return (
    <>
      {data.length > 0 && (
        <>
          <Tabs value={statusTab} onValueChange={setStatusTab}>
            <TabsList>
              {sortedList.map((item) => (
                <TabsTrigger key={item.id} value={item.id} badge={item.count}>
                  {item.id.charAt(0).toUpperCase() + item.id.slice(1)}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
          {statusTab && (
            <IconButton
              onClick={() => setStatusTab('')}
              variant="ghost"
              icon={<X />}
              className="w-5 h-5 p-0 rounded-full self-center"
            >
              Clear filter
            </IconButton>
          )}
        </>
      )}
    </>
  );
}
