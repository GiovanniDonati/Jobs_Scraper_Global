import { z } from "zod";

export const ScrapeParamsSchema = z.object({
  keywords: z.array(z.string()).min(1),
  location: z.string().optional(),
  sources: z.array(z.string()).optional(),
  page: z.number().int().positive().optional(),
  limit: z.number().int().positive().max(100).optional(),
});

export const JobSchema = z.object({
  id: z.string(),
  title: z.string(),
  company: z.string(),
  location: z.string(),
  url: z.string().url(),
  description: z.string().optional(),
  salary: z.string().optional(),
  source: z.string(),
  postedAt: z.string().optional(),
});

export const ScrapeResponseSchema = z.object({
  jobs: z.array(JobSchema),
  total: z.number(),
  cachedAt: z.string(),
});

export type ScrapeParams = z.infer<typeof ScrapeParamsSchema>;
export type Job = z.infer<typeof JobSchema>;
export type ScrapeResponse = z.infer<typeof ScrapeResponseSchema>;
