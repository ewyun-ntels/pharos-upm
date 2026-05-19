import * as z from "zod";

// Type of notification rule

export const NotificationTypeSchema = z.enum([
    "email",
    "slack",
    "snmp",
    "teams",
    "webhook",
]);
export type NotificationType = z.infer<typeof NotificationTypeSchema>;

// SNMP value type

export const ValueTypeSchema = z.enum([
    "counter32",
    "counter64",
    "gauge32",
    "objectIdentifier",
    "OctetString",
    "opaqueDouble",
    "opaqueFloat",
    "uinteger32",
]);
export type ValueType = z.infer<typeof ValueTypeSchema>;

// Authentication protocol

export const AuthProtocolSchema = z.enum([
    "md5",
    "sha",
    "sha224",
    "sha256",
    "sha384",
    "sha512",
]);
export type AuthProtocol = z.infer<typeof AuthProtocolSchema>;

// Privacy protocol

export const PrivProtocolSchema = z.enum([
    "aes",
    "aes192",
    "aes256",
    "des",
]);
export type PrivProtocol = z.infer<typeof PrivProtocolSchema>;

// Transport protocol

export const TransportSchema = z.enum([
    "tcp",
    "udp",
]);
export type Transport = z.infer<typeof TransportSchema>;

export const NotificationRuleSchema = z.object({
    "notification_type": NotificationTypeSchema,
    "rule": z.record(z.string(), z.any()),
    "timestamp": z.coerce.date().optional(),
    "updated_at": z.coerce.date().optional(),
});
export type NotificationRule = z.infer<typeof NotificationRuleSchema>;

export const AppendPduSchema = z.object({
    "oid": z.string(),
    "value": z.any(),
    "value_type": ValueTypeSchema,
});
export type AppendPdu = z.infer<typeof AppendPduSchema>;

export const Snmpv2ConfigSchema = z.object({
    "community": z.string(),
});
export type Snmpv2Config = z.infer<typeof Snmpv2ConfigSchema>;

export const Snmpv3ConfigSchema = z.object({
    "auth_passphrase": z.string().optional(),
    "auth_protocol": AuthProtocolSchema.optional(),
    "authoritative_engine_id": z.string().optional(),
    "priv_passphrase": z.string().optional(),
    "priv_protocol": PrivProtocolSchema.optional(),
    "username": z.string(),
});
export type Snmpv3Config = z.infer<typeof Snmpv3ConfigSchema>;

export const TrapPduRuleSchema = z.object({
    "label_name": z.string(),
    "oid": z.string(),
    "value_type": ValueTypeSchema.optional(),
});
export type TrapPduRule = z.infer<typeof TrapPduRuleSchema>;

export const QueryNotificationRuleSchema = z.object({
    "append_pdu": z.array(AppendPduSchema).optional(),
    "description": z.string().optional(),
    "description_oid": z.string().optional(),
    "exclude_filter": z.record(z.string(), z.any()).optional(),
    "id": z.string().optional(),
    "name": z.string(),
    "port": z.number().optional(),
    "snmpv2_config": Snmpv2ConfigSchema.optional(),
    "snmpv3_config": Snmpv3ConfigSchema.optional(),
    "status_oid": z.string().optional(),
    "target": z.string(),
    "transport": TransportSchema.optional(),
    "trap_pdu_rule": z.array(TrapPduRuleSchema).optional(),
    "value_oid": z.string().optional(),
    "version": z.number().optional(),
});
export type QueryNotificationRule = z.infer<typeof QueryNotificationRuleSchema>;
