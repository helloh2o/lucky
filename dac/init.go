package dac

import "context"

// InitData 需要初始化的数据
func InitData() {
	if n, _ := Default().redisClient.Exists(context.Background(), GetNotOpenGamesKey()).Result(); n <= 0 {
		Default().redisClient.HSet(context.Background(), GetNotOpenGamesKey(), "init", "games")
	}
}
