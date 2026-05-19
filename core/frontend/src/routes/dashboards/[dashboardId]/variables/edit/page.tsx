import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  useDashboardData,
  useDashboardAutoLoad,
  useDashboardActions,
} from '@features/dashboard/hooks/use-dashboard-store';
import { useDashboardNavigationGuard } from '@features/dashboard/hooks/use-dashboard-navigation-guard';
import {PageHeader} from '@components/page-header';
import { VariableEditor } from '@features/dashboard/variables/components/VariableEditor/VariableEditor';
import type { FilterConfig } from '@pharos/shared/types/dashboard';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import {
  Alert, AlertDescription, AlertTitle,
  AlertDialog, AlertDialogContent, AlertDialogHeader,
  AlertDialogTitle, AlertDialogDescription,
  AlertDialogFooter, AlertDialogAction,
} from '@pharos/shared/components/ui';
import { AlertTriangle } from 'lucide-react';

export default function DashboardVariableEditPage() {
  const navigate = useNavigate();
  const { dashboardId } = useParams<{ dashboardId: string }>();
  const [searchParams] = useSearchParams();
  const editVariableId = searchParams.get('id') ?? undefined;

  const [variableData, setVariableData] = useState<FilterConfig>({
    id: '',
    type: 'select',
    kind: 'headerL',
  });

  const [showAlert, setShowAlert] = useState(false);
  const [alertConfig, setAlertConfig] = useState({
    title: '',
    description: '',
    buttonName: 'Confirm',
    action: () => {},
  });

  useDashboardAutoLoad(dashboardId || '');
  useDashboardNavigationGuard(dashboardId);

  const { loading, filters } = useDashboardData();
  const { setFilters } = useDashboardActions();

  const showAlertDialog = (
    title: string,
    description: string,
    buttonName: string = 'Confirm',
    action: () => void = () => {},
  ) => {
    setAlertConfig({ title, description, buttonName, action });
    setShowAlert(true);
  };

  const backToList = () => {
    const params = new URLSearchParams(window.location.search);
    params.delete('id');
    const queryString = params.toString();
    const baseUrl = `/dashboards/${dashboardId}/settings`;
    const url = queryString ? `${baseUrl}?${queryString}#variables` : `${baseUrl}#variables`;
    navigate(url);
  };

  useEffect(() => {
    if (!editVariableId) {
      const params = new URLSearchParams(window.location.search);
      params.delete('id');
      const queryString = params.toString();
      const baseUrl = `/dashboards/${dashboardId}/settings`;
      const url = queryString ? `${baseUrl}?${queryString}#variables` : `${baseUrl}#variables`;
      navigate(url, { replace: true });
      return;
    }
    const variable = filters.find(f => f.id === editVariableId);
    if (variable) {
      setVariableData(variable);
    } else {
      setVariableData({ id: '', type: 'select', kind: 'headerL' });
    }
  }, [editVariableId, filters, dashboardId, navigate]);

  if (loading) return <main className="w-full h-full"><LoadingIndicator /></main>;
  if (!dashboardId) return (
    <main className="w-full h-full p-4">
      <Alert variant="destructive">
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>Dashboard ID is required.</AlertDescription>
      </Alert>
    </main>
  );

  const handleSubmit = (data: FilterConfig) => {
    if (!data.id?.trim()) {
      showAlertDialog('Input Error', 'ID is required. Please enter an ID.');
      return;
    }
    if (data.type === 'select' && !data.query?.trim()) {
      showAlertDialog('Input Error', 'Query is required for data-driven variable types. Please enter a query.');
      return;
    }

    if (editVariableId !== data.id) {
      if (filters.some(f => f.id === data.id)) {
        showAlertDialog('Duplicate Error', 'ID already exists. Please use a different ID.');
        return;
      }
      setFilters([...filters.filter(f => f.id !== editVariableId), data]);
    } else {
      setFilters(filters.map(f => f.id === editVariableId ? data : f));
    }
    backToList();
  };

  return (
    <main className="w-full h-full overflow-y-auto">
      <PageHeader title="Edit Variable" breadcrumbLabel="Edit Variable" />
      <section className="flex flex-col w-full px-5 pt-5 pb-5">
        <VariableEditor
          initialType={variableData.type}
          initialData={variableData}
          onSubmit={handleSubmit}
          onCancel={backToList}
          onChange={setVariableData}
          dashboardId={dashboardId}
        />
      </section>

      <AlertDialog open={showAlert} onOpenChange={setShowAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{alertConfig.title}</AlertDialogTitle>
            <AlertDialogDescription>{alertConfig.description}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={alertConfig.action}>{alertConfig.buttonName}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </main>
  );
}
