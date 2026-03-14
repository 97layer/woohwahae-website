import { z } from 'zod';

const notesSchema = z.array(z.string().trim().min(1).max(120)).max(6).optional().default([]);
const sourceIdsSchema = z.array(z.string().trim().min(1).max(80)).max(6).optional().default([]);
const targetAccountSchema = z.enum(['97layer', 'woosunhokr', 'woohwahae']).optional().default('97layer');

export const adminBrandPublishSchema = z.object({
  channel: z.string().trim().min(1).max(40),
  target_account: targetAccountSchema,
  title: z.string().trim().min(1).max(160),
  body: z.string().trim().min(1).max(1200),
  topic_tag: z.string().trim().max(120).optional().default(''),
  source_ids: sourceIdsSchema,
  notes: notesSchema,
});
