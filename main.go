package main

import (
	logincontroller "go-web-socket/internal/controllers/loginController"
	"go-web-socket/internal/controllers/userController"
	"go-web-socket/internal/socket"
	"go-web-socket/internal/utils/logger"
	"go-web-socket/internal/utils/migration"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	logger.InitLogger()
	migration.RunMigration()

	app := gin.Default()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://192.168.0.124:3000"}, // Ajuste conforme seu frontend
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	app.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	app.POST("/login", logincontroller.Login)
	app.GET("/user/:username", userController.GetUser)
	app.POST("/create-user", userController.CreateUser)
	app.POST("/upload-user-avatar/:user_id", userController.UploadUserAvatar)
	app.POST("/change-user-avatar:user_id", userController.UploadUserAvatar)
	app.GET("/users", userController.GetUsers)
	app.PUT("/edit-user/:user_id", userController.EditUser)
	//socket
	app.GET("/ws/user/:user_id", socket.HandleSocket)
	app.GET("/ws/online-users", socket.GetOnlineUsers)
	app.POST("/ws/send-private-message", socket.SendMessage)

	app.Run()
}
