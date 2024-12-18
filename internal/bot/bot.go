package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBot struct {
    bot tgbotapi.BotAPI
}

func NewTgBot(token string) *TgBot {
    bot, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        panic(err)
    }
    return &TgBot{bot: *bot}
}

func (tg *TgBot) Start() {
    u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, "+update.Message.From.FirstName+"!")
			tg.bot.Send(reply)
		}
	}
}

