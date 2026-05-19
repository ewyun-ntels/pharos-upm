import React, {useState} from 'react';
import {Undo2} from 'lucide-react';
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from '@pharos/shared/components/ui';
import {IconButton} from '@pharos/shared/components/ui-extension';

type SchemaEditorActionsProps = {
  isDirty: boolean;
  isSaving: boolean;
  onSave: (description?: string) => void;
  onReset: () => void;
};

export function SchemaEditorActions({isDirty, isSaving, onSave, onReset}: SchemaEditorActionsProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [description, setDescription] = useState('');

  const handleSaveClick = () => {
    setDescription('');
    setDialogOpen(true);
  };

  const handleConfirm = () => {
    setDialogOpen(false);
    onSave(description.trim() || undefined);
  };

  return (
    <>
      <IconButton
        variant="outline"
        icon={<Undo2 />}
        onClick={onReset}
        disabled={!isDirty || isSaving}
      >
        Previous
      </IconButton>
      <Button
        size="default"
        className="cursor-pointer"
        onClick={handleSaveClick}
        disabled={!isDirty || isSaving}
      >
        {isSaving ? 'Saving...' : 'Save'}
      </Button>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Save Configuration</DialogTitle>
            <DialogDescription>
              Optionally add a description for this change (shown in history).
            </DialogDescription>
          </DialogHeader>
          <div className="py-2">
            <Label htmlFor="save-description" className="text-sm mb-1 block">
              Description (optional)
            </Label>
            <Input
              id="save-description"
              placeholder="e.g. Added department field"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleConfirm()}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleConfirm}>Save</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
