package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func PublicRouter(router *gin.RouterGroup) {
	router.POST("/session/new", NewSessionHandler)
}

func NewSessionHandler(c *gin.Context) {
	var req SessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, newErrorResponse(err))
		return
	}

	pass := os.Getenv("ADMIN_PASSWORD")
	log.Println(pass, req.Password)

	if strings.Compare(pass, req.Password) != 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, newErrorResponse(fmt.Errorf("wrong password")))
		return
	}

	sess := sessions.Default(c)

	sess.Set("auth", "admin")

	if err := sess.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, newErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, newSuccessResponse("success"))

}
