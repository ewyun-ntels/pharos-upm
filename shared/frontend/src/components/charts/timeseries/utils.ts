import {x86} from 'murmurhash3js';

const CHART_COLOR_PALETTE: {[key: string]: string} = {
  chart_1:  'hsl(258, 90%, 66%, 1.0)', // #8B5CF6
  chart_2:  'hsl(216, 94%, 55%, 1.0)', // #2079F9
  chart_3:  'hsl(145, 60%, 47%, 1.0)', // #31BE6C
  chart_4:  'hsl(28, 80%, 52%, 1.0)',  // #E67E22
  chart_5:  'hsl(238, 77%, 65%, 1.0)', // #5D65EE
  chart_6:  'hsl(65, 100%, 42%, 1.0)', // #C9D700
  chart_7:  'hsl(44, 84%, 49%, 1.0)',  // #EBB00E
  chart_8:  'hsl(6, 78%, 57%, 1.0)',   // #E74C3C
  chart_9:  'hsl(167, 97%, 45%, 1.0)', // #03DCAD
  chart_10: 'hsl(87, 67%, 43%, 1.0)',  // #78B424
  chart_11: 'hsl(0, 54%, 41%, 1.0)',   // #A03030
  chart_12: 'hsl(204, 70%, 53%, 1.0)', // #3498DB
  chart_13: 'hsl(194, 77%, 60%, 1.0)', // #49C3EA
  chart_14: 'hsl(277, 44%, 41%, 1.0)', // #7D3C98
  chart_15: 'hsl(293, 64%, 54%, 1.0)', // #C743D7
  chart_16: 'hsl(146, 63%, 38%, 1.0)', // #239C56
  chart_17: 'hsl(229, 56%, 48%, 1.0)', // #3858C0
};

export function getColorByKey(key: string): string {
  const hash1 = x86.hash32(key, 0);
  const hash2 = x86.hash32(key, 123456);
  const combinedHash = Math.abs(hash1 ^ hash2);
  const colorIndex = combinedHash % 17;
  return CHART_COLOR_PALETTE['chart_' + (colorIndex + 1)].replace('fillOpacity', '1');
}

export function getSortValue(key: string, metric: any[], chartOptions: any): number {
  const values = metric.map(d => Number(d[key] ?? 0));

  if (chartOptions.showMax) return Math.max(...values);
  if (chartOptions.showMin) return Math.min(...values);
  if (chartOptions.showAverage) {
    const validValues = values.filter(v => !isNaN(v));
    return validValues.length > 0 ? validValues.reduce((a, b) => a + b) / validValues.length : 0;
  }
  if (chartOptions.showTotal) return values.reduce((a, b) => a + b, 0);

  // Default: showLast 또는 아무것도 없을 때 최신 값
  return values[values.length - 1] || 0;
}

export function createUniqueColorManager() {
  const colorMap = new Map<string, string>();
  const usedColors = new Set<string>();

  return (legendLabel: string): string => {
    if (colorMap.has(legendLabel)) return colorMap.get(legendLabel)!;

    let color = getColorByKey(legendLabel);
    let attempts = 0;

    while (usedColors.has(color) && attempts < 17) {
      const fallbackIndex = Math.abs(x86.hash32(legendLabel, ++attempts)) % 17;
      color = CHART_COLOR_PALETTE['chart_' + (fallbackIndex + 1)].replace('fillOpacity', '1');
    }

    colorMap.set(legendLabel, color);
    usedColors.add(color);
    return color;
  };
}
