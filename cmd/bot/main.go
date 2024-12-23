package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aidosgal/lenshub/internal/bot"
	"github.com/aidosgal/lenshub/internal/config"
	"github.com/aidosgal/lenshub/internal/repository"
	"github.com/aidosgal/lenshub/internal/service"
    _ "github.com/lib/pq"
)


func main() {
    cfg := config.MustLoad()

	postgresURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
    db, err := sql.Open("postgres", postgresURL)
	if err != nil {
        panic(err)
	}
	defer db.Close()
    
    userRepository := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepository)

    orderService := service.NewOrderService(db)

    responseService := service.NewResponseService(db)
    bot := bot.NewTgBot(cfg.Telegram, userService, orderService, responseService)

    log.Print("bot starting...")
    bot.Start()
}
