import { z } from 'zod';

export const adminBrandPublishThreadsSchema = z.object({
  approval_id: z.string().trim().min(1).max(120),
});
