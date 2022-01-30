package telegram

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type TelegramClient struct {
	client *http.Client
	ctx    context.Context
	db     *mongo.Database
}

func NewClient(ctx context.Context, db *mongo.Database) *TelegramClient {
	return &TelegramClient{
		db:  db,
		ctx: ctx,
	}
}

func (tc *TelegramClient) Start() {
	tc.client = &http.Client{}

}
