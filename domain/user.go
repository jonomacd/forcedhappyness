package domain

import "time"

type User struct {
	ID           string
	Name         string
	Email        string
	RegisterDate time.Time
	Avatar       string
}
