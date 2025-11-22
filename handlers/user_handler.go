package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"DeNet/database"
	"DeNet/models"
	"DeNet/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetUserStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithValidationError(w, "Invalid user ID format", "User ID must be a valid UUID")
		return
	}

	var user models.User
	err = database.DB.QueryRow(
		`SELECT id, username, email, balance, referrer_id, created_at, updated_at 
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.ReferrerID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		utils.RespondWithNotFound(w, "User")
		return
	}
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	rows, err := database.DB.Query(
		`SELECT id, user_id, task_type, points, completed_at 
		 FROM tasks WHERE user_id = $1 ORDER BY completed_at DESC`,
		userID,
	)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.UserID, &task.TaskType, &task.Points, &task.CompletedAt); err != nil {
			log.Printf("Error scanning task: %v", err)
			utils.RespondWithInternalError(w, err)
			return
		}
		tasks = append(tasks, task)
	}

	var totalPoints int
	err = database.DB.QueryRow(
		`SELECT COALESCE(SUM(points), 0) FROM tasks WHERE user_id = $1`,
		userID,
	).Scan(&totalPoints)
	if err != nil {
		totalPoints = 0
	}

	status := models.UserStatus{
		User:          &user,
		CompletedTasks: tasks,
		TotalPoints:   totalPoints,
	}

	respondWithJSON(w, http.StatusOK, status)
}

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // по умолчанию топ 10
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 10
		}
	}

	rows, err := database.DB.Query(
		`SELECT id, username, balance 
		 FROM users 
		 ORDER BY balance DESC 
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}
	defer rows.Close()

	var leaderboard []models.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Balance); err != nil {
			log.Printf("Error scanning leaderboard entry: %v", err)
			utils.RespondWithInternalError(w, err)
			return
		}
		entry.Rank = rank
		leaderboard = append(leaderboard, entry)
		rank++
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"leaderboard": leaderboard,
	})
}

func CompleteTask(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		TaskType string `json:"task_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithValidationError(w, "Invalid request body", "Request body must be valid JSON")
		return
	}

	if req.TaskType == "" {
		utils.RespondWithValidationError(w, "task_type is required", "Field 'task_type' cannot be empty")
		return
	}

	points := getTaskPoints(req.TaskType)
	if points == 0 {
		utils.RespondWithValidationError(w, "Unknown task type", "Valid task types: telegram_subscribe, twitter_follow, referral_code, email_verify, profile_complete")
		return
	}

	var taskExists bool
	err = database.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM tasks WHERE user_id = $1 AND task_type = $2)`,
		userID, req.TaskType,
	).Scan(&taskExists)
	if err != nil {
		log.Printf("Error checking task: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}
	if taskExists {
		utils.RespondWithConflict(w, "Task already completed")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}
	defer tx.Rollback()

	taskID := uuid.New()
	_, err = tx.Exec(
		`INSERT INTO tasks (id, user_id, task_type, points, completed_at) 
		 VALUES ($1, $2, $3, $4, NOW())`,
		taskID, userID, req.TaskType, points,
	)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	_, err = tx.Exec(
		`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`,
		points, userID,
	)
	if err != nil {
		log.Printf("Error updating balance: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	if req.TaskType == "referral_code" {
		var referrerID uuid.UUID
		err = tx.QueryRow(`SELECT referrer_id FROM users WHERE id = $1`, userID).Scan(&referrerID)
		if err == nil && referrerID != uuid.Nil {
			referrerBonus := points / 2
			tx.Exec(
				`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`,
				referrerBonus, referrerID,
			)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Task completed successfully",
		"task_id": taskID,
		"points":  points,
	})
}

func SetReferrer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithValidationError(w, "Invalid user ID format", "User ID must be a valid UUID")
		return
	}

	var req struct {
		ReferrerID string `json:"referrer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithValidationError(w, "Invalid request body", "Request body must be valid JSON")
		return
	}

	if req.ReferrerID == "" {
		utils.RespondWithValidationError(w, "referrer_id is required", "Field 'referrer_id' cannot be empty")
		return
	}

	referrerID, err := uuid.Parse(req.ReferrerID)
	if err != nil {
		utils.RespondWithValidationError(w, "Invalid referrer ID format", "Referrer ID must be a valid UUID")
		return
	}

	var userExists bool
	err = database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists)
	if err != nil || !userExists {
		utils.RespondWithNotFound(w, "User")
		return
	}

	var referrerExists bool
	err = database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, referrerID).Scan(&referrerExists)
	if err != nil || !referrerExists {
		utils.RespondWithNotFound(w, "Referrer")
		return
	}

	if userID == referrerID {
		utils.RespondWithValidationError(w, "User cannot refer themselves", "A user cannot set themselves as their own referrer")
		return
	}

	var currentReferrerID *uuid.UUID
	err = database.DB.QueryRow(`SELECT referrer_id FROM users WHERE id = $1`, userID).Scan(&currentReferrerID)
	if err != nil {
		log.Printf("Error checking referrer: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}
	if currentReferrerID != nil {
		utils.RespondWithConflict(w, "User already has a referrer")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE users SET referrer_id = $1, updated_at = NOW() WHERE id = $2`,
		referrerID, userID,
	)
	if err != nil {
		log.Printf("Error setting referrer: %v", err)
		utils.RespondWithInternalError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Referrer set successfully",
	})
}

func getTaskPoints(taskType string) int {
	taskPoints := map[string]int{
		"telegram_subscribe": 50,
		"twitter_follow":     50,
		"referral_code":      100,
		"email_verify":       25,
		"profile_complete":   30,
	}
	return taskPoints[taskType]
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

