package utils

import (
	"context"
	"encoding/json"
	"golang_task/models"

	"github.com/redis/go-redis/v9"
)

func PostQueue(post *models.Post, rdb *redis.Client, add bool) {
	// Send post id and author id to queue to delete from followers timeline
	queueKey := "post_queue"
	data, _ := json.Marshal(map[string]interface{}{
		"post_id":    post.ID,
		"author_id":  post.AuthorID,
		"created_at": uint(post.CreatedAt.Unix()),
		"is_add":     add,
	})

	rdb.RPush(context.Background(), queueKey, data)
	
}