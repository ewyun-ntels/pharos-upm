import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@pharos/shared/components/ui';
import { IconButton } from '@pharos/shared/components/ui-extension';
import { Undo2 } from 'lucide-react';
import { useToast } from '@hooks/use-toast';
import { useDashboardData, useDashboardActions } from '@features/dashboard/hooks/use-dashboard-store';
import JsonEditor from '@features/dashboard/components/DashboardHeader/JsonEditor';
import type { DashboardConfig } from '@pharos/shared/types/dashboard';

export function JsonEditorTab() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { dashboard } = useDashboardData();
  const { updateDashboard, saveDashboard } = useDashboardActions();

  const [jsonData, setJsonData] = useState<string>('');
  const [hasChanges, setHasChanges] = useState(false);
  const [resetKey, setResetKey] = useState(0);

  // Sync jsonData with dashboard changes
  useEffect(() => {
    if (dashboard) {
      setJsonData(JSON.stringify(dashboard, null, 2));
      setHasChanges(false);
    }
  }, [dashboard]);

  const handleSave = async () => {
    try {
      const parsedData: DashboardConfig = JSON.parse(jsonData);
      updateDashboard(parsedData);
      await saveDashboard();
      toast({ description: 'Dashboard saved successfully.' });
      setHasChanges(false);
    } catch (error) {
      console.error('Error:', error);
      toast({ description: 'Failed to save dashboard. Invalid JSON format.', variant: 'destructive' });
    }
  };

  const handleReset = () => {
    if (dashboard) {
      setJsonData(JSON.stringify(dashboard, null, 2));
      setHasChanges(false);
      setResetKey(prev => prev + 1); // Force remount JsonEditor
    }
  };

  const handleJsonChange = (value: string) => {
    setJsonData(value);
    setHasChanges(true);
  };

  return (
    <div className="px-5 py-1">
      <div className="max-w-7xl space-y-4">
        <div className='flex flex-row justify-between items-center mb-3'>
          <p className="text-sm text-muted-foreground">
            Edit dashboard configuration as JSON. Be careful when editing, invalid JSON will prevent saving.
          </p>
          <IconButton
            icon={<Undo2 />}
            variant="outline"
            onClick={handleReset}
            disabled={!hasChanges}
          >
            Reset
          </IconButton>
        </div>

        <div className="overflow-hidden">
          <JsonEditor
            key={resetKey}
            initialData={dashboard ? JSON.stringify(dashboard, null, 2) : undefined}
            onChange={handleJsonChange}
          />
        </div>

        <div className="flex gap-2 pt-4">
          <Button
            onClick={() => navigate(-1)}
            size="default"
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            size="default"
            variant="default"
            disabled={!hasChanges}
          >
            Save
          </Button>
        </div>
      </div>
    </div>
  );
}
