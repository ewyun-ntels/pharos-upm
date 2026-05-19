import React, {useState, useEffect, useRef} from 'react';
import {Input} from '@pharos/shared/components/ui';
import {Pencil} from 'lucide-react';

interface InlineEditProps {
  value: string;
  placeholder: string;
  onSave: (value: string) => void;
}

export const InlineEdit = ({value, placeholder, onSave}: InlineEditProps) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setEditValue(value);
  }, [value]);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleSave = () => {
    setIsEditing(false);
    if (editValue !== value) {
      onSave(editValue);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSave();
    } else if (e.key === 'Escape') {
      setEditValue(value);
      setIsEditing(false);
    }
  };

  if (isEditing) {
    return (
      <Input
        ref={inputRef}
        value={editValue}
        onChange={(e) => setEditValue(e.target.value)}
        onBlur={handleSave}
        onKeyDown={handleKeyDown}
        onClick={(e) => e.stopPropagation()}
        className="h-6 text-sm py-0 px-2 w-32"
        placeholder={placeholder}
      />
    );
  }

  return (
    <span
      className="text-sm font-medium truncate cursor-text hover:text-primary flex items-center gap-1 group/label"
      onClick={(e) => {
        e.stopPropagation();
        setIsEditing(true);
      }}
      title="Click to edit label"
    >
      {value || placeholder}
      <Pencil className="h-3 w-3 opacity-0 group-hover/label:opacity-50" />
    </span>
  );
};
