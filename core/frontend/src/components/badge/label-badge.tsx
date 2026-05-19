import React, {useEffect, useState} from 'react';
import {x86} from 'murmurhash3js';
import {useToast} from '@hooks/use-toast';

const labelColorList = [
  {
    light: {keyColor: '#CDBFE5', valueColor: '#DFD6F0'},
    dark: {keyColor: '#3C324E', valueColor: '#4B3E62'},
  },
  {
    light: {keyColor: '#C5C3E6', valueColor: '#DEDCFA'},
    dark: {keyColor: '#31304D', valueColor: '#403F62'},
  },
  {
    light: {keyColor: '#B9C9E8', valueColor: '#D5DCEC'},
    dark: {keyColor: '#29335E', valueColor: '#313D71'},
  },
  {
    light: {keyColor: '#C8DDF6', valueColor: '#E0EBF8'},
    dark: {keyColor: '#002D4F', valueColor: '#003B67'},
  },
  {
    light: {keyColor: '#CCE4F0', valueColor: '#DBEFF9'},
    dark: {keyColor: '#1E394E', valueColor: '#234661'},
  },
  {
    light: {keyColor: '#BDDBD8', valueColor: '#CCE7E5'},
    dark: {keyColor: '#2C4F4F', valueColor: '#355F5F'},
  },
  {
    light: {keyColor: '#C7E0C5', valueColor: '#D5EAD4'},
    dark: {keyColor: '#3E4C44', valueColor: '#44574C'},
  },
  {
    light: {keyColor: '#D5CEC5', valueColor: '#EAE4DD'},
    dark: {keyColor: '#423D36', valueColor: '#524C43'},
  },
];

// 다크 모드 상태 감지 (실시간 반응 하기 위해)
function useDarkMode(): boolean {
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    const check = () => setIsDark(document.documentElement.classList.contains('dark'));
    check();

    const observer = new MutationObserver(() => check());
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    });

    return () => observer.disconnect();
  }, []);

  return isDark;
}

function getColorByKey(key: string, isDark: boolean) {
  const hash = x86.hash32(key, 0);
  const colorIndex = Math.abs(hash) % labelColorList.length;
  return isDark ? labelColorList[colorIndex].dark : labelColorList[colorIndex].light;
}

interface LabelBadgeProps {
  labels: Record<string, string | number | boolean> | null | undefined;
  enableCopy?: boolean;
}

export default function LabelBadge({labels, enableCopy = false}: LabelBadgeProps) {
  const isDark = useDarkMode();
  const {toast} = useToast();

  if (!labels || typeof labels !== 'object') return null;

  const handleCopyToClipboard = async (value: string | number | boolean) => {
    const textToCopy = String(value);

    try {
      await navigator.clipboard.writeText(textToCopy);
      toast({description: 'Copied successfully.'});
    } catch (err) {
      console.error('Copy failed:', err);
    }
  };

  return (
    <div className="flex flex-wrap items-center gap-1">
      {Object.entries(labels).map(([key, value]) => {
        const colors = getColorByKey(key, isDark);
        return (
          <div key={key} className="flex w-fit overflow-hidden rounded cursor-default">
            <span
              className="px-[6px] py-[2px] text-xs text-foreground break-word"
              style={{backgroundColor: colors.keyColor}}
            >
              {key}
            </span>
            <span
              className={`px-[6px] py-[2px] min-w-8 text-xs text-foreground truncate max-w-[150px] whitespace-nowrap ${
                enableCopy ? 'cursor-pointer hover:opacity-80' : ''
              }`}
              style={{backgroundColor: colors.valueColor}}
              title={String(value)}
              {...(enableCopy && {onClick: () => handleCopyToClipboard(value)})}
            >
              {String(value)}
            </span>
          </div>
        );
      })}
    </div>
  );
}
