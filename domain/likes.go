package domain

import "time"

type Like struct {
	PostID string
	UserID string
	Date   time.Time
}
