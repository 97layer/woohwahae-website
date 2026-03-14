import { z } from 'zod';

const quickworkRoleSchema = z.enum(['implementer', 'verifier', 'planner', 'designer']).default('implementer');
const notesSchema = z.array(z.string().trim().min(1).max(120)).max(6).optional().default([]);

export const adminJobsActionSchema = z.discriminatedUnion('action', [
  z.object({
    action: z.literal('assign'),
    summary: z.string().trim().min(1).max(240),
    role: quickworkRoleSchema,
    allowed_paths: z.array(z.string().trim().min(1).max(240)).max(12).optional().default([]),
    payload_json: z.string().trim().max(4000).optional().default(''),
  }),
  z.object({
    action: z.literal('heartbeat'),
    limit: z.coerce.number().int().min(1).max(6).optional().default(1),
    dispatch: z.boolean().optional().default(true),
  }),
  z.object({
    action: z.literal('dispatch'),
    job_id: z.string().trim().min(1).max(120),
  }),
  z.object({
    action: z.literal('pause'),
    job_id: z.string().trim().min(1).max(120),
    notes: notesSchema,
  }),
  z.object({
    action: z.literal('cancel'),
    job_id: z.string().trim().min(1).max(120),
    notes: notesSchema,
  }),
  z.object({
    action: z.literal('promote'),
    limit: z.coerce.number().int().min(1).max(6).optional().default(1),
    dispatch: z.boolean().optional().default(true),
  }),
]);
