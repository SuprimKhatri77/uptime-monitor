package types

import "github.com/suprimkhatri77/uptime-monitor/api/internal/constants"

type AppError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Code    string          `json:"code,omitempty"`
	Errors  []AppError      `json:"errors,omitempty"`
	Data    any             `json:"data,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}

func Success(message string, data any) APIResponse {
	resp := APIResponse{
		Success: true,
		Message: message,
	}
	if data != nil {
		resp.Data = data
	}
	return resp
}

func SuccessWithMeta(message string, data any, meta PaginationMeta) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	}
}

func Error(message, code string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Code:    code,
	}
}

func ValidationErrorResponse(message string, errors []AppError) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Code:    constants.ValidationFailed,
		Errors:  errors,
	}
}
