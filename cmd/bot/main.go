package main

import (
	"log"

	"github.com/aidosgal/lenshub/internal/bot"
	"github.com/aidosgal/lenshub/internal/config"
)


func main() {
    cfg := config.MustLoad()

    bot := bot.NewTgBot(cfg.Telegram)

    log.Print("bot starting...")
    bot.Start()
}
