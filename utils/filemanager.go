package utils

import (
	"TelegramBot/config"
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// fileExists - проверяет существование файла
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// SendRedGif - отправляет заранее подготовленный Red GIF
func SendRedGif(bot *tgbotapi.BotAPI, message *tgbotapi.Message, assetPaths *config.AssetPaths) {
	// Формируем полный путь к GIF-файлу
	gifPath := filepath.Join(assetPaths.Gifs, "red_answer.gif")

	// Проверяем, существует ли файл
	if !fileExists(gifPath) {
		log.Printf("GIF не найден: %s", gifPath)
		msg := tgbotapi.NewMessage(message.Chat.ID, "У тебя нет доступа!") //❌ GIF не найден
		bot.Send(msg)
		return
	}

	// Отправляем GIF как документ (Telegram API лучше обрабатывает GIF таким образом)
	document := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FilePath(gifPath))
	document.Caption = "У тебя нет доступа!"

	_, err := bot.Send(document)
	if err != nil {
		log.Printf("Ошибка отправки GIF: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка отправки GIF")
		bot.Send(msg)
	} else {
		log.Printf("GIF отправлен: %s", gifPath)
	}
}
