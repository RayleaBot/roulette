package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	rayleabot "github.com/RayleaBot/RayleaBot/sdk/go"
	"github.com/RayleaBot/roulette/internal/plugin"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := rayleabot.Run(ctx, rayleabot.Options{}, plugin.New()); err != nil {
		log.Fatal(err)
	}
}
