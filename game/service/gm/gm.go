package gm

import (
	"fmt"
	"game/dispatch"
	"game/log"
	"game/service"
	t "game/typedefine"
	"github.com/gin-gonic/gin"
	"sync"
)

func GmHandle(ctx *gin.Context) {
	gmName := ctx.Param("name")
	handle := service.GetGm(gmName)
	if handle == nil {
		t.Reply(ctx, -1, fmt.Sprintf("gm %s is nil", gmName))
		return
	}
	data := map[string]string{}
	if err := t.BindParam(ctx, &data); err != nil {
		t.Reply(ctx, -1, fmt.Sprintf("data format err: %s", err.Error()))
		return
	}
	var code int
	var out interface{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	dispatch.PushSystemSyncMsg(fmt.Sprintf("gmHandle:%s", gmName), func() {
		code, out = handle(data)
		wg.Done()
	})
	wg.Wait()
	t.Reply(ctx, code, out)
	log.Infof("handleGM code:%d name:%s\n%v", code, gmName, out)
}
