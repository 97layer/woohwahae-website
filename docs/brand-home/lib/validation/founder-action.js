import { z } from 'zod';

const notesSchema = z.array(z.string().trim().min(1).max(120)).max(6).optional().default([]);

export const founderStartSchema = z.object({
  flow_id: z.string().trim().min(1).max(120),
  work_item_id: z.string().trim().min(1).max(120),
  approval_id: z.string().trim().min(1).max(120),
  title: z.string().trim().min(1).max(160),
  intent: z.string().trim().min(1).max(240),
  notes: notesSchema,
});

export const founderApproveSchema = z.object({
  flow_id: z.string().trim().min(1).max(120),
  notes: notesSchema,
});

export const founderReleaseSchema = z.object({
  flow_id: z.string().trim().min(1).max(120),
  release_id: z.string().trim().min(1).max(120),
  deploy_id: z.string().trim().min(1).max(120),
  target: z.string().trim().max(80).optional().default('vm'),
  channel: z.string().trim().max(80).optional().default('cockpit'),
  notes: notesSchema,
});

export const founderRollbackSchema = z.object({
  flow_id: z.string().trim().min(1).max(120),
  rollback_id: z.string().trim().min(1).max(120),
  notes: notesSchema,
});
