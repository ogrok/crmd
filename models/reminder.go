package models

type Reminder struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Recurrence  string `json:"recurrence,omitempty"`
	Timestamp   int64  `json:"time"`
}