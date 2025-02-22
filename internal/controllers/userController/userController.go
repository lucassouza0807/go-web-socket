package userController

import (
	"encoding/base64"
	"go-web-socket/config"
	"go-web-socket/internal/models"
	s3uploadservice "go-web-socket/internal/services/S3UploadService"
	userService "go-web-socket/internal/services/UserService"
	useHash "go-web-socket/internal/utils/hash"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func EditUser(ctx *gin.Context) {
	var requestBody models.User

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Nenhum campo foi passado na requisição",
		})
		return
	}

	user, err := userService.EditUser(ctx.Param("user_id"), requestBody)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"details": err.Error(),
		})

		return
	}

	user.Password = ""

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usúario editado com sucesso",
		"user":    user,
	})

}

func GetUser(ctx *gin.Context) {
	var user models.User
	userName := ctx.Param("username")

	db, err := config.GetDatabaseConnection()

	if err != nil {
		log.Printf("Error while connecting to database: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	result := db.Where("username", userName).Find(&user)

	if err := result.Error; err != nil {
		log.Fatalf("Error while make query: %v", err.Error())
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Erro while making query",
		})

		return
	}

	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"mesage": "Usuário não econtrado",
		})

		return
	}

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func GetUsers(ctx *gin.Context) {
	var users []models.User

	db, err := config.GetDatabaseConnection()

	if err != nil {
		log.Printf("Error while connecting to database: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	db.Find(&users)

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
	})

}

func UpdateUserAvatar(ctx *gin.Context) {
	if ctx.Query("filename") == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Filename was not provided",
		})

		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Erro ao obter o arquivo",
			"error":   err.Error(),
		})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Erro ao abrir o arquivo",
			"error":   err.Error(),
		})
		return
	}
	defer fileContent.Close()

	fileBytes, err := ioutil.ReadAll(fileContent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Erro ao ler o arquivo",
			"error":   err.Error(),
		})
		return
	}

	base64Encoded := base64.StdEncoding.EncodeToString(fileBytes)

	fileURl, err := s3uploadservice.ReplaceFile(base64Encoded, ctx.Query("filename"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Foto de perfil atualizada com sucesso!",
		"file":    fileURl,
	})
}

func UploadUserAvatar(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Erro ao obter o arquivo",
			"error":   err.Error(),
		})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Erro ao abrir o arquivo",
			"error":   err.Error(),
		})
		return
	}
	defer fileContent.Close()

	fileBytes, err := ioutil.ReadAll(fileContent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Erro ao ler o arquivo",
			"error":   err.Error(),
		})
		return
	}

	base64Encoded := base64.StdEncoding.EncodeToString(fileBytes)

	fileURl, err := s3uploadservice.Upload(base64Encoded)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	user, err := userService.EditUser(ctx.Param("user_id"), models.User{
		Avatar: fileURl,
	})

	if err != nil {
		ctx.JSON(http.StatusNonAuthoritativeInfo, gin.H{
			"message": err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Foto de perfil atualizada com sucesso!",
		"user":    user,
	})
}

func CreateUser(ctx *gin.Context) {
	db, err := config.GetDatabaseConnection()

	if err != nil {
		log.Printf("Error while connecting to database: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	var requestBody models.User

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		log.Printf("Error while binding JSON: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "No fields were passed in the request body",
		})

		return
	}

	userId := uuid.New()

	hashed_password, err := useHash.HashPassword(requestBody.Password)

	if err != nil {
		log.Fatalf("Error while hashing PASSWORD: %v", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Sorry but we got a problem",
		})

		return

	}

	user := models.User{
		UserId:   userId.String(),
		Username: requestBody.Username,
		Name:     requestBody.Name,
		Password: hashed_password,
	}

	result := db.Create(&user)

	if err := result.Error; err != nil {
		log.Printf("Error while creating user: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created",
	})
}
