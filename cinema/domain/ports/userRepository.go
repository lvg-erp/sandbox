package ports

import "cinema/domain/entities"

// Пользователи

type UserRepository interface {
	CreateUser(entities.User) error
	GetUser(email string) (entities.User, error)
	GetUserByID(id int) (entities.User, error)
	DeleteUserSessions(userID int) error
}
