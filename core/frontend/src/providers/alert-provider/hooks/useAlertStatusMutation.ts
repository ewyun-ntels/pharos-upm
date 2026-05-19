import { useUpdate, type HttpError } from '@/lib/data-provider';
import { ALERT_RESOURCES } from '../types';
import { 
  AlertValueSchema,
  AlertStatusMaskRequestSchema,
  type AlertValue,
  type AlertStatusMaskRequest
} from '@pharos/shared/types/alert';

/**
 * Custom hook for updating alert status with Zod validation
 * 
 * @example
 * const { mutate } = useUpdateAlertStatus();
 * 
 * mutate('alert-123', {
 *   status: 'normal',
 *   severity: 'Normal'
 * });
 */
export function useUpdateAlertStatus() {
  const { mutate, ...rest } = useUpdate<AlertValue, HttpError, Partial<AlertValue>>();

  const updateWithValidation = (id: string, data: Partial<AlertValue>) => {
    // Partial validation - only validate provided fields
    const validated = AlertValueSchema.partial().parse(data);
    
    mutate({
      resource: ALERT_RESOURCES.STATUS,
      id,
      values: validated as Partial<AlertValue>,
    });
  };

  return {
    mutate: updateWithValidation,
    ...rest,
  };
}

/**
 * Custom hook for masking alert status with Zod validation
 * 
 * @example
 * const { mutate } = useMaskAlertStatus();
 * 
 * // Mask alert for 1 hour
 * mutate('alert-123', { mask: true });
 * 
 * // Unmask alert
 * mutate('alert-123', { mask: false });
 */
export function useMaskAlertStatus() {
  const { mutate, ...rest } = useUpdate<AlertValue, HttpError, AlertStatusMaskRequest>();

  const maskWithValidation = (id: string, data: AlertStatusMaskRequest) => {
    const validated = AlertStatusMaskRequestSchema.parse(data);
    
    mutate({
      resource: ALERT_RESOURCES.STATUS_MASK,
      id,
      values: validated,
    });
  };

  return {
    mutate: maskWithValidation,
    ...rest,
  };
}
