package models

import "time"

type Series struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Genre     string    `json:"genre"`
	Status    string    `json:"status"`
	Episodes  int       `json:"episodes"`
	Rating    float64   `json:"rating"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}