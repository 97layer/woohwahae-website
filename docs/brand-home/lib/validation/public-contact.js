import { z } from 'zod';

export const publicContactSchema = z.object({
  name: z.string().trim().min(1).max(80),
  email: z.email().max(120),
  message: z.string().trim().min(8).max(2000),
});
