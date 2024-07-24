package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"spamcontrol/spamControl"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	sc := spamControl.NewSpamControl(redisClient, 60, 10)

	ok, err := sc.Check(context.Background(), 1)

	if err != nil {
		fmt.Println(err)
	} else if !ok {
		fmt.Println("user is a spamer")
	} else {
		fmt.Println("user is ok")
	}
}
