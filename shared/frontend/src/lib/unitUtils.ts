import convert, { BestConversion, BestKind, Unit } from 'convert';
import { fromUnixTime } from '@pharos/shared/components';

type UnitDateRate = 'bytes/sec(IEC)' | 'bits/sec(IEC)' | 'bytes/sec(SI)' | 'bits/sec(SI)';
type UnitData = 'bytes(IEC)' | 'bytes(SI)' | 'bits(IEC)' | 'bits(SI)';
type UnitThroughput = 'counts/sec(cps)' | 'requests/sec(rps)' | 'queries/sec(qps)' | 'ops/sec(ops)';
type UnitSignal = 'dBm';
export type UnitType =
  | 'milliseconds'
  | 'seconds'
  | 'local_time'
  | 'short'
  | 'numeric'
  | 'Base64_decoding'
  | '%(1-100)'
  | UnitData
  | UnitDateRate
  | UnitThroughput
  | UnitSignal
  | undefined;

export const tableUnitTypeList = [
  'milliseconds',
  'seconds',
  'local_time',
  'short',
  'numeric',
  'Base64_decoding',
  '%(1-100)',
  'bytes(IEC)',
  'bytes(SI)',
  'bits(IEC)',
  'bits(SI)',
  'bytes/sec(IEC)',
  'bits/sec(IEC)',
  'bytes/sec(SI)',
  'bits/sec(SI)',
  'counts/sec(cps)',
  'requests/sec(rps)',
  'queries/sec(qps)',
  'ops/sec(ops)',
  'dBm',
] as const;

export const unitTypeList = tableUnitTypeList.filter(
  (item) => !['local_time', 'Base64_decoding', 'seconds', 'number'].includes(item),
);

// Date formatting functions moved to shared
export function formatLocalTime(date: Date): string {
  const dateString = date.toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });

  const timeString = date.toLocaleTimeString('ko-KR', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });

  return `${dateString} ${timeString}`; // 2024.07.14. 오전 12:00
}

export function formatLocalTime24(date: Date): string {
  const dateString = date.toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });

  const timeString = date.toLocaleTimeString('ko-KR', {
    hour: 'numeric',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });

  return `${dateString} ${timeString}`;
}

export const convertToLocalTime = (
  value: string | Date,
): string => {
  let utcDate: Date;

  if (value instanceof Date) {
    utcDate = value;
  } else if (value.includes('Z') || /[+-]\d{2}:\d{2}$/.test(value)) {
    utcDate = new Date(value);
  } else {
    // 시간대 정보가 없으면 UTC로 가정
    utcDate = new Date(value + 'Z');
  }

  const unixTimestamp = utcDate.getTime() / 1000;

  return formatLocalTime24(fromUnixTime(unixTimestamp));
};

