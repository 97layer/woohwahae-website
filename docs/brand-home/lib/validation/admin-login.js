import { z } from 'zod';

export const adminLoginSchema = z.object({
  token: z.string().trim().min(20).max(2048).optional(),
  password: z.string().min(4).max(512).optional(),
  next: z.string().trim().max(240).optional(),
}).superRefine((value, ctx) => {
  const token = typeof value.token === 'string' ? value.token.trim() : '';
  const password = typeof value.password === 'string' ? value.password.trim() : '';
  if (!token && !password) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'token or password is required',
      path: ['token'],
    });
  }
});
