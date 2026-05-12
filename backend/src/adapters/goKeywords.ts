import { logWarn } from "../logger";

const GO_BASE_URL =
  process.env.GO_SCRAPER_URL?.trim() || "http://scraper-go:8080";

type Keyword = string;

export function normalizeKeywords(keywords: unknown): Keyword[] | null {
  if (!Array.isArray(keywords)) {
    return null;
  }

  return [
    ...new Set(
      keywords.map((item) => String(item ?? "").trim()).filter(Boolean),
    ),
  ];
}

export async function loadKeywords(
  fallback: Keyword[] = [],
): Promise<Keyword[]> {
  try {
    const response = await fetch(`${GO_BASE_URL}/api/keywords`);

    if (!response.ok) {
      return fallback;
    }

    const data = await response.json();

    return normalizeKeywords(data.keywords) ?? fallback;
  } catch (error) {
    logWarn("Erro ao carregar keywords do Go", {
      error: (error as Error).message,
    });

    return fallback;
  }
}

export async function saveKeywords(
  keywords: unknown,
): Promise<Keyword[] | null> {
  const normalized = normalizeKeywords(keywords);

  if (normalized === null) {
    return null;
  }

  const response = await fetch(`${GO_BASE_URL}/api/keywords`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      keywords: normalized,
    }),
  });

  if (!response.ok) {
    throw new Error(`Erro ao salvar keywords no Go`);
  }

  const data = await response.json();

  return normalizeKeywords(data.keywords);
}