function convertUnit(
  value: number | string,
  unit: UnitType,
  decimalPlaces: number = 2,
): BestConversion<any, any> {
  const convertAndFormat = (from: Unit, system?: BestKind, suffix = '') => {
    const result = convert(value as number, from).to('best', system);
    return {
      quantity: result.quantity,
      unit: result.unit,
      toString: () => `${result.quantity.toFixed(decimalPlaces)} ${result.unit}${suffix}`,
    };
  };

  const convertShortNumber = (value: number, suffix: string) => {
    return {
      quantity: value,
      unit: '',
      toString: () => `${formatNumberWithUnit(value)} ${suffix}`,
    };
  };

  switch (unit) {
    case 'bytes/sec(IEC)':
      return convertAndFormat('bytes', 'imperial', '/s');
    case 'bytes/sec(SI)':
      return convertAndFormat('bytes', 'metric', '/s');
    case 'bits/sec(IEC)':
      return convertAndFormat('bits', 'imperial', '/s');
    case 'bits/sec(SI)':
      return convertAndFormat('bits', 'metric', '/s');
    case 'bytes(IEC)':
      return convertAndFormat('bytes', 'imperial');
    case 'bytes(SI)':
      return convertAndFormat('bytes', 'metric');
    case 'bits(IEC)':
      return convertAndFormat('bits', 'imperial');
    case 'bits(SI)':
      return convertAndFormat('bits', 'metric');
    case 'milliseconds': {
      const result = convert(value as number, 'ms').to('best');
      return {
        quantity: result.quantity,
        unit: result.unit,
        toString: () => `${result.quantity.toFixed(decimalPlaces)} ${result.unit}`,
      };
    }
    case 'seconds': {
      const numValue = value as number;
      if (numValue === 0) {
        return {
          quantity: 0,
          unit: 's',
          toString: () => '0 s',
        };
      }
      const result = convert(numValue, 's').to('best');
      return {
        quantity: result.quantity,
        unit: result.unit,
        toString: () =>
          `${result.quantity % 1 === 0 ? result.quantity.toFixed(0) : result.quantity.toFixed(decimalPlaces)} ${result.unit}`,
      };
    }
    case 'counts/sec(cps)':
      return convertShortNumber(value as number, 'c/s');
    case 'requests/sec(rps)':
      return convertShortNumber(value as number, 'req/s');
    case 'queries/sec(qps)':
      return convertShortNumber(value as number, 'qps');
    case 'ops/sec(ops)':
      return convertShortNumber(value as number, 'ops/s');
    case 'short': {
      return {
        quantity: value,
        unit: '',
        toString: () => `${formatNumberWithUnit(value as number, decimalPlaces)}`,
      };
    }
    case '%(1-100)':
      return {
        quantity: value,
        unit: '%',
        toString: () => `${(value as number).toFixed(decimalPlaces)}%`,
      };
    case 'local_time':
      const localTimeStr = convertToLocalTime(String(value));
      return {
        quantity: value,
        unit: '',
        toString: () => localTimeStr,
      };
    case 'numeric':
      const formattedNumber = formatNumberWithCommas(Number(value), decimalPlaces);
      return {
        quantity: value,
        unit: '',
        toString: () => formattedNumber.toString(),
      };
    case 'dBm':
      return {
        quantity: value,
        unit: 'dBm',
        toString: () => `${(value as number).toFixed(decimalPlaces)} dBm`,
      };
    case 'Base64_decoding':
      return {
        quantity: value,
        unit: '',
        toString: () => {
          const stringValue = String(value);
          if (!/^[A-Za-z0-9+/]*={0,2}$/.test(stringValue.trim())) {
            return stringValue;
          }

          try {
            const decoded = JSON.parse(atob(stringValue.trim()));
            return JSON.stringify(decoded);
          } catch {
            return stringValue;
          }
        },
      };

    default:
      return {
        quantity: value || '-',
        unit: '',
        toString: () => value?.toString() || '-',
      };
  }
}

const formatNumber = (num: number, d: number = 2) => (num % 1 === 0 ? num.toFixed(0) : num.toFixed(d));

export function formatNumberWithCommas(num: number, decimalPlaces: number = 2): string {
  const multiplier = Math.pow(10, decimalPlaces);
  const roundedNum = Math.round(num * multiplier) / multiplier;

  const [integerPart, decimalPart] = roundedNum.toString().split('.');

  const formattedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, ',');

  if (decimalPart && parseFloat(decimalPart) > 0) {
    const formattedDecimal = `.${decimalPart.padEnd(decimalPlaces, '0').slice(0, decimalPlaces)}`;
    return `${formattedInteger}${formattedDecimal}`;
  }

  return formattedInteger;
}

export function formatNumberWithUnit(num: number, decimalPlaces: number = 2): string {
  if (num >= 1_000_000_000) {
    return formatNumber(num / 1_000_000_000, decimalPlaces) + 'G';
  } else if (num >= 1_000_000) {
    return formatNumber(num / 1_000_000, decimalPlaces) + 'M';
  } else if (num >= 1_000) {
    return formatNumber(num / 1_000, decimalPlaces) + 'K';
  } else if (num == 0) {
    return '0';
  }

  return num.toFixed(decimalPlaces);
}

export { convertUnit };
