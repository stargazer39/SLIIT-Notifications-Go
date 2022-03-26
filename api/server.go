package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"stargazer/SLIIT-Notifications/bot"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Instance struct {
	server      *http.Server
	restart_bot func()
}

type StateResponse struct {
	Message string      `json:"message"`
	Error   bool        `json:"error"`
	Data    interface{} `json:"data,omitempty"`
}

type UsersResponse struct {
	Error bool            `json:"error"`
	Users []bot.SLIITUser `json:"users"`
}

type SitesResponse struct {
	Error bool            `json:"error"`
	Sites []bot.SLIITSite `json:"sites"`
}

var db *mongo.Database
var server *http.Server
var restart_listener func()

func Start(mdb *mongo.Database, ctx context.Context) error {
	router := gin.Default()
	db = mdb

	router.Use(cors.Default())
	router.Use(sessions.Sessions("user", sessions.NewCookieStore(generateRandom())))

	public := router.Group("/api/public")
	private := router.Group("/api/private")

	PublicRouter(public)
	PrivateRouter(private)

	router.NoRoute(func(c *gin.Context) {
		c.File("./ui/build/index.html")
	})

	router.Use(static.Serve("/", static.LocalFile("./ui/build/", true)))

	server = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	hErr := make(chan error)

	go func() {
		hErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-hErr:
		return e
	}
}

func RegisterBotRestarter(restart_func func()) {
	restart_listener = restart_func
}

func Stop(ctx context.Context) error {
	return server.Shutdown(ctx)
}

func Obfuscate(i string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(i)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", h.Sum(nil)), nil
}

func generateRandom() []byte {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		log.Panic(err)
	}

	return buf
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
