package route

import (
	"context"
	"fmt"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/cache"
	"github.com/helloh2o/lucky/im/server/constants"
	"github.com/helloh2o/lucky/im/server/immsg"
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/natsq"
	"github.com/helloh2o/lucky/utils"
	v12c "github.com/kataras/iris/v12/context"
	"time"
)

func WriteData(ctx *v12c.Context, code int, data interface{}) {
	warp := map[string]interface{}{
		"code": code,
		"data": data,
	}
	err := ctx.JSON(&warp)
	log.Error(err)
}

func InitHandler() {
	Processor.RegisterHandler(PeerMsg, &immsg.PeerMsg{}, func(args ...interface{}) {
		msg := args[lucky.Msg].(*immsg.PeerMsg)
		msg.TimeFmt = utils.FormatTime2String(time.Now())
		ctx := args[lucky.Conn].(*v12c.Context)
		raw := args[lucky.Raw].([]byte)
		err := pub(msg.ToUser, raw)
		if err != nil {
			WriteData(ctx, constants.ErrorNormal, err.Error())
		} else {
			WriteData(ctx, constants.SUCCESS, msg)
		}
	})
	Processor.RegisterHandler(GroupMsg, &immsg.PeerGroupMsg{}, func(args ...interface{}) {
		msg := args[lucky.Msg].(*immsg.PeerGroupMsg)
		msg.TimeFmt = utils.FormatTime2String(time.Now())
		ctx := args[lucky.Conn].(*v12c.Context)
		raw := args[lucky.Raw].([]byte)
		err := pub(msg.ToGroup, raw)
		if err != nil {
			WriteData(ctx, constants.ErrorNormal, err.Error())
		} else {
			WriteData(ctx, constants.SUCCESS, msg)
		}
	})
}

func pub(subject string, data []byte) error {
	if err := natsq.Pub(subject, data, false); err != nil {
		// 离线消息
		key := fmt.Sprintf(constants.MsgRedisWaitList, subject)
		return cache.RedisC.RPush(context.Background(), key, string(data)).Err()
	}
	return nil
}
