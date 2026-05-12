import { logWarn } from "../logger";
import type { ScrapeParams, ScrapeResponse } from "./types/scraper.types";
import {
  ScrapeParamsSchema,
  ScrapeResponseSchema,
} from "./types/scraper.types";

export type { ScrapeParams, ScrapeResponse };

const GO_SCRAPER_URL = process.env.GO_SCRAPER_URL ?? "http://localhost:8081";

export async function searchJobs(
  params: ScrapeParams,
): Promise<ScrapeResponse> {
  const validatedParams = ScrapeParamsSchema.parse(params);

  const res = await fetch(`${GO_SCRAPER_URL}/scrape`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(validatedParams),
  });

  if (!res.ok) {
    throw new Error(`Go scraper: ${res.status} ${res.statusText}`);
  }

  const raw = await res.json();
  const parsed = ScrapeResponseSchema.safeParse(raw);

  if (!parsed.success) {
    logWarn("Go scraper: resposta fora do contrato", {
      error: parsed.error.message,
    });
    throw new Error("Go scraper: resposta invalida");
  }

  return parsed.data;
}
