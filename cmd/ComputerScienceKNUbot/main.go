package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	handlers "github.com/yanferens/StarHelperBot/internal/handler"
	repository "github.com/yanferens/StarHelperBot/internal/repository"
	service "github.com/yanferens/StarHelperBot/internal/service"
)

func main() {
	dsn := os.Getenv("dsn1")
	TOKEN := os.Getenv("TOKEN1")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Помилка підключення до БД:", err)
	}
	defer db.Close()
	repo := repository.NewRepository(db)
	menuService := service.NewMenuService(repo)
	URL := "https://pumped-tahr-amazed.ngrok-free.app/bot"
	ctx := context.Background()
	token := TOKEN
	if token == "" {
		log.Fatal("TOKEN env var required")
	}

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	err = bot.DeleteWebhook(ctx, &telego.DeleteWebhookParams{
		DropPendingUpdates: true,
	})
	if err != nil {
		log.Fatalf("delete webhook error: %v", err)
	}
	// Set webhook
	err = bot.SetWebhook(ctx, &telego.SetWebhookParams{
		URL:         URL,
		SecretToken: bot.SecretToken(),
	})
	if err != nil {
		log.Fatalf("set webhook error: %v", err)
	}

	mux := http.NewServeMux()

	updates, err := bot.UpdatesViaWebhook(ctx, telego.WebhookHTTPServeMux(mux, "/bot", bot.SecretToken()))
	if err != nil {
		log.Fatalf("updates webhook error: %v", err)
	}

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("http listen error: %v", err)
		}
	}()

	// Create bot handler and specify from where to get updates
	bh, _ := th.NewBotHandler(bot, updates)
	handlers.InitMenuHandlers(menuService, bh, bot)
	handlers.InitStarExchange(bh, bot)

	// Stop handling updates
	defer func() { _ = bh.Stop() }()
	//InitHandlers(bh, bot)

	_ = bh.Start()
}
