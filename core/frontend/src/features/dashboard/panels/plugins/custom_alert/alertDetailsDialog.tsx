import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import AlertBadge from '@components/badge/alert-badge';
import LabelBadge from '@components/badge/label-badge';
import {convertToLocalTime} from '@pharos/shared/lib/unitUtils';

interface AlertDetailsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dialogTitle: string;
  alertData: any;
}

export default function AlertDetailsDialog({
  dialogTitle,
  open,
  onOpenChange,
  alertData,
}: AlertDetailsDialogProps) {
  if (!alertData) return null;

  const Field = ({label, children}: {label: string; children: React.ReactNode}) => (
    <div className="flex flex-row gap-5 min-h-5">
      <h4 className="font-medium text-sm text-muted-foreground min-w-24">{label}</h4>
      {children}
    </div>
  );

  const GridRow = ({children}: {children: React.ReactNode}) => (
    <div className="grid grid-cols-1 py-2 border-b">{children}</div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{dialogTitle}</DialogTitle>
        </DialogHeader>

        <div>
          <GridRow>
            <Field label="Name">
              <p className="text-sm">{alertData.labels?.name || '-'}</p>
            </Field>
          </GridRow>

          <GridRow>
            <Field label="Time">
              <p className="text-sm">{convertToLocalTime(alertData.timestamp)}</p>
            </Field>
          </GridRow>

          <GridRow>
            <Field label="Severity">
              <AlertBadge id={alertData.severity} />
            </Field>
          </GridRow>
          <GridRow>
            <Field label="Information">
              <div className="flex flex-wrap gap-2">
                <LabelBadge labels={alertData.labels} enableCopy={true} />
              </div>
            </Field>
          </GridRow>
          <GridRow>
            <Field label="Description">
              <p className="text-sm">{alertData.description}</p>
            </Field>
          </GridRow>
        </div>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)} size="sm" variant="outline" type="button">
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
