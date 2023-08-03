package redis_caching

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"hezzl_task_5/types"
	"os"
	"strconv"
	"time"
)

var ctx = context.Background()
var RedisClient *redis.Client

func RedisConnect() error {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "",
		DB:       db,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

func GetGoods(key string) (bool, []types.Goods) {
	jsonData, _ := RedisClient.Get(ctx, key).Result()
	if jsonData == "" {
		return false, nil
	}

	var resp []types.Goods
	_ = json.Unmarshal([]byte(jsonData), &resp)

	return true, resp
}

func SetGoods(key string, goods []types.Goods) {
	jsonData, _ := json.Marshal(goods)
	RedisClient.Set(ctx, key, jsonData, time.Minute)
}

func InvalidateGoods() error {
	err := RedisClient.Del(ctx, "goods").Err()
	return err
}
