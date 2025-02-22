package userService

import (
	"fmt"
	"go-web-socket/config"
	"go-web-socket/internal/models"
)

func EditUser(userId string, data models.User) (models.User, error) {
	db, err := config.GetDatabaseConnection()

	var user models.User

	if err != nil {
		return user, fmt.Errorf("erro ao abrir conexção com banco de dados: %v", err.Error())
	}

	result := db.Model(&models.User{}).Where("user_id = ?", userId).Updates(&models.User{
		Name:   data.Name,
		Avatar: data.Avatar,
	}).Scan(&user)

	if result.RowsAffected == 0 {
		return user, fmt.Errorf("usuário não encontrado")
	}

	if err := result.Error; err != nil {
		return user, fmt.Errorf("erro ao atualizar usuário: %v", err)

	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	return user, nil

}

func UpdateUserAvatar(userId string, fileurl string) error {
	db, err := config.GetDatabaseConnection()

	if err != nil {
		return fmt.Errorf("erro ao abrir conexção com banco de dados: %v", err.Error())

	}

	result := db.Model(&models.User{}).Where("user_id = ?", userId).Updates(&models.User{
		Avatar: fileurl,
	})

	if result.RowsAffected == 0 {
		return fmt.Errorf("usúario não econtrado")
	}

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	return nil

}
