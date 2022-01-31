package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"stargazer/SLIIT-Notifications/api"
	"stargazer/SLIIT-Notifications/bot"
	"stargazer/SLIIT-Notifications/helpers"
	"strconv"
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

	// Create cache folders
	if err := helpers.CreateFolders(); err != nil {
		log.Panic(err)
	}

	mongo_uri := os.Getenv("MONGO_URI")

	client, cErr := mongo.NewClient(options.Client().ApplyURI(mongo_uri))
	if cErr != nil {
		log.Panicln(cErr)
	}

	ctx, cancel_ctx := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel_ctx()

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

	db := client.Database(os.Getenv("DATABASE"))

	interval, iErr := strconv.Atoi(os.Getenv("INTERVAL"))

	if iErr != nil {
		interval = 5
	}

	bot_context, cancel := context.WithCancel(context.Background())

	defer cancel()

	// Start API
	a := api.NewInstance(db, bot_context)
	go a.Start()

	sliit_bot := bot.NewBot(bot_context, db, time.Second*time.Duration(interval))

	// Changed
	file, fErr := os.Create("./.cache/log.txt")

	if fErr != nil {
		log.Panic(fErr)
	}

	defer file.Close()

	sliit_bot.RegisterChangeListener(func(h *bot.SLIITHistory) {
		log.Println(h)
		file.Write([]byte(fmt.Sprintf("%s changed \n", h.SiteID)))
	})

	go sliit_bot.Start()

	<-exit_event
	log.Printf("Exiting boii")
	sliit_bot.Stop()
	a.Stop(bot_context)
}
