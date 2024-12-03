package utils

import (
	"encoding/json"
	"net/http"
)

// Response 統一的API響應結構
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONResponse 發送JSON響應
func JSONResponse(w http.ResponseWriter, status int, success bool, data interface{}, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: success,
		Data:    data,
		Error:   err,
	}

	json.NewEncoder(w).Encode(response)
}

// SuccessResponse 發送成功響應
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, true, data, "")
}

// ErrorResponse 發送錯誤響應
func ErrorResponse(w http.ResponseWriter, status int, err string) {
	JSONResponse(w, status, false, nil, err)
}

// ValidationError 發送驗證錯誤響應
func ValidationError(w http.ResponseWriter, err string) {
	ErrorResponse(w, http.StatusBadRequest, err)
}

// ServerError 發送服務器錯誤響應
func ServerError(w http.ResponseWriter, err string) {
	ErrorResponse(w, http.StatusInternalServerError, err)
}

// UnauthorizedError 發送未授權錯誤響應
func UnauthorizedError(w http.ResponseWriter) {
	ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
}
