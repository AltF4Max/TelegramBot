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

	// Устанавливаем режим отладки (опционально)
	bot.Debug = true

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
				messageText += fmt.Sprintf("   👤 @%s\n", user.TelegramUsername) //update.Message.From.UserName
			}
			messageText += "\n" // отступ между пользователями
		}
		msg := tgbotapi.NewMessage(chatID, messageText)
		bot.Send(msg)
	}

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
					case "addusercontact":
						chatID := update.Message.Chat.ID

						// Если не существует - создаем, если существует - используем
						config.UserS_D[chatID] = &config.UserStateData{State: config.StateWaitingFIO}

						msg := tgbotapi.NewMessage(chatID, "Введите ФИО в формате:\nИван Иванов Иванович")
						bot.Send(msg)
						continue
					}
				}

				// Обработка состояний (если пользователь в процессе диалога)
				if _, exists := config.UserS_D[update.Message.Chat.ID]; exists {
					handleUserState(bot, update, config.UserS_D[chatID].State)
					continue
				}

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
		firstName, lastName, middleName, err := utils.SplitTextToThreeVars(text) //Добавить если больше 3 слов
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Неправильный формат ФИО. Введите Фамилия Имя Отчество") //
			bot.Send(msg)
			return
		}
		config.UserS_D[chatID].FirstName = firstName
		config.UserS_D[chatID].LastName = lastName
		config.UserS_D[chatID].MiddleName = middleName
		config.UserS_D[chatID].State = config.StateWaitingDate

		msg := tgbotapi.NewMessage(chatID, "Теперь введите дату рождения в формате ГГГГ-ММ-ДД:\nНапример: 1990-05-15")
		bot.Send(msg)
		// Сохраняем дату и запрашиваем TelegramUsername
	case config.StateWaitingDate:
		birthDate, err := time.Parse("2006-01-02", text) //Добавить проверку на дату
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Неправильный формат даты. Введите ГГГГ-ММ-ДД:")
			bot.Send(msg)
			return
		}

		config.UserS_D[chatID].BirthDate = birthDate
		config.UserS_D[chatID].State = config.StateWaitingUsername

		msg := tgbotapi.NewMessage(chatID, "Теперь введите Telegram username (без @):\nНапример: ivanov90")
		bot.Send(msg)
		// Сохраняем TelegramUsername и добавляем в БД
	case config.StateWaitingUsername:
		config.UserS_D[chatID].TelegramUsername = text //Можно удалить если поменять?

		exists, err := database.AddUserContact(config.UserS_D[chatID].FirstName, config.UserS_D[chatID].LastName, config.UserS_D[chatID].MiddleName, config.UserS_D[chatID].TelegramUsername, config.UserS_D[chatID].BirthDate) //Поменять + проверка на ру буквы
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при добавлении")
			bot.Send(msg)
			return
		}
		if exists {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Пользователь c Telegram username: %s уже существует в БД", config.UserS_D[chatID].TelegramUsername)) //
			bot.Send(msg)
			delete(config.UserS_D, chatID)
		}
		delete(config.UserS_D, chatID)
	}
}
