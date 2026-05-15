import { eq } from "drizzle-orm";
import { db } from "../../db/client";
import { userPreferences, users } from "../../db/schema";
import { DB } from "../../db/types/types";
import { OAuthProfile } from "../types/auth.types";
import { UpdateProfileData, User } from "../types/user.types";
import { createUser } from "./functions/createUser";
import { findOrCreateUser } from "./functions/findOrCreateUser";

export class UsersService {
  constructor(private readonly tx: DB = db) {}

  async getOrCreateFromOAuth(provider: string, profile: OAuthProfile) {
    return findOrCreateUser({ provider, profile });
  }

  async getUserById(id: string): Promise<User | undefined> {
    return this.tx.query.users.findFirst({
      where: (u, { eq }) => eq(u.id, id),
    });
  }

  async updateProfile(userId: string, data: UpdateProfileData): Promise<User> {
    const result = await this.tx
      .update(users)
      .set({
        ...data,
        updatedAt: new Date(),
      })
      .where(eq(users.id, userId))
      .returning();

    if (!result[0]) {
      throw new Error(`Usuário ${userId} não encontrado`);
    }

    // Quando começar a usar o Valkey, aqui é o lugar de invalidar o cache:
    // await this.valkey.del(`user:${userId}`);

    return result[0];
  }

  async setupNewUser(profile: OAuthProfile) {
    return this.tx.transaction(async (tx) => {
      const user = await createUser({ profile }, tx);
      await tx.insert(userPreferences).values({ userId: user.id });
      return user;
    });
  }
}
