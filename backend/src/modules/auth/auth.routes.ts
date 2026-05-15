import { Router } from "express";
import { UsersService } from "../users/users.service";
import { AuthController } from "./auth.controller";
import { AuthService } from "./auth.service";

const router = Router();

const usersService = new UsersService();
const authService = new AuthService();
const authController = new AuthController(authService);

router.get("/:provider/url", (req, res) => authController.getUrl(req, res));
router.get("/:provider/callback", (req, res) =>
  authController.callback(req, res),
);

export { router as authRoutes };
