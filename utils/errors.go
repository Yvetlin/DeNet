package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func RespondWithError(w http.ResponseWriter, code int, message string, errorCode string) {
	log.Printf("Error [%s]: %s (HTTP %d)", errorCode, message, code)
	
	response := ErrorResponse{
		Error: message,
		Code:  errorCode,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func RespondWithValidationError(w http.ResponseWriter, message string, details string) {
	log.Printf("Validation error: %s - %s", message, details)
	
	response := ErrorResponse{
		Error:   message,
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func RespondWithNotFound(w http.ResponseWriter, resource string) {
	RespondWithError(w, http.StatusNotFound, resource+" not found", "NOT_FOUND")
}

func RespondWithUnauthorized(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

func RespondWithConflict(w http.ResponseWriter, message string) {
	RespondWithError(w, http.StatusConflict, message, "CONFLICT")
}

func RespondWithInternalError(w http.ResponseWriter, err error) {
	log.Printf("Internal server error: %v", err)
	RespondWithError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
}

