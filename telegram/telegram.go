package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type TelegramClient struct {
	client          *http.Client
	token           string
	close_event     chan bool
	update_listener []func(Update) error
	uri             string
}

type UpdateRequest struct {
	Offset  int32 `json:"offset"`
	Timeout int   `json:"timeout"`
}

type MessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type MessageResponse struct {
	OK bool `json:"ok"`
}

type UpdateResponse struct {
	OK      bool     `json:"ok"`
	Updates []Update `json:"result"`
}

type Update struct {
	UpdateID int32   `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int32  `json:"message_id"`
	From      User   `json:"from"`
	Chat      User   `json:"chat"`
	Date      int32  `json:"date"`
	Text      string `json:"text"`
}

type User struct {
	ID        int32  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"username"`
	Type      string `json:"type"`
}

func NewClient(token string) *TelegramClient {
	endpoint := "https://api.telegram.org"
	bot_uri := fmt.Sprintf("%s/bot%s", endpoint, token)

	return &TelegramClient{
		token:       token,
		close_event: make(chan bool),
		uri:         bot_uri,
	}
}

func (tc *TelegramClient) Start(ctx context.Context) error {
	tc.client = &http.Client{
		Timeout: time.Second * 30,
	}

	// Start long polling loop
	// err_count := 1

	last_id := int32(0)
	// last := 0

	go func() {
	outof:
		for {
		retry:
			ur := UpdateRequest{
				Offset:  last_id + 1,
				Timeout: 25,
			}

			jBytes, bErr := json.Marshal(ur)

			if bErr != nil {
				return
			}

			r := bytes.NewReader(jBytes)

			req, err := http.NewRequest("POST", tc.uri+"/getUpdates", r)

			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "FBI")
			req.Header.Add("Content-Type", "application/json")

			if err != nil {
				log.Println(err)
				break
			}

			resp, err := tc.client.Do(req)

			if err != nil {
				/* if err == http.ErrHandlerTimeout {
					continue
				}

				log.Println(err)
				err_count++
				if err_count < 5 {
					time.Sleep(time.Second * 5)
					goto retry
				}
				break */
				time.Sleep(time.Second * 5)
				goto retry
			}

			bytes, err := io.ReadAll(resp.Body)

			if err != nil {
				break
			}

			// log.Println(string(bytes))

			var res UpdateResponse

			if err := json.Unmarshal(bytes, &res); err != nil {
				log.Println(err)
			}

			// log.Println(res)

		sendloop:
			for _, u := range res.Updates {
				for f := 0; f < len(tc.update_listener); f++ {
					if err := tc.update_listener[f](u); err != nil {
						log.Print(err)
						break sendloop
					}
					last_id = u.UpdateID
				}
			}

			log.Println(last_id)
			select {
			case <-tc.close_event:
				break outof
			case <-ctx.Done():
				break outof
			default:
			}
			// time.Sleep(time.Second * 5)
		}
		close(tc.close_event)
	}()
	select {
	case <-tc.close_event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (tc *TelegramClient) IncomingUpdateListener(f func(Update) error) {
	tc.update_listener = append(tc.update_listener, f)
}

func (tc *TelegramClient) SendMessage(chat_id string, message string) error {
	m := MessageRequest{
		ChatID: chat_id,
		Text:   message,
	}

	jBytes, bErr := json.Marshal(m)

	if bErr != nil {
		return bErr
	}

	r := bytes.NewReader(jBytes)

	req, err := http.NewRequest("POST", tc.uri+"/sendMessage", r)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "FBI")
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return err
	}

	resp, err := tc.client.Do(req)

	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var res MessageResponse

	if err := json.Unmarshal(bytes, &res); err != nil {
		log.Println(err)
	}

	if !res.OK {
		return fmt.Errorf("message to %s not sent", message)
	}

	return nil
}
