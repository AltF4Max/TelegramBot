package database

import (
	"TelegramBot/config"
	"fmt"
	"time"
)

// Проверяет есть ли userID в БД
func UserExists(userID int64) (bool, error) { //isTelegramUsernameExists
	query := "SELECT COUNT(*) FROM users WHERE user_id = ?"
	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count > 0, err
}

func isTelegramUsernameExists(username string) (bool, error) { // UserExists
	query := `SELECT COUNT(*) FROM user_contacts WHERE telegram_username = ?`
	var count int
	err := DB.QueryRow(query, username).Scan(&count)
	return count > 0, err
}

// Проверяет у кого Birthdays
func GetTodayBirthdays() ([]config.UserContact, error) {
	today := time.Now()
	query := `
        SELECT id, first_name, last_name, middle_name, birth_date, telegram_username 
        FROM user_contacts 
        WHERE MONTH(birth_date) = ? AND DAY(birth_date) = ?
    `

	rows, err := DB.Query(query, today.Month(), today.Day())
	if err != nil {
		return nil, fmt.Errorf("Ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var users []config.UserContact
	for rows.Next() {
		var user config.UserContact
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.MiddleName, &user.BirthDate, &user.TelegramUsername); err != nil {
			return nil, fmt.Errorf("Ошибка сканирования: %w", err)
		}

		// Вычисляем возраст сразу после сканирования
		age := today.Year() - user.BirthDate.Year()
		birthdayThisYear := time.Date(today.Year(), user.BirthDate.Month(), user.BirthDate.Day(), 0, 0, 0, 0, today.Location())
		if today.Before(birthdayThisYear) {
			age--
		}
		user.Age = age
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Ошибка итерации: %w", err)
	}
	return users, nil
}

// AddUserContact добавляет нового пользователя в базу данных
func AddUserContact(user *config.UserStateData) (bool, error) {
	// Сначала проверяем существует ли username
	exists, err := isTelegramUsernameExists(user.TelegramUsername)
	if err != nil {
		return false, fmt.Errorf("Ошибка проверки username: %w", err)
	}
	if exists {
		return true, fmt.Errorf("Username %s уже существует", user.TelegramUsername)
	}

	query := `
        INSERT INTO user_contacts 
        (first_name, last_name, middle_name, birth_date, telegram_username) 
        VALUES (?, ?, ?, ?, ?)
    `

	_, err = DB.Exec(query, user.FirstName, user.LastName, user.MiddleName, user.BirthDate, user.TelegramUsername)
	if err != nil {
		return false, fmt.Errorf("Ошибка при добавлении пользователя: %w", err)
	}

	return false, nil
}
