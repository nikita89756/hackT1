package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "hack/docs"
	"hack/internal/middleware"
	"hack/internal/service"
	"hack/internal/service/parser"
	"hack/internal/storage"
	log "hack/pkg/logger"
	_ "hack/pkg/models/swagger"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title JWT Authentication API
// @version 1.0
// @description This is a JWT authentication service with registration and signin capabilities
// @host localhost:8080
// @BasePath /
type Server struct {
	port     string
	storage  storage.Storage
	jwtToken string
	tokenTTL time.Duration
	logger   *log.Logs
	mlconnection  string
	parserconnection string
}

func NewServer(port string, storage storage.Storage, jwtToken string, TokenTTL time.Duration, logger *log.Logs,mlconnection string,parserconnection string) *Server {
	return &Server{port: port, storage: storage, jwtToken: jwtToken, tokenTTL: TokenTTL, logger: logger,mlconnection: mlconnection,parserconnection: parserconnection}
}

func (s *Server) Start() error {
	m := middleware.Middleware{s.jwtToken}
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(middleware.CORSMiddleware())
	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	auth := engine.Group("/auth")
	{
		auth.PUT("/registration", s.registrationHandler)
		auth.GET("/signin", s.signInHandler)
	}
	api := engine.Group("/api", m.AuthMiddleware())
	{
		api.GET("/health", health)
		api.GET("/sendMessage", s.sendMessage)
		api.POST("/createModelFromFile", s.createModelFromFile)
		api.POST("/createModelFromURL", s.createModelFromURL)
		api.GET("/getHistory", s.getHistory)
		api.GET("/getAllChats", s.getAllChats)
	}

	err := engine.Run(s.port)
	return err
}

// @Summary Create model from URL
// @Description Create a model from a document URL
// @Tags api
// @Accept json
// @Produce json
// @Param request body RequestChat true "URL of the document"
// @Success 200 {object} Chat
// @Failure 400 {object} modelsSwagger.Error
// @Router /api/createModelFromURL [get]
// @Security ApiKeyAuth
func (s *Server) createModelFromURL(ctx *gin.Context) {
	id, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No id"})
		return
	}
	var modelInfo RequestChat
	if err := ctx.BindJSON(&modelInfo); err != nil {
		s.logger.Error("Ошибка в createModelFromUrl")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chatID, err := s.storage.CreateChat(id.(int), modelInfo.ChatName, modelInfo.ModelName, modelInfo.Instruction, modelInfo.Embending)
	if err != nil {
		s.logger.Error("Ошибка в createModelFromUrl при попытке создать новый чат")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	text, err := parser.ExtractText(modelInfo.URL,s.parserconnection)
	if err != nil {
		s.logger.Error("Ошибка в  при попытке получить конвертированный текст")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	var requestBody bytes.Buffer

	multipartWriter := multipart.NewWriter(&requestBody)

	fileWriter, err := multipartWriter.CreateFormFile("file", "textfile.txt")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	_, err = io.Copy(fileWriter, bytes.NewBufferString(text))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	chatIDWriter, err := multipartWriter.CreateFormField("file_name")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	_, err = chatIDWriter.Write([]byte(strconv.Itoa(chatID)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Закрытие multipart writer
	err = multipartWriter.Close()
	if err != nil {
		fmt.Printf("Ошибка при закрытии multipart writer: %v\n", err)
		return
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "http://"+s.mlconnection+"/create_faiss_index", &requestBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Установка заголовка Content-Type
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Отправка HTTP-запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	if resp.StatusCode == http.StatusInternalServerError {
		ctx.Status(http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	ctx.JSON(http.StatusOK, gin.H{"text": text, "chatId": chatID})
}

// @Summary Create model from file
// @Description Create a model from an uploaded file
// @Tags api
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to process"
// @Success 200 {object} Chat
// @Failure 400 {object} modelsSwagger.Error
// @Router /api/createModelFromFile [get]
// @Security ApiKeyAuth
func (s *Server) createModelFromFile(ctx *gin.Context) {

	id, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No id"})
		return
	}
	var modelInfo ModelInfo
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}
	modelInfo.ChatName = ctx.PostForm("chat_name")
	modelInfo.ModelName = ctx.PostForm("model_name")
	modelInfo.Instruction = ctx.PostForm("instruction")
	modelInfo.Embending = ctx.PostForm("embending")

	chatID, err := s.storage.CreateChat(id.(int), modelInfo.ChatName, modelInfo.ModelName, modelInfo.Instruction, modelInfo.Embending)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}
	fileType := file.Header.Get("Content-Type")
	// if err:= ctx.BindJSON(&modelInfo);err!=nil{
	// 	ctx.JSON(http.StatusBadRequest,gin.H{"error": fmt.Sprintf("Error while parsing Json:%v",err.Error())})
	// 	return
	// }
	dst := "./uploads/" + file.Filename

	// Создаем директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads/", os.ModePerm); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	// Сохраняем файл на сервере
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	text, err := parser.ExtractText(dst,s.parserconnection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	os.Remove(dst)
	var requestBody bytes.Buffer

	// Создание multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Создание поля для файла
	fileWriter, err := multipartWriter.CreateFormFile("file", "textfile.txt")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Запись текста в поле для файла
	_, err = io.Copy(fileWriter, bytes.NewBufferString(text))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	chatIDWriter, err := multipartWriter.CreateFormField("chat_id")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Запись user_id в поле
	ch := chatID
	_, err = chatIDWriter.Write([]byte(strconv.Itoa(chatID)))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	modelNameWriter, err := multipartWriter.CreateFormField("model_name")
	_, err = modelNameWriter.Write([]byte(modelInfo.ModelName))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Закрытие multipart writer
	err = multipartWriter.Close()
	if err != nil {
		fmt.Printf("Ошибка при закрытии multipart writer: %v\n", err)
		return
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "http://"+s.mlconnection+"/create_faiss_index", &requestBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}

	// Установка заголовка Content-Type
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Отправка HTTP-запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Ошибка в при попытке получить конвертированный текств")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error while parsing data:%v", err.Error())})
		return
	}
	defer resp.Body.Close()

	ctx.JSON(http.StatusOK, gin.H{"file_type": fileType, "text": text, "chat_id": ch})
}

// @Summary Send message
// @Description Send a message and get a response
// @Tags api
// @Accept json
// @Produce json
// @Param request body RequestMessage true "Message details"
// @Success 200 {object} ResponseTest
// @Failure 400 {object} modelsSwagger.Error
// @Router /api/setMessage [get]
// @Security ApiKeyAuth
func (s *Server) sendMessage(ctx *gin.Context) {
	var req RequestMessage
	var resp ResponseTest
	id, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No id"})
		return
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//TODO когда модель будет готова надо отправлять instruction modelName emdending
	jsonData,err:=json.Marshal(req)
	answer, err := http.NewRequest("POST", "http://"+s.mlconnection+"/answering", bytes.NewBuffer(jsonData))
	answer.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(answer)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Bad answer from ml"})
		return
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO вставка в таблицы сообщений
	_, err = s.storage.InsertMessage(req.ChatId, id.(int), req.Question, resp.Answer)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	fmt.Print(id)
	ctx.JSON(http.StatusOK, gin.H{"answer": resp.Answer})
}

// @Summary Get all chats
// @Description Get all chats for the authenticated user
// @Tags api
// @Accept json
// @Produce json
// @Success 200 {array} []storage.Chat
// @Failure 400 {object} modelsSwagger.Error
// @Failure 503 {object} modelsSwagger.Error
// @Router /api/getAllChats [get]
// @Security ApiKeyAuth
func (s *Server) getAllChats(ctx *gin.Context) {
	id, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No id"})
		return
	}
	chats, err := s.storage.GetAllChatsInformation(id.(int))
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"chats": chats})
}

