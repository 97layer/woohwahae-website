import { z } from 'zod';

const tagSchema = z.array(z.string().trim().min(1).max(60)).max(8).optional().default([]);
const routeSchema = z.enum(['97layer', 'woosunhokr', 'woohwahae']);

export const adminSourceIntakeSchema = z.object({
  intake_class: z.enum(['manual_drop', 'authorized_connector', 'public_collector']).optional().default('manual_drop'),
  policy_color: z.enum(['green', 'yellow', 'red']).optional().default('green'),
  title: z.string().trim().max(160).optional().default(''),
  url: z.union([z.literal(''), z.string().trim().url().max(400)]).optional().default(''),
  excerpt: z.string().trim().min(1).max(4000),
  founder_note: z.string().trim().max(800).optional().default(''),
  domain_tags: tagSchema,
  worldview_tags: tagSchema,
  suggested_routes: z.array(routeSchema).min(1).max(3).optional().default(['97layer']),
});
