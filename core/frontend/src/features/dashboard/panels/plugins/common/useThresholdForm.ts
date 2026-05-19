import {useState} from 'react';

export type Threshold = {
  value?: string;
  color?: string;
  type?: 'solid' | 'dash' | 'dot';
  lineWidth?: number;
};

interface WithThresholds {
  thresholds?: Threshold[];
}

export function useThresholdForm<T extends WithThresholds>(
  handleChange: (updater: (draft: T) => void) => void,
) {
  const [newThreshold, setNewThreshold] = useState<Threshold>({
    value: '',
    color: '#ff0000',
    type: 'solid',
    lineWidth: 2,
  });

  const handleAddThreshold = () => {
    if (!newThreshold.value) return;
    handleChange(draft => {
      if (!draft.thresholds) draft.thresholds = [];
      draft.thresholds.push({...newThreshold});
    });
    setNewThreshold({value: '', color: '#ff0000', type: 'solid', lineWidth: 2});
  };

  const handleRemoveThreshold = (index: number) => {
    handleChange(draft => {
      draft.thresholds = draft.thresholds?.filter((_, i) => i !== index);
    });
  };

  const handleUpdateThreshold = (index: number, updates: Partial<Threshold>) => {
    handleChange(draft => {
      if (draft.thresholds) {
        draft.thresholds[index] = { ...draft.thresholds[index], ...updates };
      }
    });
  };

  return {newThreshold, setNewThreshold, handleAddThreshold, handleRemoveThreshold, handleUpdateThreshold};
}
