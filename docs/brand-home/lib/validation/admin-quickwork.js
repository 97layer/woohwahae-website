import { z } from 'zod';

const quickworkRoleSchema = z.enum(['implementer', 'verifier', 'planner', 'designer']).default('implementer');

export const adminQuickworkSchema = z.object({
  summary: z.string().trim().min(1).max(240),
  role: quickworkRoleSchema,
  allowed_paths: z.array(z.string().trim().min(1).max(240)).max(12).optional().default([]),
  payload_json: z.string().trim().max(4000).optional().default(''),
});
