import { Router } from "express";
import { UsersController } from "./users.controller";
import { UsersService } from "./users.service";

const router = Router();

const userService = new UsersService();
const userController = new UsersController(userService);

router.get("/me", (req, res) => userController.me(req, res));
router.patch("/profile", (req, res) => userController.update(req, res));
router.post("/oauth", (req, res) => userController.handleOAuth(req, res));

export { router as userRoutes };
