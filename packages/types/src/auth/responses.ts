import { z } from "zod";
import {
  createObjectResponseSchema,
  createSuccessResponseSchema,
} from "../api/response.js";
import { UserSchema } from "../user/user.js";

export const HealthResponseSchema = createSuccessResponseSchema();
export type HealthResponse = z.infer<typeof HealthResponseSchema>;

export const AuthUserResponseSchema = createObjectResponseSchema(UserSchema);
export type AuthUserResponse = z.infer<typeof AuthUserResponseSchema>;

export const AuthSuccessResponseSchema = createSuccessResponseSchema();
export type AuthSuccessResponse = z.infer<typeof AuthSuccessResponseSchema>;

export const MeResponseSchema = createObjectResponseSchema(UserSchema);
export type MeResponse = z.infer<typeof MeResponseSchema>;
