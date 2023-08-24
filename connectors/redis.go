package connectors

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
)

func RedisConnection(host string, port string, pwd string, db *int) *redis.Client {
	defaultDb := 0
	if db == nil {
		db = &defaultDb
	}
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: pwd,
		DB:       *db,
	})
}

type RedisConnector struct {
	client *redis.Client
	ctx    context.Context
}

func (c *RedisConnector) Add(queueName string, id string, data []byte) error {
	return c.client.HSet(c.ctx, queueName, id, data).Err()
}

func (c *RedisConnector) GetQueueLength(queueName string) (int64, error) {
	return c.client.HLen(c.ctx, queueName).Result()
}

func (c *RedisConnector) GetFirstElement(queueName string) (string, []byte, error) {
	cursor := uint64(0)
	count := int64(1)
	result, _, err := c.client.HScan(c.ctx, queueName, cursor, "*", count).Result()
	if err != nil {
		return "", nil, err
	}
	if len(result) < 2 {
		return "", nil, fmt.Errorf("wrong data inside queue")
	}
	return result[0], []byte(result[1]), err
}

func (c *RedisConnector) RemoveById(queueName string, id string) error {
	err := c.client.HDel(c.ctx, queueName, id).Err()
	return err
}

func (c *RedisConnector) RemoveFirstElement(queueName string) ([]byte, error) {
	id, data, err := c.GetFirstElement(queueName)
	if err != nil {
		return nil, err
	}
	err = c.RemoveById(queueName, id)
	return data, err

}

func (c *RedisConnector) Trim(queueName string, start int64, end int64) error {
	hashFields, err := c.client.HKeys(c.ctx, queueName).Result()
	if err != nil {
		return err
	}

	if len(hashFields) <= int(end) {
		end = int64(len(hashFields) - 1)
	}

	fieldsToDelete := hashFields[start:end]
	if len(fieldsToDelete) == 0 {
		return nil
	}

	_, err = c.client.HDel(c.ctx, queueName, fieldsToDelete...).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *RedisConnector) NewProgressiveId(queueName string) (string, error) {
	newID, err := c.client.Incr(c.ctx, queueName+":id").Result()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(newID)), nil
}

func NewRedisConnector(ctx context.Context, client *redis.Client) Connector {
	return &RedisConnector{ctx: ctx, client: client}
}
