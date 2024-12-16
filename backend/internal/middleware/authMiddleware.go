package middleware

import (
	"hack/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Middleware struct{
	Token string
}

func (m *Middleware)AuthMiddleware() gin.HandlerFunc{
	return func(ctx *gin.Context){
		token := ctx.GetHeader("Authorization")
		if token == ""{
			ctx.JSON(http.StatusUnauthorized , gin.H{"error":"lost token"})
			ctx.Abort()
			return

		}
		id , err:= service.ParseToken(m.Token,token)
		if err!=nil{
			ctx.JSON(http.StatusUnauthorized , gin.H{"error":err.Error()})
			ctx.Abort()
			return
		}
		ctx.Set("userId",id)
	}
}