package api

import (
	"fmt"
	"log"
	"net/http"
	"stargazer/SLIIT-Notifications/bot"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PrivateRouter(router *gin.RouterGroup) {
	router.Use(AuthRequired)

	router.GET("users", GetUsers)
	router.GET("/users/:id/sites", GetUserSites)
	router.GET("/sites", GetAllSites)
	router.POST("/sites/:id/disable", DisableSite)
	router.POST("/sites/:id/enable", EnableSite)
	router.POST("/bot/restart", RestartBot)
	router.GET("/history/:id", GetHistory)
}

func AuthRequired(c *gin.Context) {
	s := sessions.Default(c)
	auth := s.Get("auth")

	if auth == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, newErrorResponse(fmt.Errorf("unauthorized. try harder.")))
		return
	}

	c.Next()
}

func GetUsers(c *gin.Context) {
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
}

func GetUserSites(c *gin.Context) {
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
}

func GetAllSites(c *gin.Context) {
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
}

func DisableSite(c *gin.Context) {
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
}

func EnableSite(c *gin.Context) {
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
}

func RestartBot(c *gin.Context) {
	if restart_listener == nil {
		c.AbortWithStatusJSON(400, newErrorResponse(fmt.Errorf("bot restarter not registered")))
		return
	}

	restart_listener()

	c.JSON(200, newSuccessResponse("success"))
}

func GetHistory(c *gin.Context) {
	id := c.Param("id")

	obj, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.AbortWithStatusJSON(400, newErrorResponse(err))
		return
	}

	filter := bson.M{
		bot.SLIITHistoryK.Get("ID"): obj,
	}

	res := db.Collection("history").FindOne(c, filter)

	var history bot.SLIITHistory

	if err := res.Decode(&history); err != nil {
		c.AbortWithStatusJSON(400, newErrorResponse(err))
		return
	}

	c.JSON(200, StateResponse{
		Error:   false,
		Message: "success",
		Data:    history,
	})

}
