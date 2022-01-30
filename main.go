package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"stargazer/SLIIT-Notifications/bot"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Panicln(err)
	}

	// Handle exit
	exit_event := make(chan os.Signal, 1)
	signal.Notify(exit_event, os.Interrupt, syscall.SIGTERM)

	mongo_uri := os.Getenv("MONGO_URI")

	client, cErr := mongo.NewClient(options.Client().ApplyURI(mongo_uri))
	if cErr != nil {
		log.Panicln(cErr)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	conErr := client.Connect(ctx)

	if conErr != nil {
		log.Panicln(conErr)
	}

	defer client.Disconnect(ctx)
	databases, err := client.ListDatabaseNames(ctx, bson.M{})

	if err != nil {
		log.Panicln(err)
	}

	log.Println(databases)

	db := client.Database("SLIITNotificationsGo2")
	bot_context, cancel := context.WithCancel(context.Background())

	defer cancel()

	bot := bot.NewBot(bot_context, db)

	go bot.Start()

	<-exit_event
	log.Printf("Exiting boii")
	bot.Stop()
}
