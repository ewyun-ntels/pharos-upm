export interface OptionData {
  label: string;
  content: React.ReactNode;
}

export interface EditorSelectContextType {
  selectedValue: string;
  setSelectedValue: (value: string) => void;
  optionsData: Record<string, OptionData>;
  setOptionData: (value: string, label: string, data: React.ReactNode) => void;
  disabledOptions: Set<string>;
  setOptionDisabled: (value: string, disabled: boolean) => void;
}
