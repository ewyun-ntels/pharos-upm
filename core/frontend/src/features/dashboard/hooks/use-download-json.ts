import {useCallback} from 'react';

function formatTimestamp(): string {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}_${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
}

export const useDownloadJSON = () => {
  return useCallback((data: any, filename = 'data.json') => {
    try {
      const baseName = filename.endsWith('.json') ? filename.slice(0, -5) : filename;
      const finalFilename = `${baseName}_${formatTimestamp()}.json`;

      const jsonString = JSON.stringify(data, null, 2);
      const blob = new Blob([jsonString], {type: 'application/json'});
      const url = URL.createObjectURL(blob);

      const link = document.createElement('a');
      link.href = url;
      link.download = finalFilename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('JSON 다운로드 중 오류:', error);
    }
  }, []);
};
