package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"task-management/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

const taskTTL = 5 * time.Minute
const listTTL = 1 * time.Minute

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func taskKey(id int64) string {
	return fmt.Sprintf("task:%d", id)
}

func listKey(offset, limit int64, search string) string {
	return fmt.Sprintf("tasks:%d:%d:%s", offset, limit, search)
}

type cachedTaskList struct {
	Tasks []models.Task `json:"tasks"`
	Total int64         `json:"total"`
}

func (c *RedisCache) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	data, err := c.client.Get(ctx, taskKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *RedisCache) SetTask(ctx context.Context, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, taskKey(task.Id), data, taskTTL).Err()
}

func (c *RedisCache) DeleteTask(ctx context.Context, id int64) error {
	return c.client.Del(ctx, taskKey(id)).Err()
}

func (c *RedisCache) GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error) {
	data, err := c.client.Get(ctx, listKey(offset, limit, search)).Bytes()
	if err != nil {
		return nil, 0, err
	}
	var cached cachedTaskList
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, 0, err
	}
	return cached.Tasks, cached.Total, nil
}

func (c *RedisCache) SetTasks(ctx context.Context, offset, limit int64, search string, tasks []models.Task, total int64) error {
	data, err := json.Marshal(cachedTaskList{Tasks: tasks, Total: total})
	if err != nil {
		return err
	}
	return c.client.Set(ctx, listKey(offset, limit, search), data, listTTL).Err()
}

func (c *RedisCache) InvalidateTasks(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, "tasks:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
