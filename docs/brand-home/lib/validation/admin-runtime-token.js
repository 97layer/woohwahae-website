import { z } from 'zod';

export const adminRuntimeTokenSchema = z.object({
  token: z.string().trim().min(1).max(512),
});
