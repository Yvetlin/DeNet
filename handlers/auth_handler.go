package handlers

import (
	"log"
	"net/http"

	"DeNet/database"
	"DeNet/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GenerateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithValidationError(w, "Invalid user ID format", "User ID must be a valid UUID")
		return
	}

	var exists bool
	err = database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil || !exists {
		utils.RespondWithNotFound(w, "User")
		return
	}

	token, err := utils.GenerateToken(userID)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
		"user_id": userID.String(),
	})
}

