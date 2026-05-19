import React from 'react';
import {useNavigate, useSearchParams} from 'react-router-dom';
import {useOne} from '@/lib/data-provider';
import {PageHeader} from '@components/page-header';
import {UserEditor} from '@features/user';
import {USER_PROVIDER_NAME, USER_RESOURCES} from '@providers/user-provider';
import type {User} from '@pharos/shared/types/user';
import {ScrollArea} from '@pharos/shared/components/ui/scroll-area';

/**
 * User Edit Page
 *
 * Create new or edit existing user
 * - URL: /users/edit         (create new)
 * - URL: /users/edit?id=xxx  (edit existing, xxx = username)
 */
export default function UserEditPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const userId = searchParams.get('id');
  const openType = userId ? 'edit' : 'create';

  const {
    query: {data: userData, isLoading},
  } = useOne<User>({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.USER,
    id: userId ?? '',
    queryOptions: {
      enabled: !!userId,
      retry: false,
    },
  });

  const user = userData?.data;

  const handleSuccess = () => {
    navigate('/users');
  };

  const handleCancel = () => {
    navigate('/users');
  };

  if (userId && isLoading) {
    return (
      <main className="w-full h-full">
        <ScrollArea className="w-full h-full">
          <PageHeader title="Edit User" breadcrumbLabel={openType === 'create' ? 'New User' : 'Edit User'} />
          <section className="flex flex-col w-full px-5 pt-5 pb-5">
            <div className="flex items-center justify-center py-8">
              <p>Loading user...</p>
            </div>
          </section>
        </ScrollArea>
      </main>
    );
  }

  if (userId && !isLoading && !user) {
    return (
      <main className="w-full h-full">
        <ScrollArea className="w-full h-full">
          <PageHeader title="Edit User" breadcrumbLabel={openType === 'create' ? 'New User' : 'Edit User'} />
          <section className="flex flex-col w-full px-5 pt-5 pb-5">
            <div className="flex items-center justify-center py-8">
              <p>User not found</p>
            </div>
          </section>
        </ScrollArea>
      </main>
    );
  }

  return (
    <main className="w-full h-full">
      <ScrollArea className="w-full h-full">
        <PageHeader title={openType === 'create' ? 'New User' : 'Edit User'} breadcrumbLabel={openType === 'create' ? 'New User' : 'Edit User'} />
        <section className="flex flex-col w-full px-5 pt-5 pb-5">
          <UserEditor
            openType={openType}
            data={user}
            onSuccess={handleSuccess}
            onCancel={handleCancel}
          />
        </section>
      </ScrollArea>
    </main>
  );
}
