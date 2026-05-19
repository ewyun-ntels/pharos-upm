import React, { useState } from 'react';
import Form from '@components/rjsf';
import { Switch, Separator, Label } from '@pharos/shared/components/ui';
import { customizeValidator } from '@rjsf/validator-ajv8';
import ajvErrors from 'ajv-errors';
import ArrayFieldTemplate from '@components/rjsf/ArrayFieldTemplate';
import ArrayFieldItemTemplate from '@components/rjsf/ArrayFieldItemTemplate';
import BaseInputTemplate from '@components/rjsf/BaseInputTemplate';
import FieldTemplate from '@components/rjsf/FieldTemplate';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { Title } from '@pharos/shared/components/ui-extension';

type RoleMetadata = { key: string; displayName: string; description: string };
type RoleGroup = { group: string; displayName: string; roles: RoleMetadata[] };

type UserFormAndPermissionsProps = {
  formRef: React.Ref<any>;
  schema: any;
  uiSchema: any;
  /** RJSF formData(userinfo) 를 받아 처리 — RHF 값은 caller가 직접 관리 */
  handleSave: (rjsfFormData: any) => void;
  setSaveDisabled: (disabled: boolean) => void;
  userinfo: any;
  permissions: Record<string, boolean>;
  setPermissions: React.Dispatch<React.SetStateAction<Record<string, boolean>>>;
  openType: 'create' | 'edit';
  isSuperAdmin: boolean;
  roleGroups: RoleGroup[];
};

export const UserFormAndPermissions = ({
  formRef,
  schema,
  uiSchema,
  handleSave,
  setSaveDisabled,
  userinfo,
  permissions,
  setPermissions,
  openType,
  isSuperAdmin,
  roleGroups,
}: UserFormAndPermissionsProps) => {
  // formData는 RJSF가 관리하는 userinfo 전용 — permissions는 parent state가 source of truth
  const [formData, setFormData] = useState(() => {
    if (schema.properties?.userinfo) {
      return { userinfo: openType === 'edit' ? userinfo : {} };
    }
    return {};
  });

  const validator = customizeValidator();
  ajvErrors(validator.ajv);

  // permission 변경 핸들러 — parent state만 업데이트
  const handlePermissionChange = (permissionKey: string, isChecked: boolean) => {
    setPermissions((prev) => ({ ...prev, [permissionKey]: isChecked }));
    if (openType === 'edit') {
      setSaveDisabled(false);
    }
  };

  return (
    <div>
      <Form
        ref={formRef}
        schema={schema}
        validator={validator}
        templates={{
          FieldTemplate,
          ArrayFieldTemplate,
          ArrayFieldItemTemplate,
        }}
        widgets={{
          TextWidget: BaseInputTemplate,
          InputWidget: BaseInputTemplate,
          EmailWidget: BaseInputTemplate,
        }}
        uiSchema={uiSchema}
        formData={formData}
        liveValidate={true}
        onChange={({ formData }) => {
          setFormData(formData);
          if (openType === 'edit') {
            setSaveDisabled(false);
          }
        }}
        onSubmit={({ formData }) => {
          handleSave(formData);
        }}
        onError={() => {}}
        showErrorList={false}
        className="space-y-2"
      />
      <div className="mt-8 space-y-4">
        {roleGroups.map((group, groupIndex) => (
          <div key={group.group}>
            {groupIndex > 0 && <Separator className="my-4" />}
            <Title variant="h3">{group.displayName}</Title>
            <div className="space-y-4">
              {group.roles
                .filter(
                  (role) =>
                    isSuperAdmin || role.key !== PermissionKeysSchema.enum['role:super_admin'],
                )
                .map((role) => {
                  const switchId = `${group.group}-${role.key}`;
                  return (
                    <div key={role.key} className="space-y-1.5">
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={permissions[role.key] || false}
                          onCheckedChange={(checked) => handlePermissionChange(role.key, checked)}
                          id={switchId}
                        />
                        <Label htmlFor={switchId} className="text-sm font-medium cursor-pointer">
                          {role.displayName}
                        </Label>
                      </div>
                      {role.description && (
                        <p className="text-xs text-muted-foreground ml-11 leading-relaxed">
                          {role.description}
                        </p>
                      )}
                    </div>
                  );
                })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
