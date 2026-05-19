/**
 * Notification Rule Form Types
 *
 * React Hook Form용 타입 정의
 * - Zod 스키마 기반
 * - Schema에서 자동 생성된 타입 활용
 */

import { z } from 'zod';
import {
  Snmpv2ConfigSchema,
  Snmpv3ConfigSchema,
  TransportSchema,
  TrapPduRuleSchema,
  AppendPduSchema,
} from '@pharos/shared/types/notification';

/**
 * SNMP Notification Rule Form Schema
 * - 공통 필드 + 버전별 config를 통합
 * - version 필드에 따라 snmpv2_config 또는 snmpv3_config 중 하나만 필수
 */
export const SnmpNotificationRuleFormSchema = z.discriminatedUnion('version', [
  // SNMP v2c
  z.object({
    name: z.string().min(1, 'Name is required'),
    description: z.string().optional(),
    target: z.string().min(1, 'Target is required'),
    port: z.coerce.number().min(1).max(65535).default(162),
    transport: TransportSchema.default('udp'),
    version: z.literal(2),
    snmpv2_config: Snmpv2ConfigSchema,
    value_oid: z.string().optional(),
    description_oid: z.string().optional(),
    status_oid: z.string().optional(),
    exclude_filter: z.record(z.string(), z.any()).optional(),
    trap_pdu_rule: z.array(TrapPduRuleSchema).optional().default([]),
    append_pdu: z.array(AppendPduSchema).optional().default([]),
  }),
  // SNMP v3
  z.object({
    name: z.string().min(1, 'Name is required'),
    description: z.string().optional(),
    target: z.string().min(1, 'Target is required'),
    port: z.coerce.number().min(1).max(65535).default(162),
    transport: TransportSchema.default('udp'),
    version: z.literal(3),
    snmpv3_config: Snmpv3ConfigSchema,
    value_oid: z.string().optional(),
    description_oid: z.string().optional(),
    status_oid: z.string().optional(),
    exclude_filter: z.record(z.string(), z.any()).optional(),
    trap_pdu_rule: z.array(TrapPduRuleSchema).optional().default([]),
    append_pdu: z.array(AppendPduSchema).optional().default([]),
  }),
]);

export type SnmpNotificationRuleForm = z.infer<typeof SnmpNotificationRuleFormSchema>;