// @Summary Get all messages from chat
// @Description Get all messages from a specific chat
// @Tags api
// @Accept json
// @Produce json
// @Param request body Chat true "Message details"
// @Success 200 {array} []storage.Message
// @Failure 400 {object} modelsSwagger.Error
// @Failure 503 {object} modelsSwagger.Error
// @Router /api/getHistory [get]
// @Security ApiKeyAuth
func (s *Server) getHistory(ctx *gin.Context) {
	var chat Chat
	_, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No id"})
		return
	}
	if err := ctx.BindJSON(&chat); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	messages, err := s.storage.GetMessages(chat.ChatId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body signinUser true "User registration details"
// @Success 200 {object} modelsSwagger.Error
// @Failure 400 {object} modelsSwagger.Error
// @Router /auth/registration [put]
func (s *Server) registrationHandler(ctx *gin.Context) {
	var user signinUser
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := s.storage.SetNewUser(user.Email, user.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"error": ""})
}

// @Summary User sign in
// @Description Authenticate a user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body signinUser true "User credentials"
// @Success 200 {object} modelsSwagger.Token
// @Failure 400 {object} modelsSwagger.Error
// @Router /auth/signin [get]
func (s *Server) signInHandler(ctx *gin.Context) {
	var user signinUser
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := s.storage.GetUserByEmail(user.Email, user.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := service.GenerateToken(user.Email, user.Password, u.ID, s.jwtToken, s.tokenTTL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 "OK"
// @Router /api/health [get]
// @Security ApiKeyAuth
func health(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

type signinUser struct {
	Email    string `json:"email" example:"user@example.com" binding:"required"`
	Password string `json:"password" example:"password123" binding:"required"`
}

type RequestMessage struct {
	ChatId   int    `json:"chat_id"`
	Question string `json:"question"`
	ModelName string `json:"model_name"`
}

type ModelInfo struct {
	ChatName    string `json:"chat_name"`
	ModelName   string `json:"model_name"`
	Instruction string `json:"instruction"`
	Embending   string `json:"embending"`
}

type RequestChat struct {
	URL         string `json:"url"`
	ChatName    string `json:"chat_name"`
	ModelName   string `json:"model_name"`
	Instruction string `json:"instruction"`
	Embending   string `json:"embending"`
}

type RequestMessageTest struct {
	Question string `json:"question"`
	FileName string `json:"file_name"`
	ModelName string `json:"model_name"`
}

type ResponseTest struct {
	Answer string `json:"answer"`
}

type Chat struct {
	ChatId int `json:"chat_id"`
}
