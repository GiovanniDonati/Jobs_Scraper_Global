import { randomBytes } from "crypto";
import { Request, Response } from "express";
import { getIronSession } from "iron-session";
import { z } from "zod";
import {
  AuthCallbackParamsSchema,
  OAuthProviderSchema,
} from "../types/auth.types.js";

import { AuthService } from "./auth.service.js";

interface SessionData {
  oauth_state?: string;
  userId?: string;
}

const sessionOptions = {
  password: process.env.SESSION_SECRET!,
  cookieName: "vagas_session",

  cookieOptions: {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    maxAge: 60 * 60 * 24 * 7,
  },
};

export class AuthController {
  constructor(private readonly authService: AuthService) {}

  async getUrl(req: Request, res: Response) {
    try {
      const session = await getIronSession<SessionData>(
        req,
        res,
        sessionOptions,
      );

      const provider = OAuthProviderSchema.parse(req.params.provider);

      const state = randomBytes(16).toString("hex");

      console.log("provider:", provider);
      console.log("generated oauth state:", state);

      session.oauth_state = state;

      await session.save();

      const url = await this.authService.getAuthUrl(provider, state);

      return res.json({ url });
    } catch (error) {
      if (error instanceof z.ZodError) {
        return res.status(400).json({
          error: "Provider inválido",
          details: error.message,
        });
      }

      return res.status(400).json({
        error: (error as Error).message,
      });
    }
  }

  async callback(req: Request, res: Response) {
    console.log("=== CALLBACK DEBUG ===");
    console.log("headers cookie:", req.headers.cookie);
    console.log("query:", req.query);

    const session = await getIronSession<SessionData>(req, res, sessionOptions);
    console.log("session oauth_state:", session.oauth_state);

    try {
      const callbackUrl = `${req.protocol}://${req.get("host")}${req.originalUrl}`;
      console.log("callbackUrl:", callbackUrl);

      const params = AuthCallbackParamsSchema.parse({
        provider: req.params.provider,
        code: req.query.code,
        state: req.query.state,
        callbackUrl,
      });
      console.log("received state:", params.state);

      if (!session.oauth_state) {
        console.log("OAuth state ausente na sessão");
        return res.status(400).json({
          error: "OAuth state ausente",
        });
      }

      if (session.oauth_state !== params.state) {
        console.log("OAuth state inválido");
        console.log("session:", session.oauth_state);
        console.log("received:", params.state);
        return res.status(400).json({
          error: "OAuth state inválido",
        });
      }

      delete session.oauth_state;
      await session.save();
      console.log("OAuth state validado com sucesso");

      const result = await this.authService.handleCallback({
        ...params,
        callbackUrl,
      });
      console.log("OAuth callback concluído com sucesso");

      return res.json(result);
    } catch (error) {
      console.error("OAuth callback error:", error);

      if (error instanceof z.ZodError) {
        return res.status(400).json({
          error: "Parâmetros de callback inválidos",
          details: error.format(),
        });
      }

      const message = error instanceof Error ? error.message : "Erro interno";

      return res.status(500).json({
        error: message,
      });
    }
  }
}
