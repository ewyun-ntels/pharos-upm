import React, {useState} from 'react';
import {KeyRound} from 'lucide-react';
import {
  Button,
  Input,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@pharos/shared/components/ui';
import {Title} from '@pharos/shared/components/ui-extension';
import {useUpdate} from '@/lib/data-provider';
import {useToast} from '@hooks/use-toast';
import {USER_PROVIDER_NAME} from '@providers/user-provider';
import {usePasswordRequirements} from '@features/user/hooks';

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  username: string;
}

export function ChangePasswordDialog({open, onOpenChange, username}: ChangePasswordDialogProps) {
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const {toast} = useToast();
  const {mutate: updatePassword, mutation} = useUpdate();
  const passwordRequirements = usePasswordRequirements();

  const isSubmitting = mutation.isPending;

  const validate = (): string => {
    if (!newPassword) return 'New password is required.';
    if (newPassword.length < passwordRequirements.min_length) {
      return passwordRequirements.requirementMsg;
    }
    if (
      passwordRequirements.min_uppercase &&
      (newPassword.match(/[A-Z]/g)?.length ?? 0) < passwordRequirements.min_uppercase
    ) {
      return passwordRequirements.requirementMsg;
    }
    if (
      passwordRequirements.min_lowercase &&
      (newPassword.match(/[a-z]/g)?.length ?? 0) < passwordRequirements.min_lowercase
    ) {
      return passwordRequirements.requirementMsg;
    }
    if (
      passwordRequirements.min_digits &&
      (newPassword.match(/[0-9]/g)?.length ?? 0) < passwordRequirements.min_digits
    ) {
      return passwordRequirements.requirementMsg;
    }
    if (
      passwordRequirements.min_special &&
      (newPassword.match(/[^A-Za-z0-9]/g)?.length ?? 0) < passwordRequirements.min_special
    ) {
      return passwordRequirements.requirementMsg;
    }
    if (newPassword !== confirmPassword) return 'Passwords do not match.';
    return '';
  };

  const handleSubmit = () => {
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }
    setError('');
    updatePassword(
      {
        resource: 'user-password',
        id: username,
        dataProviderName: USER_PROVIDER_NAME,
        values: {password: newPassword},
      },
      {
        onSuccess: () => {
          toast({description: 'Password has been changed.'});
          handleClose();
        },
        onError: (err: any) => {
          const msg = err?.response?.data?.error ?? err?.message ?? 'Failed to change password.';
          setError(msg);
        },
      },
    );
  };

  const handleClose = () => {
    setNewPassword('');
    setConfirmPassword('');
    setError('');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Change Password — {username}</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-2">
          <div>
            <Title variant="formLabel" htmlFor="cp-new">
              New password <span className="text-red-500">*</span>
            </Title>
            <Input
              id="cp-new"
              type="password"
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value.replace(/\s/g, ''));
                setError('');
              }}
              autoComplete="new-password"
            />
            {passwordRequirements.requirementMsg && (
              <p className="text-xs text-muted-foreground mt-1">{passwordRequirements.requirementMsg}</p>
            )}
          </div>
          <div>
            <Title variant="formLabel" htmlFor="cp-confirm">
              Confirm password <span className="text-red-500">*</span>
            </Title>
            <Input
              id="cp-confirm"
              type="password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value.replace(/\s/g, ''));
                setError('');
              }}
              autoComplete="new-password"
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!newPassword || !confirmPassword || isSubmitting}
          >
            <KeyRound className="h-4 w-4 mr-1" />
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
