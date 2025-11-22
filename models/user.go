package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Balance   int       `json:"balance" db:"balance"`
	ReferrerID *uuid.UUID `json:"referrer_id,omitempty" db:"referrer_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Task struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	TaskType    string    `json:"task_type" db:"task_type"`
	Points      int       `json:"points" db:"points"`
	CompletedAt time.Time `json:"completed_at" db:"completed_at"`
}

type UserStatus struct {
	User         *User   `json:"user"`
	CompletedTasks []Task `json:"completed_tasks"`
	TotalPoints  int    `json:"total_points"`
}

type LeaderboardEntry struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Balance   int       `json:"balance" db:"balance"`
	Rank      int       `json:"rank"`
}

