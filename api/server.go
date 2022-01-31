package api

import (
	"context"
	"net/http"
	"stargazer/SLIIT-Notifications/bot"
	"stargazer/SLIIT-Notifications/keyreader"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Instance struct {
	db     *mongo.Database
	ctx    context.Context
	server *http.Server
}

type ErrorResponse struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

type UsersResponse struct {
	Error bool            `json:"error"`
	Users []bot.SLIITUser `json:"users"`
}

type SitesResponse struct {
	Error bool            `json:"error"`
	Sites []bot.SLIITSite `json:"sites"`
}

func NewInstance(db *mongo.Database, ctx context.Context) *Instance {
	return &Instance{
		db:  db,
		ctx: ctx,
	}
}

func (i *Instance) Start() error {
	s := gin.Default()
	db := i.db
	ctx := i.ctx

	s.GET("/api/users", func(c *gin.Context) {
		var users []bot.SLIITUser
		k := keyreader.NewReader(bot.SLIITUser{}, "bson")

		opts := options.Find()
		opts.SetProjection(bson.M{
			k.Get("ID"):       1,
			k.Get("Username"): 1,
		})

		cur, curErr := db.Collection("user").Find(ctx, bson.M{}, opts)

		if curErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(curErr))
			return
		}

		if err := cur.All(ctx, &users); err != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(err))
			return
		}

		c.JSON(200, newUsersResponse(users))
	})

	s.GET("/api/sites", func(c *gin.Context) {
		var sites []bot.SLIITSite
		cur, curErr := db.Collection("sites").Find(ctx, bson.M{})

		if curErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(curErr))
			return
		}

		if err := cur.All(ctx, &sites); err != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(err))
			return
		}

		c.JSON(200, newSitesResponse(sites))
	})

	s.StaticFS("/dashboard", gin.Dir("./dashboard", false))

	i.server = &http.Server{
		Addr:    ":8080",
		Handler: s,
	}

	return i.server.ListenAndServe()
}

func (i *Instance) Stop(ctx context.Context) error {
	return i.server.Shutdown(ctx)
}

func newErrorResponse(err error) ErrorResponse {
	return ErrorResponse{
		Error:   true,
		Message: err.Error(),
	}
}

func newUsersResponse(users []bot.SLIITUser) UsersResponse {
	return UsersResponse{
		Error: false,
		Users: users,
	}
}

func newSitesResponse(sites []bot.SLIITSite) SitesResponse {
	return SitesResponse{
		Error: false,
		Sites: sites,
	}
}
