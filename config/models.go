package config

import "time"

type UserContact struct {
	ID               int       `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	MiddleName       string    `json:"middle_name"`
	BirthDate        time.Time `json:"birth_date"`
	TelegramUsername string    `json:"telegram_username"`
	Age              int
}

// Глобальная map для хранения состояний пользователей
var MapUserStateData = make(map[int64]*UserStateData)

type UserStateData struct {
	State            string
	FirstName        string
	LastName         string
	MiddleName       string
	BirthDate        time.Time
	TelegramUsername string
}

// Состояния
const (
	StateWaitingFIO      = "waiting_fio"
	StateWaitingDate     = "waiting_date"
	StateWaitingUsername = "waiting_username"
)
