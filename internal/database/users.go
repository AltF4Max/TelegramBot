package database

import (
	"TelegramBot/config"
	"fmt"
	"time"
)

// Проверяет есть ли userID в БД
func UserExists(userID int64) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE user_id = ?"
	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	return count > 0, err
}

// Проверяет у кого сегодня др
func GetTodayBirthdays() ([]config.UserContact, error) {
	today := time.Now()
	query := `
        SELECT id, first_name, last_name, middle_name, birth_date, telegram_username 
        FROM user_contacts 
        WHERE MONTH(birth_date) = ? AND DAY(birth_date) = ?
    `

	rows, err := DB.Query(query, today.Month(), today.Day())
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var users []config.UserContact
	for rows.Next() {
		var user config.UserContact
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.MiddleName, &user.BirthDate, &user.TelegramUsername); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}

		// Вычисляем возраст сразу после сканирования
		age := today.Year() - user.BirthDate.Year()
		if today.YearDay() < user.BirthDate.YearDay() {
			age--
		}
		user.Age = age

		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации: %w", err)
	}

	return users, nil
}

// AddUserContact добавляет нового пользователя в базу данных
func AddUserContact(firstName, lastName, middleName, telegramUsername string, birthDate time.Time) (bool, error) {
	// Сначала проверяем существует ли username
	exists, err := isTelegramUsernameExists(telegramUsername)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки username: %w", err)
	}
	if exists {
		return true, fmt.Errorf("username %s уже существует", telegramUsername)
	}

	query := `
        INSERT INTO user_contacts 
        (first_name, last_name, middle_name, birth_date, telegram_username) 
        VALUES (?, ?, ?, ?, ?)
    `

	_, err = DB.Exec(query, firstName, lastName, middleName, birthDate, telegramUsername)
	if err != nil {
		return false, fmt.Errorf("ошибка при добавлении пользователя: %w", err)
	}

	return false, nil
}

func isTelegramUsernameExists(username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_contacts WHERE telegram_username = ?`
	err := DB.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки: %w", err)
	}
	return count > 0, nil
}
