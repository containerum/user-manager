package routes

import (
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"github.com/gin-gonic/gin"
)

type Services struct {
	MailClient *clients.MailClient
	DB         *models.DB
	AuthClient auth.AuthClient
}

var svc Services

type Error struct {
	Error string `json:"error"`
}

func SetupRoutes(app *gin.Engine, services Services) {
	svc = services
	user := app.Group("/user")
	{
		user.POST("/create", userCreateHandler)
		user.POST("/sign_up/resend", linkResendHandler)
	}
}
