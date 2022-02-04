package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"stargazer/SLIIT-Notifications/bot"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Instance struct {
	db          *mongo.Database
	server      *http.Server
	restart_bot func()
}

type StateResponse struct {
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

func NewInstance(db *mongo.Database) *Instance {
	return &Instance{
		db: db,
	}
}

func (i *Instance) Start(ctx context.Context) error {
	s := gin.Default()
	db := i.db
	// ctx := i.ctx

	s.GET("/api/users", func(c *gin.Context) {
		var users []bot.SLIITUser

		opts := options.Find()
		opts.SetProjection(bson.M{
			bot.SLIITUserK.Get("ID"):       1,
			bot.SLIITUserK.Get("Username"): 1,
		})

		cur, curErr := db.Collection("user").Find(c, bson.M{}, opts)

		if curErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(curErr))
			return
		}

		if err := cur.All(c, &users); err != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(err))
			return
		}

		c.JSON(200, newUsersResponse(users))
	})

	s.GET("/api/users/:id/sites", func(c *gin.Context) {
		var sites []bot.SLIITSite

		id := c.Param("id")

		obj, objErr := primitive.ObjectIDFromHex(id)

		if objErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(objErr))
			return
		}

		filter := bson.M{
			bot.SLIITSiteK.Get("UserID"): obj,
		}

		cur, curErr := db.Collection("sites").Find(c, filter)

		if curErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(curErr))
			return
		}

		if err := cur.All(c, &sites); err != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(err))
			return
		}

		c.JSON(200, newSitesResponse(sites))
	})

	s.GET("/api/sites", func(c *gin.Context) {
		var sites []bot.SLIITSite
		cur, curErr := db.Collection("sites").Find(c, bson.M{})

		if curErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(curErr))
			return
		}

		if err := cur.All(c, &sites); err != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(err))
			return
		}

		c.JSON(200, newSitesResponse(sites))
	})

	s.POST("/api/sites/:id/disable", func(c *gin.Context) {
		id := c.Param("id")

		update := bson.M{
			"$set": bson.M{
				bot.SLIITSiteK.Get("Disabled"): true,
			},
		}

		log.Println(update)

		obj, objErr := primitive.ObjectIDFromHex(id)

		if objErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(objErr))
			return
		}

		filter := bson.M{
			bot.SLIITSiteK.Get("ID"): obj,
		}

		res, uErr := db.Collection("sites").UpdateOne(c, filter, update)

		if res.MatchedCount == 0 {
			c.AbortWithStatusJSON(400, newErrorResponse(fmt.Errorf("no such site with id %s", id)))
			return
		}

		if uErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(uErr))
			return
		}

		c.JSON(200, newSuccessResponse("success"))
	})

	s.POST("/api/sites/:id/enable", func(c *gin.Context) {
		id := c.Param("id")

		update := bson.M{
			"$set": bson.M{
				bot.SLIITSiteK.Get("Disabled"): false,
			},
		}
		log.Println(update)

		obj, objErr := primitive.ObjectIDFromHex(id)

		if objErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(objErr))
			return
		}

		filter := bson.M{
			bot.SLIITSiteK.Get("ID"): obj,
		}

		res, uErr := db.Collection("sites").UpdateOne(c, filter, update)

		if res.MatchedCount == 0 {
			c.AbortWithStatusJSON(400, newErrorResponse(fmt.Errorf("no such site with id %s", id)))
			return
		}

		if uErr != nil {
			c.AbortWithStatusJSON(400, newErrorResponse(uErr))
			return
		}

		c.JSON(200, newSuccessResponse("success"))
	})

	s.POST("/api/bot/restart", func(c *gin.Context) {
		if i.restart_bot == nil {
			c.AbortWithStatusJSON(400, newErrorResponse(fmt.Errorf("bot restarter not registered")))
			return
		}

		i.restart_bot()

		c.JSON(200, newSuccessResponse("success"))
	})

	s.StaticFS("/ui", gin.Dir("./ui", false))

	i.server = &http.Server{
		Addr:    ":8080",
		Handler: s,
	}

	hErr := make(chan error)

	go func() {
		hErr <- i.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-hErr:
		return e
	}
}

func (i *Instance) RegisterBotRestarter(restart_func func()) {
	i.restart_bot = restart_func
}

func (i *Instance) Stop(ctx context.Context) error {
	return i.server.Shutdown(ctx)
}

func newErrorResponse(err error) StateResponse {
	return StateResponse{
		Error:   true,
		Message: err.Error(),
	}
}

func newSuccessResponse(msg string) StateResponse {
	return StateResponse{
		Error:   false,
		Message: msg,
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
