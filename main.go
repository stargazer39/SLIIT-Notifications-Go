package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"stargazer/SLIIT-Notifications/api"
	"stargazer/SLIIT-Notifications/bot"
	"stargazer/SLIIT-Notifications/helpers"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Panicln(err)
	}

	/* Test code start */
	/* t := telegram.NewClient(os.Getenv("TELEGRAM_BOT_TOKEN"))

	t.IncomingUpdateListener(func(u telegram.Update) error {
		log.Println(u.Message.Text)
		e := t.SendMessage(fmt.Sprint(u.Message.From.ID), "Hello "+u.Message.Text)
		log.Println(e)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if err := t.Start(ctx); err != nil {
		log.Panic(err)
	}

	return */
	/* Test code end */

	// Handle exit
	proc_exit_event := make(chan os.Signal, 1)
	signal.Notify(proc_exit_event, os.Interrupt)

	// Create Main context
	main_ctx, cancel_main := context.WithCancel(context.Background())

	defer cancel_main()

	// Create cache folders
	if err := helpers.CreateFolders(); err != nil {
		log.Panic(err)
	}

	mongo_uri := os.Getenv("MONGO_URI")

	client, cErr := mongo.NewClient(options.Client().ApplyURI(mongo_uri))

	if cErr != nil {
		log.Panicln(cErr)
	}

	// Set timeout context for connecting
	ctx, cancel_ctx := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel_ctx()

	conErr := client.Connect(ctx)

	if conErr != nil {
		log.Panicln(conErr)
	}

	defer client.Disconnect(main_ctx)

	db := client.Database(os.Getenv("DATABASE"))

	interval, iErr := strconv.Atoi(os.Getenv("INTERVAL"))

	if iErr != nil {
		interval = 5
	}

	// Start API
	a := api.NewInstance(db)
	go func() {
		if err := a.Start(main_ctx); err != nil {
			log.Println(err)
		}
	}()

	restart_event := make(chan bool)
	stop_event := make(chan bool)
	wait_to_stop := make(chan bool)

	go func() {
		for {
			// Make bot's own context
			bot_context, cancel_bot_context := context.WithCancel(main_ctx)

			defer cancel_bot_context()

			sliit_bot := bot.NewBot(db, time.Second*time.Duration(interval))

			// Changed log
			file, fErr := os.Create("./.cache/log.txt")

			if fErr != nil {
				log.Panic(fErr)
			}

			defer file.Close()

			sliit_bot.RegisterChangeListener(func(h *bot.SLIITHistory) {
				log.Println(h)
				file.Write([]byte(fmt.Sprintf("%s changed \n", h.SiteID)))
			})

			// Start sliit bot
			go func() {
				if err := sliit_bot.Start(bot_context); err != nil {
					log.Println(err)
				}
			}()

			stop := false

			select {
			case <-restart_event:
			case <-bot_context.Done():
				log.Print(bot_context.Err())
				return
			case <-stop_event:
				stop = true
			}

			timeout_ctx, cancel := context.WithTimeout(bot_context, time.Second*10)

			defer cancel()

			// Stop the Bot
			log.Println("Stopping Bot")
			if err := sliit_bot.Stop(timeout_ctx); err != nil {
				log.Print(err)
			}

			if stop {
				break
			}
		}
		wait_to_stop <- true
	}()

	// Accept command from terminal
	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			text, _ := reader.ReadString('\n')

			if text == "rs" {
				log.Println("Restarting the bot")
				restart_event <- true
			}
		}
	}()

	// Register Restarter
	a.RegisterBotRestarter(func() {
		restart_event <- true
	})

	<-proc_exit_event

	// Trigger stop and wait to stop
	close(stop_event)
	<-wait_to_stop

	timeout_ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	// Stop the API
	log.Println("Stopping the API")
	if err := a.Stop(timeout_ctx); err != nil {
		log.Println(err)
	}

	<-timeout_ctx.Done()
}
