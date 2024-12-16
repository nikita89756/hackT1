package main

import (
	"hack/internal/server"
	"hack/internal/storage/postgres"
	"hack/pkg/config"
	"hack/pkg/logger"
)
const(
	configPath = "./cmd/hack/config/config.yaml"
	logFilePath = "logs/app.log"
)
func main(){
	
	cfg:= config.LoadConfig(configPath)
	logger,loginfo,logerr:=log.InitLogger()
	defer loginfo.Close()
	defer logerr.Close()
	logger.Info("Starting server")
	storage := storage.GetDB(cfg.User,cfg.Password,cfg.Host,cfg.Port,cfg.DB,cfg.ShaSalt)
	s:=server.NewServer(cfg.Address,storage ,cfg.JWTKey,cfg.TokenTTL,logger,cfg.Mlconnection,cfg.ParserConnection)
	s.Start()
}

// func configLogger(){

// }