package storage

import "time"

// import (
// 	"crypto/sha1"
// 	"errors"
// 	"fmt"
// )

// const (
// 	salt = "hgrtduaklnc789ajgsdt123jasty"
// )

type User struct {
	ID       int    `json: user_id`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Chat struct {
	ChatID      int       `db:"chat_id"`
	UserID      int       `db:"user_id"`
	ChatName    string    `db:"chat_name"`
	ModelName   string    `db:"model_name"`
	Embending		string		`db:"embending"`
	FileURL     string    `db:"file_url"`
	Instruction string    `db:"instruction"`
	CreatedAt   time.Time `db:"created_at"`
}

type Message struct {
	MessageID int       `db:"message_id"`
	ChatID    int       `db:"chat_id"`
	UserID    int       `db:"user_id"`
	Question  string    `db:"question"`
	AIAnswer  string    `db:"ai_answer"`
	SentAt    time.Time `db:"sent_at"`
}

// type Storage struct {
// 	storage map[string]User
// }

// func NewStorage() *Storage {
// 	return &Storage{make(map[string]User)}
// }

// func (s *Storage) SetNewUser(email, password string) error {
// 	if _, exist := s.storage[email]; exist {
// 		return errors.New("User with this email is exists")
// 	}
// 	user := User{
// 		Email:    email,
// 		Password: generatePasswordHash(password),
// 		ID:       1,
// 	}
// 	s.storage[email] = user
// 	return nil
// }

// func (s *Storage) GetUserByEmail(email , password string ) (User, error) {
// 	user, exists := s.storage[email]
// 	if !exists{
// 		return User{} , errors.New("user with such email is not exists")
// 	}

// 	if user.Password != generatePasswordHash(password){
// 		return User{} , errors.New("Uncorrect password")
// 	}
// 	return user , nil
// }

// func generatePasswordHash(password string) string {
// 	hash := sha1.New()
// 	hash.Write([]byte(password))

//		return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
//	}
type Storage interface {
	GetUserByEmail(string, string) (User, error)
	SetNewUser(string, string) (int, error)
	CreateChat(userId int ,chatName string,modelName string,instruction string,embending string)(int,error)
	InsertMessage(chatId int , userId int ,question string , aiAnswer string)(int,error)
	GetChatInformation(chatId int , userId int)(*Chat,error)
	GetAllChatsInformation(userId int)([]Chat,error)
	GetMessages(chatId int)([]Message,error)
}
