/**
 * SNMP Rule Editor - Unified component with tabs for v2/v3
 *
 * - React Hook Form + zodResolver
 * - 공통 필드는 상단에 표시
 * - 버전별 필드는 탭으로 구분
 * - 탭 전환 시 version 필드 자동 업데이트
 */

'use client';

import React, { useState, useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button, Input, Label, Textarea, Tabs, TabsList, TabsTrigger, TabsContent } from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import { Bell, Network, Shield, ListTree } from 'lucide-react';
import type { QueryNotificationRule } from '@pharos/shared/types/notification';
import {
  SnmpNotificationRuleFormSchema,
  type SnmpNotificationRuleForm,
} from '@features/notification/types/form';
import { TrapPduRuleEditor } from './TrapPduRuleEditor';
import { AppendPduEditor } from './AppendPduEditor';

interface SnmpRuleEditorProps {
  initialVersion?: 'v2' | 'v3';
  initialData?: Partial<QueryNotificationRule>;
  onSubmit?: (data: SnmpNotificationRuleForm) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

const getDefaultValues = (data?: Partial<QueryNotificationRule>): SnmpNotificationRuleForm => {
  const version = data?.version === 3 ? 3 : 2;

  const commonDefaults = {
    name: data?.name || '',
    description: data?.description || '',
    target: data?.target || '',
    port: data?.port || 162,
    transport: data?.transport || 'udp',
    value_oid: data?.value_oid || '',
    description_oid: data?.description_oid || '',
    status_oid: data?.status_oid || '',
    trap_pdu_rule: data?.trap_pdu_rule || [],
    append_pdu: data?.append_pdu || [],
  };

  if (version === 2) {
    return {
      ...commonDefaults,
      version: 2,
      snmpv2_config: {
        community: data?.snmpv2_config?.community || 'public',
      },
    } as SnmpNotificationRuleForm;
  }

  return {
    ...commonDefaults,
    version: 3,
    snmpv3_config: {
      username: data?.snmpv3_config?.username || '',
      authoritative_engine_id: data?.snmpv3_config?.authoritative_engine_id || '',
      auth_protocol: data?.snmpv3_config?.auth_protocol || 'sha256',
      auth_passphrase: data?.snmpv3_config?.auth_passphrase || '',
      priv_protocol: data?.snmpv3_config?.priv_protocol || 'aes',
      priv_passphrase: data?.snmpv3_config?.priv_passphrase || '',
    },
  } as SnmpNotificationRuleForm;
};

export function SnmpRuleEditor({
  initialVersion = 'v2',
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: SnmpRuleEditorProps) {
  const [activeTab, setActiveTab] = useState<'v2' | 'v3'>(initialVersion);

  const {
    register,
    control,
    handleSubmit,
    watch,
    reset,
    formState: { errors },
  } = useForm<SnmpNotificationRuleForm>({
    resolver: zodResolver(SnmpNotificationRuleFormSchema) as any,
    defaultValues: getDefaultValues(initialData),
  });

  const currentVersion = watch('version');

  // 탭 전환 시 version 필드 업데이트
  useEffect(() => {
    const newVersion = activeTab === 'v2' ? 2 : 3;
    if (currentVersion !== newVersion) {
      const currentValues = watch();

      // 공통 필드 추출
      const {
        name, description, target, port, transport,
        value_oid, description_oid, status_oid,
        trap_pdu_rule, append_pdu
      } = currentValues;

      const commonFields = {
        name, description, target, port, transport,
        value_oid, description_oid, status_oid,
        trap_pdu_rule, append_pdu
      };

      // 버전별 config 설정
      const versionConfig = newVersion === 2
        ? { version: 2, snmpv2_config: { community: 'public' } }
        : {
            version: 3,
            snmpv3_config: {
              username: '',
              authoritative_engine_id: '',
              auth_protocol: 'sha256' as const,
              auth_passphrase: '',
              priv_protocol: 'aes' as const,
              priv_passphrase: '',
            }
          };

      reset({ ...commonFields, ...versionConfig } as SnmpNotificationRuleForm);
    }
  }, [activeTab, currentVersion, watch, reset]);

  const handleFormSubmit = (data: SnmpNotificationRuleForm) => {
    onSubmit?.(data);
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
      {/* Basic Information - 공통 필드 */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <Bell className="h-5 w-5" />
          Basic Information
        </h3>
        <div className="space-y-4">
          <div>
            <Label htmlFor="name" className="text-sm font-normal">
              Rule Name *
            </Label>
            <Input
              id="name"
              {...register('name')}
              placeholder="e.g., SNMP Trap Notification"
              className={errors.name ? 'border-red-500' : ''}
            />
            {errors.name && <p className="text-sm text-red-500 mt-1">{errors.name.message}</p>}
          </div>

          <div>
            <Label htmlFor="description" className="text-sm font-normal">
              Description
            </Label>
            <Textarea
              id="description"
              {...register('description')}
              placeholder="Describe this notification rule..."
              rows={3}
            />
          </div>
        </div>
      </div>

      {/* Network Configuration - 공통 필드 */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <Network className="h-5 w-5" />
          Network Configuration
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label htmlFor="target" className="text-sm font-normal">
              Target (IP Address) *
            </Label>
            <Input
              id="target"
              {...register('target')}
              placeholder="192.168.1.100"
              className={errors.target ? 'border-red-500' : ''}
            />
            {errors.target && <p className="text-sm text-red-500 mt-1">{errors.target.message}</p>}
          </div>

          <div>
            <Label htmlFor="port" className="text-sm font-normal">
              Port
            </Label>
            <Input
              id="port"
              type="number"
              {...register('port')}
              placeholder="162"
              className={errors.port ? 'border-red-500' : ''}
            />
            {errors.port && <p className="text-sm text-red-500 mt-1">{errors.port.message}</p>}
          </div>
        </div>

        <div>
          <Label htmlFor="transport" className="text-sm font-normal">
            Transport
          </Label>
          <Controller
            name="transport"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select transport" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="udp">UDP</SelectItem>
                  <SelectItem value="tcp">TCP</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>
      </div>

      {/* SNMP Version Tabs - 버전별 필드 */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">SNMP Version & Configuration</h3>
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as 'v2' | 'v3')}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="v2">SNMP v2c</TabsTrigger>
            <TabsTrigger value="v3">SNMP v3</TabsTrigger>
          </TabsList>

          {/* SNMP v2c Tab */}
          <TabsContent value="v2" className="space-y-4 mt-4">
            <div>
              <Label htmlFor="community" className="text-sm font-normal">
                Community String *
              </Label>
              <Input
                id="community"
                {...register('snmpv2_config.community' as any)}
                placeholder="public"
                className={
                  currentVersion === 2 && (errors as any).snmpv2_config?.community
                    ? 'border-red-500'
                    : ''
                }
              />
              {currentVersion === 2 && (errors as any).snmpv2_config?.community && (
                <p className="text-sm text-red-500 mt-1">
                  {(errors as any).snmpv2_config.community.message}
                </p>
              )}
              <p className="text-xs text-muted-foreground mt-1">
                SNMP community string for authentication
              </p>
            </div>
          </TabsContent>

          {/* SNMP v3 Tab */}
          <TabsContent value="v3" className="space-y-4 mt-4">
            <div className="space-y-4">
              <div className="flex items-center gap-2 text-sm font-semibold">
                <Shield className="h-4 w-4" />
                Security Configuration
              </div>

              <div>
                <Label htmlFor="username" className="text-sm font-normal">
                  Username *
                </Label>
                <Input
                  id="username"
                  {...register('snmpv3_config.username' as any)}
                  placeholder="snmpuser"
                  className={
                    currentVersion === 3 && (errors as any).snmpv3_config?.username
                      ? 'border-red-500'
                      : ''
                  }
                />
                {currentVersion === 3 && (errors as any).snmpv3_config?.username && (
                  <p className="text-sm text-red-500 mt-1">
                    {(errors as any).snmpv3_config.username.message}
                  </p>
                )}
              </div>

              <div>
                <Label htmlFor="authoritative_engine_id" className="text-sm font-normal">
                  Authoritative Engine ID (Hex)
                </Label>
                <Input
                  id="authoritative_engine_id"
                  {...register('snmpv3_config.authoritative_engine_id' as any)}
                  placeholder="8000137003"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Optional. Format: 8000137003 or 0x8000137003 or 80:00:13:70:03
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="auth_protocol" className="text-sm font-normal">
                    Authentication Protocol
                  </Label>
                  <Controller
                    name={'snmpv3_config.auth_protocol' as any}
                    control={control}
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select protocol" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="md5">MD5</SelectItem>
                          <SelectItem value="sha">SHA</SelectItem>
                          <SelectItem value="sha224">SHA-224</SelectItem>
                          <SelectItem value="sha256">SHA-256</SelectItem>
                          <SelectItem value="sha384">SHA-384</SelectItem>
                          <SelectItem value="sha512">SHA-512</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>

                <div>
                  <Label htmlFor="auth_passphrase" className="text-sm font-normal">
                    Authentication Passphrase
                  </Label>
                  <Input
                    id="auth_passphrase"
                    type="password"
                    {...register('snmpv3_config.auth_passphrase' as any)}
                    placeholder="Enter passphrase"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="priv_protocol" className="text-sm font-normal">
                    Privacy Protocol
                  </Label>
                  <Controller
                    name={'snmpv3_config.priv_protocol' as any}
                    control={control}
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select protocol" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="des">DES</SelectItem>
                          <SelectItem value="aes">AES</SelectItem>
                          <SelectItem value="aes192">AES-192</SelectItem>
                          <SelectItem value="aes256">AES-256</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>

                <div>
                  <Label htmlFor="priv_passphrase" className="text-sm font-normal">
                    Privacy Passphrase
                  </Label>
                  <Input
                    id="priv_passphrase"
                    type="password"
                    {...register('snmpv3_config.priv_passphrase' as any)}
                    placeholder="Enter passphrase"
                  />
                </div>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* OID Configuration - 공통 필드 (Optional) */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">OID Configuration (Optional)</h3>
        <div className="space-y-4">
          <div>
            <Label htmlFor="value_oid" className="text-sm font-normal">
              Value OID
            </Label>
            <Input
              id="value_oid"
              {...register('value_oid')}
              placeholder="1.3.6.1.4.1.12345.1.1"
            />
          </div>

          <div>
            <Label htmlFor="description_oid" className="text-sm font-normal">
              Description OID
            </Label>
            <Input
              id="description_oid"
              {...register('description_oid')}
              placeholder="1.3.6.1.4.1.12345.1.2"
            />
          </div>

          <div>
            <Label htmlFor="status_oid" className="text-sm font-normal">
              Status OID
            </Label>
            <Input
              id="status_oid"
              {...register('status_oid')}
              placeholder="1.3.6.1.4.1.12345.1.3"
            />
          </div>
        </div>
      </div>

      {/* PDU Configuration - 공통 필드 (Optional) */}
      <div className="space-y-4">
        <h3 className="flex items-center gap-2 text-lg font-semibold">
          <ListTree className="h-5 w-5" />
          PDU Configuration (Optional)
        </h3>
        <div className="space-y-6">
          <Controller
            name={'trap_pdu_rule' as any}
            control={control}
            render={({ field }) => (
              <TrapPduRuleEditor value={field.value || []} onChange={field.onChange} />
            )}
          />

          <Controller
            name={'append_pdu' as any}
            control={control}
            render={({ field }) => (
              <AppendPduEditor value={field.value || []} onChange={field.onChange} />
            )}
          />
        </div>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4 border-t">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </form>
  );
}
