import { Request, Response } from "express";
import { OAuthProfile } from "../types/auth.types";
import { UsersService } from "./users.service";

export class UsersController {
  constructor(private readonly userService: UsersService) {}

  async me(req: Request, res: Response) {
    try {
      // Supondo que você terá um middleware de auth que coloca o user no req
      const userId = req.headers["x-user-id"] as string;

      const user = await this.userService.getUserById(userId);
      if (!user)
        return res.status(404).json({ error: "Usuário não encontrado" });

      return res.json(user);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  }

  async update(req: Request, res: Response) {
    try {
      const userId = req.headers["x-user-id"] as string;
      const data = req.body;

      const updatedUser = await this.userService.updateProfile(userId, data);
      return res.json(updatedUser);
    } catch (error: any) {
      return res.status(400).json({ error: error.message });
    }
  }

  // Rota para o seu fluxo de login/OAuth
  async handleOAuth(req: Request, res: Response) {
    try {
      const { provider, profile }: { provider: string; profile: OAuthProfile } =
        req.body;

      // Aqui usamos aquela sua lógica poderosa de findOrCreate
      const user = await this.userService.getOrCreateFromOAuth(
        provider,
        profile,
      );

      return res.status(200).json(user);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  }
}
