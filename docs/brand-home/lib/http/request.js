import { ZodError } from 'zod';

export async function readValidatedJson(request, schema) {
  let raw;
  try {
    raw = await request.json();
  } catch {
    return { ok: false, status: 400, error: 'invalid json body' };
  }

  try {
    const data = schema.parse(raw);
    return { ok: true, data };
  } catch (error) {
    if (error instanceof ZodError) {
      return {
        ok: false,
        status: 400,
        error: 'request validation failed',
        details: error.flatten(),
      };
    }
    throw error;
  }
}
