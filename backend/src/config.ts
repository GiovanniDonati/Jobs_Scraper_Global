export interface Viewport {
  width: number;
  height: number;
}

export interface AppConfig {
  headless: boolean;
  waitBetweenSearchesMs: number;
  pageTimeoutMs: number;
  maxPagesPerKeyword: number;
  viewport: Viewport;
  outputFile: string;
  pdfFile: string;
  searchLocation: string;
  searchGeoId: string;
  searchLanguage: string;
  remoteOnly: boolean;
  jobTypes: string;
  timeFilter: string;
  keywords: string[];
  keywordsStorageMode: "redis" | "env";
  cacheTtlMs: number;
  databaseUrl: string;
  redisUrl: string;
  redisKeyPrefix: string;
}

const DEFAULT_KEYWORDS = [
  "Java",
  "JavaScript",
  "React",
  "Node.js",
  "Mulesoft",
  "Desenvolvedor Junior",
  "Desenvolvedor Pleno",
];

function parseBoolean(value: string | undefined, fallback: boolean): boolean {
  if (value === undefined) return fallback;
  const normalized = value.trim().toLowerCase();
  if (["1", "true", "yes", "y", "on"].includes(normalized)) return true;
  if (["0", "false", "no", "n", "off"].includes(normalized)) return false;
  return fallback;
}

function parseNumber(value: string | undefined, fallback: number): number {
  if (value === undefined) return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function getKeywordsStorageMode(): "redis" | "env" {
  const mode = String(process.env.KEYWORDS_STORAGE_MODE ?? "redis")
    .trim()
    .toLowerCase();
  return mode === "env" ? "env" : "redis";
}

function normalizeKeywords(keywords: string[]): string[] | null {
  if (!Array.isArray(keywords)) return null;
  const result = [
    ...new Set(keywords.map((k) => String(k ?? "").trim()).filter(Boolean)),
  ];
  return result.length ? result : null;
}

function parseKeywordsFromEnv(value: string | undefined): string[] | null {
  return normalizeKeywords(
    String(value ?? "")
      .split(",")
      .map((k) => k.trim())
      .filter(Boolean),
  );
}

function parseKeywords(value: string | undefined): string[] {
  return parseKeywordsFromEnv(value) ?? DEFAULT_KEYWORDS;
}

function parseTimeFilter(value: string | undefined, fallback: string): string {
  if (!value) return fallback;
  const normalized = value.trim();
  return /^r\d+$/.test(normalized) ? normalized : fallback;
}

export function getConfig(): AppConfig {
  return {
    headless: parseBoolean(process.env.HEADLESS, false),
    waitBetweenSearchesMs: parseNumber(
      process.env.WAIT_BETWEEN_SEARCHES_MS,
      5000,
    ),
    pageTimeoutMs: parseNumber(process.env.PAGE_TIMEOUT_MS, 10000),
    maxPagesPerKeyword: parseNumber(process.env.MAX_PAGES_PER_KEYWORD, 5),
    viewport: {
      width: parseNumber(process.env.VIEWPORT_WIDTH, 1280),
      height: parseNumber(process.env.VIEWPORT_HEIGHT, 800),
    },
    outputFile: process.env.OUTPUT_FILE ?? "output/vagas_remoto.xlsx",
    pdfFile: process.env.PDF_FILE ?? "output/vagas_remoto.pdf",
    searchLocation: process.env.SEARCH_LOCATION ?? "Brasil",
    searchGeoId: process.env.SEARCH_GEO_ID ?? "106057199",
    searchLanguage: process.env.SEARCH_LANGUAGE ?? "pt",
    remoteOnly: parseBoolean(process.env.REMOTE_ONLY, true),
    jobTypes: process.env.JOB_TYPES ?? "C,F",
    timeFilter: parseTimeFilter(process.env.TIME_FILTER, "r604800"),
    keywords: parseKeywords(process.env.SEARCH_KEYWORDS),
    keywordsStorageMode: getKeywordsStorageMode(),
    cacheTtlMs: parseNumber(process.env.CACHE_TTL_MS, 10 * 60 * 1000),
    databaseUrl: process.env.DATABASE_URL?.trim() ?? "",
    redisUrl: process.env.REDIS_URL?.trim() ?? "",
    redisKeyPrefix: process.env.REDIS_KEY_PREFIX?.trim() ?? "vagas-full",
  };
}
