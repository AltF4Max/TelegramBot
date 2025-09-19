package main

import (
	"TelegramBot/config"
	"TelegramBot/internal/database"
	"TelegramBot/utils"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	database.Init()
	defer database.Close()

	// Получаем токен из переменных окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Panic("TELEGRAM_BOT_TOKEN не установлен")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	///
	chatID := int64(837127109)
	users, err := database.GetTodayBirthdays()
	if err != nil {
		log.Fatal("Ошибка:", err)
	}
	if len(users) != 0 {
		messageText := fmt.Sprintf("🎂 Сегодня у %d пользователя(ей) день рождения:\n\n", len(users))
		for i, user := range users {
			messageText += fmt.Sprintf("%d. %s %s - %d лет\n",
				i+1,
				user.FirstName,
				user.LastName,
				user.Age)
			if user.TelegramUsername != "" {
				messageText += fmt.Sprintf("   👤 @%s\n", user.TelegramUsername)
			}
			messageText += "\n" // отступ между пользователями
		}
		msg := tgbotapi.NewMessage(chatID, messageText)
		bot.Send(msg)
	}
	///
	// Устанавливаем режим отладки (опционально)
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Text != "" {
			exists, err := database.UserExists(update.Message.From.ID) //userID
			if err != nil {
				log.Printf("Ошибка проверки пользователя: %v", err)
				return
			}
			if exists {
				if update.Message.IsCommand() {
					switch update.Message.Command() {
					case "adduser":
						chatID := update.Message.Chat.ID
						// Если не существует - создаем, если существует - используем
						config.MapUserStateData[chatID] = &config.UserStateData{State: config.StateWaitingFIO}
						msg := tgbotapi.NewMessage(chatID, "Введите ФИО в формате:\nИван Иванов Иванович")
						bot.Send(msg)
						continue
					case "deleteuser":
						config.MapUserStateData[chatID] = &config.UserStateData{State: config.StateWaitingDeleteUsername}
						msg := tgbotapi.NewMessage(chatID, "Введите Telegram username (без @) пользователя для удаления в формате:\nНапример: ivanov_90")
						bot.Send(msg)
						continue
					case "showall":
						users, err := database.GetAllUsers()
						if err != nil {
							msg := tgbotapi.NewMessage(chatID, "Ошибка: "+err.Error())
							bot.Send(msg)
							continue
						}
						if len(users) == 0 {
							msg := tgbotapi.NewMessage(chatID, "📭 База данных пуста")
							bot.Send(msg)
							continue
						}
						message := "👥 Все пользователи:\n\n"
						for i, user := range users {
							message += fmt.Sprintf("   %d. %s %s %s\n", i+1, user.LastName, user.FirstName, user.MiddleName)
							message += fmt.Sprintf("   👤 @%s\n", user.TelegramUsername)
							message += fmt.Sprintf("   🎂 %s\n\n", user.BirthDate.Format("02.01.2006"))
						}
						msg := tgbotapi.NewMessage(chatID, message)
						bot.Send(msg)
						continue
					}

				}
				// Обработка состояний (если пользователь в процессе диалога)
				if _, exists := config.MapUserStateData[update.Message.Chat.ID]; exists {
					handleUserState(bot, update, config.MapUserStateData[chatID].State)
					continue
				}
				// Повторяет сообщение пользователя
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы сказали: "+update.Message.Text)
				if _, err := bot.Send(msg); err != nil {
					log.Println("Ошибка отправки сообщения:", err)
				}
			} else {
				assetPaths := config.NewAssetPaths()
				utils.SendRedGif(bot, update.Message, assetPaths)
			}
		}
	}
}

func handleUserState(bot *tgbotapi.BotAPI, update tgbotapi.Update, state string) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	switch state {
	// Сохраняем ФИО и запрашиваем дату
	case config.StateWaitingFIO:
		firstName, lastName, middleName, err := utils.SplitTextToThreeVars(text)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Неправильный формат ФИО. Введите Фамилия Имя Отчество")
			bot.Send(msg)
			return
		}
		config.MapUserStateData[chatID].FirstName = firstName
		config.MapUserStateData[chatID].LastName = lastName
		config.MapUserStateData[chatID].MiddleName = middleName
		config.MapUserStateData[chatID].State = config.StateWaitingDate

		msg := tgbotapi.NewMessage(chatID, "Теперь введите дату рождения в формате ГГГГ-ММ-ДД:\nНапример: 1990-05-15")
		bot.Send(msg)
		// Сохраняем дату и запрашиваем TelegramUsername
	case config.StateWaitingDate:
		birthDate, err := time.Parse("2006-01-02", text)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Неправильный формат даты. Введите ГГГГ-ММ-ДД:")
			bot.Send(msg)
			return
		}
		// Обрезает время, оставляя только дату
		today := time.Now().Truncate(24 * time.Hour)
		if birthDate.After(today) {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Неправильная дата. %s еще не наступил", text))
			bot.Send(msg)
			return
		}
		config.MapUserStateData[chatID].BirthDate = birthDate
		config.MapUserStateData[chatID].State = config.StateWaitingUsername

		msg := tgbotapi.NewMessage(chatID, "Теперь введите Telegram username (без @):\nНапример: ivanov_90")
		bot.Send(msg)
		// Сохраняем TelegramUsername и добавляем в БД
	case config.StateWaitingUsername:
		isValid, errorMsg := utils.IsValidUsername(text)
		if !isValid {
			msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
			bot.Send(msg)
			return
		}
		config.MapUserStateData[chatID].TelegramUsername = text

		exists, err := database.AddUserContact(config.MapUserStateData[chatID])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при добавлении")
			bot.Send(msg)
			return
		}
		if exists {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Пользователь c Telegram username: %s уже существует в БД", text))
			bot.Send(msg)
			return //
		} else {
			msg := tgbotapi.NewMessage(chatID, "✅ Контакт успешно добавлен!")
			bot.Send(msg)
		}
		delete(config.MapUserStateData, chatID)
	case config.StateWaitingDeleteUsername:
		isValid, errorMsg := utils.IsValidUsername(text)
		if !isValid {
			msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
			bot.Send(msg)
			return
		}
		exists, err := database.DeleteUserContact(text)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при удалении")
			bot.Send(msg)
			return
		}
		if exists {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Пользователь c Telegram username: %s не существует в БД", text))
			bot.Send(msg)
			return //
		} else {
			msg := tgbotapi.NewMessage(chatID, "✅ Контакт успешно удален!")
			bot.Send(msg)
		}
		delete(config.MapUserStateData, chatID)
	}
}
