package service

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)



type tokenClaims struct{
	jwt.StandardClaims
	UserId int `json:"user_id"`
}


func GenerateToken(username,password string , id int,jwtkey string , tokenTTL time.Duration)(string,error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,&tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		id,
	})
	return token.SignedString([]byte(jwtkey))
}

func ParseToken(jwtToken,token string)(int , error){
	t,err:= jwt.ParseWithClaims(token,&tokenClaims{},func(t *jwt.Token)(interface{},error){
		if _,ok := t.Method.(*jwt.SigningMethodHMAC);!ok{
			return nil , errors.New("invalid signing method")
		}
		return []byte(jwtToken),nil
	})

	if err!=nil{
		return 0 , err
	}

	claims , ok := t.Claims.(*tokenClaims)
	if !ok{
		return 0 , errors.New("bad token")
	}
	
	return claims.UserId , nil
}