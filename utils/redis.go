package utils

import (
    "context"
    "time"

    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb *redis.Client

func InitRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379", 
        Password: "",           
        DB: 0,                  
    })
}

func SetOTP(email, otp string) error {
    err := rdb.Set(ctx, email, otp, 10*time.Minute).Err() 
    if err != nil {
        return err
    }
    return nil
}

func GetOTP(email string) (string, error) {
    otp, err := rdb.Get(ctx, email).Result()
    if err != nil {
        return "", err
    }
    return otp, nil
}
