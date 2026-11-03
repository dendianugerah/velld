package response

import (
	"encoding/json"
	"net/http"
)

type Pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

type Response struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func SendSuccess(w http.ResponseWriter, message string, data interface{}) {
	SendSuccessWithStatus(w, http.StatusOK, message, data)
}

// SendSuccessWithStatus lets handlers return a non-200 success code, such as
// 202 Accepted for an async operation or 422 for a completed-but-failed check.
func SendSuccessWithStatus(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Message: message,
		Data:    data,
	})
}

func SendPaginatedSuccess(w http.ResponseWriter, message string, data interface{}, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: message,
		Data:    data,
		Pagination: &Pagination{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

func SendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Message: message,
	})
}
