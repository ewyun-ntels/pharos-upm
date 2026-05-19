import React, {useState, useEffect} from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {convertToLocalTime} from '@pharos/shared/lib/unitUtils';
import {Badge} from '@pharos/shared/components/ui';
import {useUser} from '@pharos/shared/features/auth';
import type {Role} from '@pharos/shared/types/user';

interface DialogProps {
  open: boolean;
  onOpenChange: () => void;
  dialogTitle: string;
}

export function MyInfoDialog({dialogTitle, open, onOpenChange}: DialogProps) {
  const authUser = useUser(); // Auth Store에서 사용자 정보 가져오기

  useEffect(() => {
    // Auth Store의 사용자 정보를 myInfo에 설정
    if (authUser) {
      setMyInfo(authUser);
    }
  }, [authUser]);

  const [myInfo, setMyInfo] = useState<any>(null);

  const Field = ({label, children}: {label: string; children: React.ReactNode}) => (
    <div className="flex flex-row gap-5 min-h-5">
      <h4 className="font-medium text-sm text-muted-foreground min-w-40 first-letter:uppercase">
        {label}
      </h4>
      {children}
    </div>
  );

  const GridRow = ({children}: {children: React.ReactNode}) => (
    <div className="grid grid-cols-1 py-2">{children}</div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{dialogTitle}</DialogTitle>
        </DialogHeader>
        <div>
          <GridRow>
            <Field label="Username">
              <p className="text-sm">{myInfo?.name || '-'}</p>
            </Field>
          </GridRow>

          <GridRow>
            <Field label="Password expiration">
              <p className="text-sm">
                {myInfo?.password_expired_at ? convertToLocalTime(myInfo.password_expired_at) : '-'}
              </p>
            </Field>
          </GridRow>

          <GridRow>
            <Field label="Role">
              {(() => {
                // roles 사용 (display_name이 포함됨)
                const roles = myInfo?.roles ?? [];

                if (roles.length > 0) {
                  return roles.map((roleItem: Role) => (
                    <Badge key={roleItem.role} variant="outline">
                      {roleItem.display_name}
                    </Badge>
                  ));
                } else {
                  return <Badge variant="outline">User</Badge>;
                }
              })()}
            </Field>
          </GridRow>
          {myInfo?.attributes?.info && Object.keys(myInfo.attributes.info).length > 0 && (
            <>
              <div className="pt-8 pb-1">
                <h4 className="text-md font-semibold">User fields</h4>
              </div>
              {Object.entries(myInfo.attributes.info).map(([key, value]) => {
                // 부서, 이름, 전화번호 등 커스텀 필드명을 친화적으로 변경 (원할 경우)
                let label = key;
                if (key.toLowerCase() === 'phone' || key.toLowerCase() === 'phonenumber' || key.toLowerCase() === 'tel') label = 'Phone Number';
                if (key.toLowerCase() === 'department') label = 'Department';
                if (key.toLowerCase() === 'name' || key.toLowerCase() === 'fullname') label = 'Name';
                
                return (
                  <GridRow key={key}>
                    <Field label={label}>
                      <p className="text-sm">{String(value) || '-'}</p>
                    </Field>
                  </GridRow>
                );
              })}
            </>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onOpenChange}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
