import { z } from "zod";
import { ErrorCodeSchema } from "./error-codes.js";

export const ValidationErrorSchema = z.object({
  code: z.string(),
  field: z.string().optional(),
  message: z.string(),
});

export type ValidationError = z.infer<typeof ValidationErrorSchema>;

export const apiErrorResponseSchema = z.object({
  success: z.literal(false),
  message: z.string(),
  code: z.string(),
  errors: z.array(ValidationErrorSchema).optional(),
});
export type ApiErrorResponseSchema = z.infer<typeof apiErrorResponseSchema>;
export type ApiErrorResponse = z.infer<typeof apiErrorResponseSchema>;

export type ApiResponse<T = unknown> = {
  success: boolean;
  message?: string;
  code?: z.infer<typeof ErrorCodeSchema>;
  errors?: ValidationError[];
  data?: T;
  meta?: PaginationMeta;
};

export const PaginationMetaSchema = z.object({
  total: z.number(),
  page: z.number(),
  limit: z.number(),
  total_pages: z.number(),
  offset: z.number(),
});

export type PaginationMeta = z.infer<typeof PaginationMetaSchema>;

export function createSuccessResponseSchema() {
  return z.object({
    success: z.literal(true),
    message: z.string(),
  });
}

export function createObjectResponseSchema<T extends z.ZodTypeAny>(
  dataSchema: T,
) {
  return z.object({
    success: z.literal(true),
    message: z.string(),
    data: dataSchema,
  });
}

export function createListResponseSchema<T extends z.ZodTypeAny>(
  itemSchema: T,
) {
  return z.object({
    success: z.literal(true),
    message: z.string(),
    data: z.array(itemSchema),
    meta: PaginationMetaSchema.optional(),
  });
}

export type ApiSuccessResponse<T = unknown> = {
  success: true;
  message: string;
  data?: T;
  meta?: PaginationMeta;
};
