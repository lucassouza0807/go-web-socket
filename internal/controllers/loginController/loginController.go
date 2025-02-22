package logincontroller

import (
	"go-web-socket/config"
	"go-web-socket/internal/models"
	jwtService "go-web-socket/internal/services/JWTService"
	useHash "go-web-socket/internal/utils/hash"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Credentials struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

func Login(ctx *gin.Context) {
	var user models.User //Model to scan query results

	db, err := config.GetDatabaseConnection()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	var credentials Credentials

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Nunhum campo foi passo no corpo da requisição",
		})

		return
	}

	result := db.Where("username", credentials.Username).First(&user)

	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "O usúario fornecido não existe.",
		})

		return
	}

	passwordMatches := useHash.CheckPasswordHash(*credentials.Password, user.Password)

	if !passwordMatches {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Credenciais inválidas",
		})

		return
	}

	token, err := jwtService.CreateToken(&jwtService.UserToken{
		UserId:   user.UserId,
		Name:     user.Name,
		Username: user.Username,
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Houve um erro tentar fazer login",
		})

		return
	}

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	user.Password = ""

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"exp":   int64((time.Minute * 60 * 48).Seconds()),
		"user":  user,
	})

}
