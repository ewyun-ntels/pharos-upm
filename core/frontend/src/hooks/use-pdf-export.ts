import React, {useState, useRef, useCallback} from 'react';
import {pdf, DocumentProps} from '@react-pdf/renderer';
import {format} from '@pharos/shared/components';

type PdfComponentProps<T> = {
  data: T;
} & DocumentProps;

const useCreateChartRef = () => {
  const ref = useRef<HTMLDivElement>(null);
  return ref;
};

export function usePdfExport<T>(
  fileName: string,
  Component: React.ComponentType<PdfComponentProps<T>>,
) {
  const [isCollecting, setIsCollecting] = useState(false);
  const chartRef = useCreateChartRef();

  const handleDownload = useCallback(
    async (newData: T) => {
      setIsCollecting(true);
      try {
        const dateStr = format(new Date(), 'yyyyMMdd_HHmmss');
        const fullFileName = `${fileName}_${dateStr}`;

        const blob = await pdf(
          React.createElement(Component, {data: newData} as PdfComponentProps<T>),
        ).toBlob();
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${fullFileName}.pdf`;
        link.click();
        URL.revokeObjectURL(url);
      } finally {
        setIsCollecting(false);
      }
    },
    [Component, fileName],
  );

  return {
    chartRef,
    isCollecting,
    handleDownload,
  };
}
