import { useTableColumns } from '@pharos/shared/hooks/table-columns';
import { SelectBox } from '@pharos/shared/components/ui-extension/select/box';
import { Badge } from '@pharos/shared/components/ui';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import { Button } from '@pharos/shared/components/ui';
import { Checkbox } from '@pharos/shared/components/ui';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@pharos/shared/components/ui';
import { useToast } from '@hooks/use-toast';
import { useDelete, useList, useUpdate } from '@/lib/data-provider';
import { Trash2 } from '@pharos/shared/components';
import { useEffect, useState } from 'react';
import { DataTable } from '@pharos/shared/components/ui-extension';
import { ActionSchema } from '@pharos/shared/types/dashboard';
import { getCoreRowModel, useReactTable } from '@tanstack/react-table';

interface DialogProps {
  open: boolean;
  onOpenChange: () => void;
  dialogTitle: string;
  dashboardId?: string;
  onSave?: (newData: any) => void;
}

interface PermissionData {
  subject: string;
  object?: string;
  action: string;
}

export function PermissionDialog({ dialogTitle, open, onOpenChange, dashboardId }: DialogProps) {
  const { toast } = useToast();
  const [id] = useState<string | undefined>(dashboardId);
  const [publicChecked, setPublicChecked] = useState(false);
  const [publicPermission, setPublicPermission] = useState<string>();
  const [addPermissionData, setAddPermissionData] = useState<PermissionData>({
    subject: '',
    action: ActionSchema.enum.viewer,
  });

  const [permissionList, setPermissionList] = useState<PermissionData[]>([]);

  const { query: { data, refetch } } = useList({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.DASHBOARD_PERMISSION.replace('{id}', id as string),
    queryOptions: {
      retry: false,
      enabled: !!id,
    },
  });

  const { query: { data: userList } } = useList({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.USER,
    queryOptions: {
      retry: false,
    },
  });
  /* 
  useEffect(() => {
    if (!userList) return;

    console.log('User list data:', userList?.data.users);
  }, [userList]); */

  const { mutate: addMutate } = useUpdate({
    dataProviderName: USER_PROVIDER_NAME,
    resource: DASHBOARD_RESOURCES.DASHBOARD,
    id: id + '/permission',
    meta: {
      method: 'put',
    },
  });

  const { mutate: DeleteMutate } = useDelete();

  useEffect(() => {
    if (!data) return;

    const publicData = data.data.filter((item: any) => item.subject === '__public');
    if (publicData && publicData.length > 0) {
      setPublicChecked(true);
      setPublicPermission(publicData[0].action);
    }

    const nonPublicData: PermissionData[] = (data.data as PermissionData[]).filter(
      (item: any) => item.subject !== '__public',
    );

    nonPublicData.sort((a, b) => a.subject.localeCompare(b.subject));

    setPermissionList(nonPublicData);
  }, [data]);

  // 권한 옵션 상수화
  const PERMISSION_OPTIONS = [
    { value: ActionSchema.enum.viewer, label: 'Viewer' },
    { value: ActionSchema.enum.editor, label: 'Editor' },
    { value: ActionSchema.enum.owner, label: 'Owner' },
  ];

  // 공통 onSuccess 핸들러
  const handleSuccess = (msg: string, resetAdd?: boolean) => {
    toast({ description: msg });
    if (resetAdd) setAddPermissionData({ subject: '', action: ActionSchema.enum.viewer });
    refetch();
  };

  const callPublicUpdate = ({
    checked = publicChecked,
    action = publicPermission,
  }: {
    checked?: boolean;
    action?: string;
  }) => {
    if (!checked) return;
    addMutate(
      {
        values: {
          subject: '__public',
          object: id,
          action: action,
        },
      },
      {
        onSuccess: () => handleSuccess('Successfully saved.'),
        onError: (error) => {
          toast({ description: 'Error saving permission' });
          console.error('Error saving permission:', error);
          // refetch();
        },
      },
    );
  };

  const callPutMutate = () => {
    addMutate(
      {
        values: {
          ...addPermissionData,
          object: id,
        },
      },
      {
        onSuccess: () => handleSuccess('Successfully saved.', true),
        onError: (error) => {
          toast({ description: 'Error saving permission' });
          console.error('Error saving permission:', error);
          refetch();
        },
      },
    );
  };

  const onPublicPermissionChange = (value: string) => {
    setPublicPermission(value);
  };

  const handleDelete = (subject: string) => {
    if (!id) return;
    const encodedURI = encodeURI(id + `/permission/${subject}`);
    DeleteMutate(
      {
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: encodedURI,
      },
      {
        onSuccess: () => handleSuccess(`Permission deleted successfully`),
        onError: (error) => {
          toast({ description: `Error deleting permission` });
          console.error(`Error deleting permission for ${subject}:`, error);
          refetch();
        },
      },
    );
  };

  const { columns: rawColumns } = useTableColumns({
    columnConfigs: [
      {
        key: 'type',
        properties: { title: 'Type', size: 180 },
        customCell: () => {
          return (
            <div className="w-full flex justify-center items-center">
              <Badge variant="secondary">User</Badge>
            </div>
          );
        },
      },
      {
        key: 'subject',
        properties: { title: 'Name', size: 443 },
      },
      {
        key: 'action',
        properties: { title: 'Permission' },
        customCell: (value: string, row) => (
          <SelectBox
            size="full"
            options={PERMISSION_OPTIONS}
            onChange={(e) => {
              if (value !== e) {
                row.action = e;
                addMutate(
                  {
                    values: {
                      subject: row.subject,
                      object: id,
                      action: e,
                    },
                  },
                  {
                    onSuccess: () => toast({ description: 'Successfully saved.' }),
                  },
                );
              }
            }}
            value={row.action}
          />
        ),
      },
      {
        key: 'delete',
        properties: { title: 'Delete', size: 80, enableSorting: false },
        customCell: (_: string, row) => (
          <div className="flex justify-center">
            <Button variant="outline" onClick={() => handleDelete(row.subject)}>
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        ),
      },
    ],
  });

  const columns = rawColumns as import('@tanstack/react-table').ColumnDef<PermissionData, any>[];

  const table = useReactTable({
    data: permissionList,
    columns: columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent
          className="max-w-none "
          style={{ maxWidth: '60rem' }}
          onInteractOutside={(event) => event.preventDefault()}
        >
          <DialogHeader>
            <DialogTitle className="break-all">{dialogTitle}</DialogTitle>
            <DialogDescription className="hidden"></DialogDescription>
          </DialogHeader>
          <div
            className="overflow-y-auto"
            style={{ maxHeight: '35rem' }} // 원하는 최대 높이로 조절
          >
            <h3 className="pb-3 self-start"> Visibility</h3>

            <table
              className="w-full border-t border-b border-gray-300 border-collapse text-sm"
              style={{ tableLayout: 'fixed' }}
            >
              <colgroup>
                <col style={{ width: '5%' }} />
                <col style={{ width: '64%' }} />
                <col style={{ width: '31%' }} />
              </colgroup>
              <tbody>
                <tr>
                  <td className="border-y border-gray-300 px-4 py-2">
                    <Checkbox
                      checked={publicChecked}
                      onCheckedChange={(e: boolean) => {
                        setPublicChecked(e);
                        if (e) callPublicUpdate({ checked: e });
                        else handleDelete('__public');
                      }}
                    />
                  </td>
                  <td className="border-y border-r border-gray-300 px-4 py-2">Public</td>
                  <td className="border-y border-gray-300 px-4 py-2">
                    <div className="flex items-center justify-center h-full">
                      <SelectBox
                        size="full"
                        options={PERMISSION_OPTIONS}
                        onChange={(e) => {
                          onPublicPermissionChange(e);
                          callPublicUpdate({ action: e });
                        }}
                        value={publicPermission}
                      />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>

            <h3 className="py-3 self-start"> Permission</h3>
            {/* <TableDynamicTemplate
              columnDef={columns}
              data={permissionList}
              usePagination={false}
              tableHeight={250}
            /> */}
            <DataTable table={table} tableHeight={250} className="[&_tr>td]:border-b [&_tr>td]:border-border" tableWidthMode="fill-last" />

            <h3 className="pb-3 self-start"> Add a permission</h3>
            <table className="w-full border-t border-b border-gray-300 border-collapse text-sm">
              <colgroup>
                <col style={{ width: '20%' }} />
                <col style={{ width: '50%' }} />
                <col style={{ width: '22%' }} />
                <col style={{ width: '8%' }} />
              </colgroup>
              <tbody>
                <tr>
                  <td className="border-y border-r border-gray-300 px-4 py-2">
                    <div className="flex items-center justify-center h-full">
                      <SelectBox
                        size="full"
                        options={[
                          { value: 'user', label: 'User' },
                          /* {value: 'group', label: 'Group'},
                        {value: 'etc', label: 'ETC'}, */
                        ]}
                        // placeholder="Select a type"
                        autoSelectFirstOption={true}
                        onChange={() => { }}
                      />
                    </div>
                  </td>
                  <td className="border-y border-r border-gray-300 px-4 py-2">
                    {userList?.data && (
                      <SelectBox
                        size="full"
                        options={userList?.data
                          ?.filter((item: any) => !!item.name)
                          ?.map((item: any) => ({
                            value: item.name,
                            label: item.name,
                          }))}
                        autoSelectFirstOption={true}
                        onChange={(value) => {
                          if (value !== addPermissionData.subject) {
                            setAddPermissionData((prev) => ({
                              ...prev,
                              subject: value,
                            }));
                          }
                        }}
                        value={addPermissionData.subject}
                      />
                    )}
                    {/* <Input
                      onChange={(e) => {
                        setAddPermissionData((prev) => ({
                          ...prev,
                          subject: e.target.value,
                        }));
                      }}
                      value={addPermissionData.subject}
                    ></Input> */}
                  </td>
                  <td className="border-y border-r border-gray-300 px-4 py-2">
                    <SelectBox
                      size="full"
                      options={PERMISSION_OPTIONS}
                      value={addPermissionData.action}
                      onChange={(value: string) => {
                        setAddPermissionData((prev) => ({
                          ...prev,
                          action: value,
                        }));
                      }}
                    />
                  </td>
                  <td className="border-y border-gray-300 px-4 py-2">
                    <Button
                      className="px-3 py-1"
                      onClick={callPutMutate}
                      disabled={!addPermissionData.subject}
                    >
                      Save
                    </Button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <DialogFooter>
            <Button
              onClick={() => {
                onOpenChange();
              }}
              size="sm"
              variant="outline"
              disabled={false}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
