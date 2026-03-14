import { z } from 'zod';

export const adminSessionCommandSchema = z.object({
  action: z.enum(['refresh', 'rotate', 'revoke']),
  note: z.string().trim().max(240).optional(),
});
