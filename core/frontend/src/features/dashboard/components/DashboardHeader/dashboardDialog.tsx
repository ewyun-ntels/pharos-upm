import React, {useState, useRef} from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {useToast} from '@hooks/use-toast';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {dashboardProvider, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import JsonEditor from './JsonEditor';
import {Upload, X} from 'lucide-react';

interface ReportDialogProps {
  open: boolean;
  onOpenChange: () => void;
  data: any;
  onSave?: (newData: any) => void;
  onSuccess?: () => void;
}

export function DashboardDialog({open, onOpenChange, data, onSave, onSuccess}: ReportDialogProps) {
  const {toast} = useToast();
  const {id, ...rest} = data ?? {};
  const {createDashboard, updateDashboard, saveDashboard, _loadDashboard} = useDashboardStore();
  const currentDashboardId = useDashboardStore((state) => state.id);

  const [jsonData, setJsonData] = useState<any>(null);
  const [importMode, setImportMode] = useState<'file' | 'json'>('file');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isImporting, setIsImporting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const onSubmit = async () => {
    try {
      const parsedData = JSON.parse(jsonData);

      if (data) {
        if (currentDashboardId !== id) {
          await _loadDashboard(id);
        }
        updateDashboard(parsedData);
        await saveDashboard();
      } else {
        await createDashboard(parsedData);
      }
      toast({description: 'Successfully saved.'});
      onSuccess?.();
      onOpenChange();
    } catch (error) {
      console.error('Error:', error);
      toast({description: 'Failed to save dashboard', variant: 'destructive'});
    }
  };

  const handleSave = () => {
    if (jsonData && onSave) {
      onSave(jsonData);
      onOpenChange();
      setJsonData(null);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    setSelectedFiles(files);
  };

  const handleRemoveFile = (index: number) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const handleFileImport = async () => {
    if (selectedFiles.length === 0) {
      toast({description: 'Please select files to import', variant: 'destructive'});
      return;
    }

    setIsImporting(true);
    try {
      // 현재 존재하는 대시보드 제목 목록 가져오기
      const existingTitles = new Set<string>();
      const result = await dashboardProvider.getList({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        pagination: {mode: 'off'},
      });

      const dashboards = result.data || [];
      dashboards.forEach((d: any) => {
        if (d.config?.title) existingTitles.add(d.config.title);
      });

      // Import 중 추가되는 제목도 추적
      const importedTitles = new Set<string>();
      let successCount = 0;

      for (const file of selectedFiles) {
        try {
          const text = await file.text();
          const configData = JSON.parse(text);

          // Validation
          if (!configData || typeof configData !== 'object') {
            throw new Error('Invalid dashboard format');
          }

          if (!configData.title && !configData.panels) {
            throw new Error('Missing required dashboard fields (title or panels)');
          }

          // 중복 제목 체크 및 번호 추가
          const originalTitle = configData.title || 'Untitled Dashboard';
          let finalTitle = originalTitle;
          let counter = 1;

          while (existingTitles.has(finalTitle) || importedTitles.has(finalTitle)) {
            finalTitle = `${originalTitle} (imported ${counter})`;
            counter++;
          }

          // displayName도 동일하게 처리
          const originalDisplayName = configData.displayName || originalTitle;
          const finalDisplayName = counter > 1
            ? `${originalDisplayName} (imported ${counter - 1})`
            : originalDisplayName;

          const finalData = {
            ...configData,
            title: finalTitle,
            displayName: finalDisplayName,
          };

          await dashboardProvider.create({
            resource: DASHBOARD_RESOURCES.DASHBOARD,
            variables: finalData,
          });

          importedTitles.add(finalTitle);
          existingTitles.add(finalTitle);
          successCount++;
        } catch (error) {
          console.error(`Failed to import ${file.name}:`, error);
          toast({
            description: `Failed to import ${file.name}: ${error instanceof Error ? error.message : 'Unknown error'}`,
            variant: 'destructive',
          });
        }
      }

      if (successCount > 0) {
        toast({description: `Successfully imported ${successCount} dashboard(s)`});
        onSuccess?.();
        onOpenChange();
        setSelectedFiles([]);
      }
    } catch (error) {
      console.error('Import error:', error);
      toast({description: 'Failed to import dashboards', variant: 'destructive'});
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={() => {
        onOpenChange();
        setJsonData(null);
        setSelectedFiles([]);
        setImportMode('file');
      }}
    >
      <DialogContent className="max-w-none" style={{maxWidth: '70rem'}}>
        <DialogHeader className="mr-2">
          <DialogTitle className="break-all">
            {data ? rest?.displayName : 'Import Dashboard'}
          </DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        {/* Import 모드만 Tabs 표시 */}
        {!data && (
          <Tabs value={importMode} onValueChange={(v) => setImportMode(v as 'file' | 'json')}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="file">File Import</TabsTrigger>
              <TabsTrigger value="json">JSON Editor</TabsTrigger>
            </TabsList>

            <TabsContent value="file" className="space-y-4">
              <div className="space-y-2">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".json"
                  multiple
                  onChange={handleFileSelect}
                  className="hidden"
                  id="file-upload"
                />
                <label htmlFor="file-upload">
                  <Button
                    type="button"
                    variant="outline"
                    className="w-full"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <Upload className="w-4 h-4 mr-2" />
                    Select JSON Files
                  </Button>
                </label>
              </div>

              {selectedFiles.length > 0 && (
                <div className="space-y-2">
                  <div className="text-sm font-medium">
                    Selected Files ({selectedFiles.length}):
                  </div>
                  <div className="max-h-60 overflow-y-auto space-y-1">
                    {selectedFiles.map((file, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-2 bg-secondary rounded text-sm"
                      >
                        <span className="truncate flex-1">{file.name}</span>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveFile(index)}
                        >
                          <X className="w-4 h-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <DialogFooter>
                <Button
                  onClick={handleFileImport}
                  size="sm"
                  variant="default"
                  disabled={selectedFiles.length === 0 || isImporting}
                >
                  {isImporting ? 'Importing...' : `Import ${selectedFiles.length} file(s)`}
                </Button>
              </DialogFooter>
            </TabsContent>

            <TabsContent value="json">
              <JsonEditor
                initialData={undefined}
                onChange={(value) => {
                  setJsonData(value);
                }}
              />
              <DialogFooter>
                <Button onClick={onSubmit} size="sm" variant="default" disabled={!jsonData}>
                  Save
                </Button>
              </DialogFooter>
            </TabsContent>
          </Tabs>
        )}

        {/* Edit 모드는 JSON Editor만 */}
        {data && (
          <>
            <JsonEditor
              initialData={rest}
              onChange={(value) => {
                setJsonData(value);
              }}
            />
            <DialogFooter>
              <Button
                onClick={onSave ? handleSave : onSubmit}
                size="sm"
                variant="default"
                disabled={!jsonData}
              >
                {onSave ? 'Apply' : 'Save'}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
