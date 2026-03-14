import { z } from 'zod';

export const adminSourceDraftLaneSchema = z.object({
  draft_observation_id: z.string().trim().min(1).max(120),
  channel: z.enum(['threads']).optional().default('threads'),
});
