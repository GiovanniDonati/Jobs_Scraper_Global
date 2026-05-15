import { db } from "../../../db/client";
import { accounts } from "../../../db/schema";
import { DB } from "../../../db/types/types";
import { OAuthProfile } from "../../types/auth.types";

type CreateAccountParams = {
  userId: string;
  provider: string;
  profile: OAuthProfile;
};

export async function createAccount(
  { userId, provider, profile }: CreateAccountParams,
  tx: DB = db,
): Promise<void> {
  try {
    await tx.insert(accounts).values({
      userId,
      provider,
      providerAccountId: profile.id,
      accessToken: profile.access_token ?? null,
      refreshToken: profile.refresh_token ?? null,
      expiresAt: profile.expires_at
        ? new Date(profile.expires_at * 1000)
        : null,
    });
  } catch (err: any) {
    if (err.code !== "23505") {
      throw err;
    }
  }
}
