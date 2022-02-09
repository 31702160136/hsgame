package typedefine

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func BindParam(ctx *gin.Context, v interface{}) error {
	switch ctx.Request.Method {
	case "GET":
		if err := ctx.BindQuery(v); err != nil {
			return err
		}
	case "POST":
		if err := ctx.ShouldBind(v); err != nil {
			return err
		}
	}
	return nil
}

func Reply(ctx *gin.Context, data interface{}, code int, msgs ...string) {
	msg := strings.Join(msgs, " ")
	ctx.JSON(200, gin.H{
		"msg":  msg,
		"code": code,
		"data": data,
	})
}
