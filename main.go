package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
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

type TelegramUnsent struct {
	ChatID   int32     `bson:"chat_id,omitempty" json:"chat_id"`
	SendTime time.Time `bson:"send_time,omitempty" json:"send_time"`
	Message  string    `bson:"message,omitempty" json:"message"`
	Error    string    `bson:"error,omitempty" json:"error"`
}

var (
	TelegramUserK = keyreader.NewReader(TelegramUser{}, "bson")
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Panicln(err)
	}

	api_only := flag.Bool("api-only", false, "Serve the api only")

	flag.Parse()
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
	go func() {
		if err := api.Start(db, main_ctx); err != nil {
			log.Println(err)
		}
	}()

	if *api_only {
		<-proc_exit_event
		ctx, cancel := context.WithTimeout(main_ctx, time.Second*10)

		defer cancel()
		if err := api.Stop(ctx); err != nil {
			log.Println(err)
		}

		return
	}
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
			file, fErr := os.OpenFile("./.cache/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if fErr != nil {
				log.Panic(fErr)
			}

			defer file.Close()

			sliit_bot.RegisterChangeListener(func(h bot.SLIITHistory, uid primitive.ObjectID) {
				log.Println(h)
				unsent, err := func() ([]TelegramUnsent, error) {
					var site bot.SLIITSite

					res := db.Collection("sites").FindOne(context.TODO(), bson.M{
						bot.SLIITSiteK.Get("ID"): h.SiteID,
					})

					if err := res.Decode(&site); err != nil {
						return nil, err
					}

					// Handle notifications

					filter := bson.M{
						TelegramUserK.Get("SubscribedDeg"): uid,
					}

					cur, curErr := db.Collection("telegram_groups").Find(context.TODO(), filter)

					if curErr != nil {
						return nil, curErr
					}

					var subbed []TelegramUser

					if err := cur.All(context.TODO(), &subbed); err != nil {
						return nil, err
					}

					var t_unsent []TelegramUnsent

					for _, sub := range subbed {

						url := fmt.Sprintf(os.Getenv("URL_TEMPLATE"), h.ID.Hex())

						msg := fmt.Sprintf("Module\n<a href='%s'>%s</a> changed.\nVisit <a href='%s'>here</a> to see what changed",
							site.URL,
							site.Name,
							url,
						)

						if err := tc.SendHTML(fmt.Sprint(sub.ChatID), msg); err != nil {
							t_unsent = append(t_unsent, TelegramUnsent{ChatID: sub.ChatID, SendTime: time.Now(), Message: msg, Error: err.Error()})
						}
					}

					if len(t_unsent) > 0 {
						return t_unsent, fmt.Errorf("unsent message")
					}

					return nil, nil
				}()

				if err != nil {
					log.Println(err)
					if unsent != nil {
						jBytes, jerr := json.Marshal(unsent)

						if jerr != nil {
							log.Panicln(jerr)
						}

						if _, err := file.Write([]byte("-------------")); err != nil {
							log.Panicln(err)
						}

						if _, err := file.Write(jBytes); err != nil {
							log.Panicln(err)
						}

						if _, err := file.Write([]byte("-------------")); err != nil {
							log.Panicln(err)
						}

						generic_unsent := make([]interface{}, 0)

						for _, f := range unsent {
							generic_unsent = append(generic_unsent, f)
						}

						if _, err := db.Collection("telegram_unsent").InsertMany(context.TODO(), generic_unsent); err != nil {
							log.Panicln(err)
						}
					}
				}
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
	api.RegisterBotRestarter(func() {
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
	if err := api.Stop(timeout_ctx); err != nil {
		log.Println(err)
	}

	<-timeout_ctx.Done()
}
