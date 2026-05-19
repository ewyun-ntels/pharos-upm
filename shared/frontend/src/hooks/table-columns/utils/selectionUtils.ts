export function getSelectionKey(item: any, keyName?: string, fallbackKey?: string) {
  if (typeof item === 'string' || typeof item === 'number') {
    return item;
  }
  return item?.[keyName || fallbackKey || 'id'];
}

export function handleSelectAll({
  checked,
  data,
  keyName,
  fallbackKey,
  setSelected,
}: {
  checked: boolean;
  data: any[];
  keyName?: string;
  fallbackKey?: string;
  setSelected: (keys: string[]) => void;
}) {
  const newSelected = checked
    ? (data || []).map((item) => String(getSelectionKey(item, keyName, fallbackKey)))
    : [];
  setSelected(newSelected);
}

export function handleSelectOne({
  item,
  checked,
  selectedItems,
  keyName,
  fallbackKey,
  setSelected,
}: {
  item: any;
  checked: boolean;
  selectedItems: string[];
  keyName?: string;
  fallbackKey?: string;
  setSelected: (keys: string[]) => void;
}) {
  const key = String(getSelectionKey(item, keyName, fallbackKey));
  let newSelected = [...selectedItems];
  if (checked) {
    if (!newSelected.includes(key)) newSelected.push(key);
  } else {
    newSelected = newSelected.filter((k) => k !== key);
  }
  setSelected(newSelected);
}
