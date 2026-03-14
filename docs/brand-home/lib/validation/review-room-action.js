import { z } from 'zod';

export const reviewRoomActionSchema = z.object({
  item: z.string().trim().min(1).max(400),
  reason: z.string().trim().min(1).max(240),
  rule: z.string().trim().max(160).optional(),
  evidence: z.array(z.string().trim().min(1).max(160)).max(6).optional().default([]),
});
