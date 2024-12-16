package storage

import (
	"crypto/sha1"
	_ "database/sql"
	"errors"
	"fmt"
	"hack/internal/storage"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)


type Storage struct{
	db *sqlx.DB
	salt string
}
func GetDB(username,password,host,port,dbname,salt string) Storage {
	connString := fmt.Sprintf(
		"user=%v password=%v host=%v port=%v dbname=%v sslmode=disable",
		username,password,host,port,dbname,
	)
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		panic(fmt.Sprintf("Error while connecting to DB. Error: %v", err.Error()))
	}

	return Storage{db,salt}
}
func (s Storage) GetUserByEmail(email , password string ) (storage.User, error) {
	user:= storage.User{Email:email}
	query := `SELECT user_id , password from Users where email = $1`
	row := s.db.QueryRow(query,email)
	if err:= row.Scan(&user.ID,&user.Password); err!=nil{
		return storage.User{},err
	}
	if user.Password != generatePasswordHash(password,s.salt){
		return storage.User{} , errors.New("Uncorrect password")
	}
	return user , nil
}




func (s Storage) SetNewUser(email, password string) (int,error) {
	user:= storage.User{Email:email}
	var userId int
	query := `SELECT user_id , password from Users where email = $1`
	row := s.db.QueryRow(query,email)
	if err:= row.Scan(&user.ID,&user.Password); err==nil{
		return 0,errors.New("user with such email is exists")
	}
	
	query = `INSERT INTO Users (email,password) VALUES ($1,$2) RETURNING user_id`
	row = s.db.QueryRow(query , email ,generatePasswordHash(password,s.salt))
	if err:=row.Scan(&userId); err!=nil{
		return 0,err
	}
	return userId , nil
	
}

func (s Storage) CreateChat(userId int ,chatName string,modelName string,instruction string,embending string)(int,error){
	var chatId int
	query:=`INSERT INTO chats (user_id,chat_name,model_name,instruction,embending) VALUES ($1,$2,$3,$4,$5) RETURNING chat_id`
	row := s.db.QueryRow(query,userId ,chatName ,modelName, instruction,embending)
	if err:=row.Scan(&chatId);err!=nil{
		return 0,err
	}
	return chatId,nil
}

func (s Storage) InsertMessage(chatId int , userId int ,question string , aiAnswer string)(int,error){
	var messageId int
	query:=`INSERT INTO history (chat_id,user_id,question,ai_answer) VALUES ($1,$2,$3,$4) RETURNING message_id`
	row := s.db.QueryRow(query,chatId,userId ,question ,aiAnswer )
	if err:=row.Scan(&messageId);err!=nil{
		return 0,err
	}
	return messageId,nil
}

func (s Storage) GetChatInformation(chatId int , userId int)(*storage.Chat,error){
	chat:= storage.Chat{ChatID: chatId,UserID: userId}
	query := `SELECT chat_name,model_name,instruction,created_at,embending from chats where chat_id= $1 and user_id = $2`
	row := s.db.QueryRow(query,chatId,userId)
	if err:= row.Scan(&chat.ChatName,&chat.ModelName,&chat.Instruction,&chat.CreatedAt); err!=nil{
		return &storage.Chat{},err
	}
	return &chat , nil
}

func (s Storage) GetAllChatsInformation(userId int)([]storage.Chat,error){
	var chats []storage.Chat
	query := `SELECT chat_id,user_id,chat_name,model_name,instruction,created_at,embending from chats where user_id = $1 `
	err := s.db.Select(&chats,query,userId)
	if err != nil{
		return []storage.Chat{},err
	}
	return chats, nil
}

func (s Storage) GetMessages(chatId int)([]storage.Message,error){
	var messages []storage.Message
	query := `SELECT message_id,chat_id,user_id,question,ai_answer ,sent_at from history where chat_id = $1 `
	err := s.db.Select(&messages,query,chatId)
	if err != nil{
		return []storage.Message{},err
	}
	return messages , nil
}



func generatePasswordHash(password,salt string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}


