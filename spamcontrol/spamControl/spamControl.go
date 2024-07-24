package spamControl

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

type SpamControl struct {
	redisClient *redis.Client
	timeSize    int
	maxRequests int
}

func NewSpamControl(redisClient *redis.Client, timeSize int, maxRequests int) *SpamControl {
	return &SpamControl{
		redisClient: redisClient,
		timeSize:    timeSize,
		maxRequests: maxRequests,
	}
}

func (sc *SpamControl) Check(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("floodcontrol:%d", userID)

	now := time.Now().Unix()

	sc.redisClient.ZRemRangeByScore(key, "0", strconv.FormatInt(now-int64(sc.timeSize), 10))

	sc.redisClient.ZAdd(key, redis.Z{Score: float64(now), Member: now})

	count, err := sc.redisClient.ZCard(key).Result()
	if err != nil {
		return false, err
	}

	return count <= int64(sc.maxRequests), nil
}
