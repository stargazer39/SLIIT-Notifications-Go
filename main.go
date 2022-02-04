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
	"stargazer/SLIIT-Notifications/keyreader"
	"stargazer/SLIIT-Notifications/telegram"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Subscribed Module and the doner user id is the same
type TelegramUser struct {
	ChatID        int32              `bson:"chat_id,omitempty" json:"chat_id"`
	AddedTime     time.Time          `bson:"added_time,omitempty" json:"added_time"`
	SubscribedDeg primitive.ObjectID `bson:"deg_uid,omitempty" json:"site_id"`
}

var (
	TelegramUserK = keyreader.NewReader(TelegramUser{}, "bson")
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

	// Start Telegram client
	// New telegram client
	tc := telegram.NewClient(os.Getenv("TELEGRAM_BOT_TOKEN"))

	// Manage Telegram client

	tc.IncomingUpdateListener(func(u telegram.Update) error {
		if strings.HasPrefix(u.Message.Text, "/") {
			command := strings.Split(u.Message.Text, " ")
			exist := false

			filter := bson.M{
				TelegramUserK.Get("ChatID"): u.Message.From.ID,
			}

			n, err := db.Collection("telegram_admin").CountDocuments(context.TODO(), filter)

			if err != nil {
				return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Database error")
			}

			if n > 0 {
				exist = true
			}

			switch command[0] {
			case "/admin":
				if exist {
					return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Alredy exist")
				}

				if len(command) < 2 {
					return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Where's the password?")
				}

				if strings.Compare(command[1], os.Getenv("TELEGRAM_PASSWORD")) != 0 {
					return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Wrong password?")
				}

				// Add to Database
				t_user := TelegramUser{
					ChatID:    u.Message.From.ID,
					AddedTime: time.Now(),
				}

				if _, err := db.Collection("telegram_admin").InsertOne(context.TODO(), &t_user); err != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Database error")
				}

				return tc.SendMessage(fmt.Sprint(u.Message.From.ID), "Welcome")

			case "/list":
				if !exist {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "You are not an admin")
				}

				cur, curErr := db.Collection("user").Find(context.TODO(), bson.M{})

				if curErr != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Database error")
				}

				var users []bot.SLIITUser
				if err := cur.All(context.TODO(), &users); err != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Database error")
				}

				message := ""

				for _, usr := range users {
					message += usr.DegreeID + "\n"
				}

				return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), message)

			case "/addme":
				log.Println(u)
				if !exist {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "You are not an admin")
				}

				if len(command) < 2 {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Specify the degree you want to subscribe to")
				}

				filter = bson.M{
					bot.SLIITUserK.Get("DegreeID"): command[1],
				}
				// Get semester with the specific id
				res_usr := db.Collection("user").FindOne(context.TODO(), filter)

				if err != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Database error")
				}

				var user bot.SLIITUser

				if err := res_usr.Decode(&user); err != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "No such degree")
				}

				filter := bson.M{
					TelegramUserK.Get("SubscribedDeg"): user.ID,
				}
				// tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Hi")
				n_count, err := db.Collection("telegram_groups").CountDocuments(context.TODO(), filter)

				if err != nil {
					log.Panic(err)
				}

				if n_count > 0 {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Already added")
				}

				t_user := TelegramUser{
					ChatID:        u.Message.Chat.ID,
					AddedTime:     time.Now(),
					SubscribedDeg: user.ID,
				}

				if _, err := db.Collection("telegram_groups").InsertOne(context.TODO(), t_user); err != nil {
					return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Database error")
				}

				return tc.SendMessage(fmt.Sprint(u.Message.Chat.ID), "Success")
			}
		}

		return nil
	})

	go func() {
		if err := tc.Start(main_ctx); err != nil {
			log.Println(err)
			cancel_main()
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

			sliit_bot.RegisterChangeListener(func(h bot.SLIITHistory, uid primitive.ObjectID) {
				log.Println(h)
				site := bot.SLIITSite{ID: h.SiteID}

				res := db.Collection("sites").FindOne(context.TODO(), site)

				if err := res.Decode(&site); err != nil {
					log.Panic(err)
				}

				// Handle notifications

				filter := bson.M{
					TelegramUserK.Get("SubscribedDeg"): uid,
				}

				cur, curErr := db.Collection("telegram_groups").Find(context.TODO(), filter)

				if curErr != nil {
					log.Panicln(curErr)
				}

				var subbed []TelegramUser

				if err := cur.All(context.TODO(), &subbed); err != nil {
					log.Panicln(err)
				}

				for _, sub := range subbed {
					if err := tc.SendMessage(fmt.Sprint(sub.ChatID), h.SiteID.String()+" changed "+site.Name); err != nil {
						log.Panic(err)
					}
				}

				// tc.SendMessage()
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
