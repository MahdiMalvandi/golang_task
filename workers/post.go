package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang_task/repositories"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)
// This Function Gets Posts from queue and Adds to Author's Followers timeline
func FanOutWorker(rdb *redis.Client, db *gorm.DB) {
	ctx := context.Background()
	queueKey := "post_queue"

	// Create follow repository
	followRepo := repositories.NewFollowRepository(db)

	for {
		// Listen to queue
		result, err := rdb.BLPop(ctx, 0*time.Second, queueKey).Result()
		if err != nil {
			fmt.Printf("Worker error: %v\n", err)
			continue
		}
		var resultMap struct {
			PostID   uint `json:"post_id"`
			AuthorID uint `json:"author_id"`
			Created  int64 `json:"created_at"` 
		}
		json.Unmarshal([]byte(result[1]), &resultMap)

		// Get Author's followers
		authorFollowers, err := followRepo.GetFollowers(resultMap.AuthorID)

		if err != nil {
			fmt.Printf("Worker error: %v\n", err)
			continue
		}

		// Add posts to Author's followers timeline
		for _, follower := range authorFollowers {
			err := rdb.ZAdd(ctx, fmt.Sprintf("timeline:%d", follower.ID), redis.Z{Score: float64(resultMap.Created), Member: resultMap.PostID}).Err()
			if err != nil {
				fmt.Printf("Worker error: %v\n", err)
				continue
			}
			fmt.Printf("timeline:%d", follower.ID)

		}

	}
}


