import pino from "pino";

export const logger = pino({
  level: process.env.LOG_LEVEL ?? "info",
  timestamp: pino.stdTimeFunctions.isoTime,
  formatters: {
    level(label) {
      return { level: label };
    },
  },
});

export const logInfo = (msg: string, ctx?: object) =>
  logger.info(ctx ?? {}, msg);
export const logWarn = (msg: string, ctx?: object) =>
  logger.warn(ctx ?? {}, msg);
export const logError = (msg: string, ctx?: object) =>
  logger.error(ctx ?? {}, msg);
