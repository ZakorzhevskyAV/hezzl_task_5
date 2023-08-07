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

func GetGoodsList() (bool, types.List, int, int, error) {
	JSONGoodsList, _ := RedisClient.Get(ctx, "goods_list").Result()
	limitstr, _ := RedisClient.Get(ctx, "limit").Result()
	limit, err := strconv.Atoi(limitstr)
	if err != nil {
		return false, types.List{}, 0, 0, err
	}
	offsetstr, _ := RedisClient.Get(ctx, "offset").Result()
	offset, err := strconv.Atoi(offsetstr)
	if err != nil {
		return false, types.List{}, 0, 0, err
	}
	if JSONGoodsList == "" {
		return false, types.List{}, 0, 0, err
	}
	var goodsList types.List
	_ = json.Unmarshal([]byte(JSONGoodsList), &goodsList)

	return true, goodsList, limit, offset, err
}

func SetGoodsList(goodsList types.List, limit int, offset int) {
	JSONGoodsList, _ := json.Marshal(goodsList)
	RedisClient.Set(ctx, "goods_list", JSONGoodsList, time.Minute)
	RedisClient.Set(ctx, "limit", limit, time.Minute)
	RedisClient.Set(ctx, "offset", offset, time.Minute)
}

func InvalidateGoodsList() error {
	err := RedisClient.Del(ctx, "goods_list").Err()
	err = RedisClient.Del(ctx, "limit").Err()
	err = RedisClient.Del(ctx, "offset").Err()
	return err
}
