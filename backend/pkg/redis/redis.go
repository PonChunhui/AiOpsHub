package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var Client redis.UniversalClient

type RedisClient struct {
	Client redis.UniversalClient
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		Client: Client,
	}
}

type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"`
}

func Init() error {
	clusterMode := viper.GetBool("redis.cluster_mode")

	if clusterMode {
		clusterNodes := viper.GetStringSlice("redis.cluster_nodes")
		if len(clusterNodes) == 0 {
			host := viper.GetString("redis.host")
			port := viper.GetInt("redis.port")
			clusterNodes = []string{fmt.Sprintf("%s:%d", host, port)}
		}

		password := viper.GetString("redis.password")

		Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    clusterNodes,
			Password: password,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := Client.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("failed to connect to Redis Cluster at %v: %w", clusterNodes, err)
		}

		return nil
	}

	host := viper.GetString("redis.host")
	port := viper.GetInt("redis.port")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")

	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s:%d: %w", host, port, err)
	}

	return nil
}

func SetToken(ctx context.Context, token string, info *TokenInfo, expire time.Duration) error {
	key := fmt.Sprintf("token:%s", token)
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}
	return Client.Set(ctx, key, data, expire).Err()
}

func GetToken(ctx context.Context, token string) (*TokenInfo, error) {
	key := fmt.Sprintf("token:%s", token)
	data, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var info TokenInfo
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token info: %w", err)
	}

	return &info, nil
}

func DeleteToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("token:%s", token)
	return Client.Del(ctx, key).Err()
}

func ExistsToken(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("token:%s", token)
	result, err := Client.Exists(ctx, key).Result()
	return result > 0, err
}

func GetUserTokens(ctx context.Context, userID string) ([]string, error) {
	pattern := fmt.Sprintf("token:*")
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var userTokens []string
	for _, key := range keys {
		info, err := GetToken(ctx, key[6:])
		if err != nil {
			continue
		}
		if info.UserID == userID {
			userTokens = append(userTokens, key[6:])
		}
	}

	return userTokens, nil
}

func (rc *RedisClient) Publish(ctx context.Context, channel string, message string) error {
	return rc.Client.Publish(ctx, channel, message).Err()
}

func (rc *RedisClient) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return rc.Client.Subscribe(ctx, channel)
}

func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.Client.Set(ctx, key, value, expiration).Err()
}

func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.Client.Get(ctx, key).Result()
}

func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.Client.Del(ctx, keys...).Err()
}

func (rc *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return rc.Client.SetNX(ctx, key, value, expiration).Result()
}

func (rc *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.Client.Exists(ctx, keys...).Result()
}

func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.Client.Expire(ctx, key, expiration).Err()
}

func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.Client.TTL(ctx, key).Result()
}
