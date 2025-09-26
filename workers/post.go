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
	followRepo := repositories.NewFollowRepository(db, rdb)
	fmt.Println("[INFO] FanOutWorker started, listening to queue:", queueKey)

	for {
		// Listen to queue
		result, err := rdb.BLPop(ctx, 0*time.Second, queueKey).Result()
		if err != nil {
			fmt.Printf("[ERROR] Failed to pop from queue: %v\n", err)

			continue
		}
		fmt.Println("[INFO] Received item from queue:", result[1])
		var resultMap struct {
			PostID   uint  `json:"post_id"`
			AuthorID uint  `json:"author_id"`
			Created  int64 `json:"created_at"`
			IsAdd    bool  `json:"is_add"`
		}
		if err := json.Unmarshal([]byte(result[1]), &resultMap); err != nil {
			fmt.Printf("[ERROR] Failed to unmarshal queue item: %v\n", err)
			continue
		}
		fmt.Printf("[INFO] Parsed queue item: PostID=%d, AuthorID=%d, Created=%d\n IsAdd=%t", resultMap.PostID, resultMap.AuthorID, resultMap.Created, resultMap.IsAdd)

		// Get Author's followers
		authorFollowers, err := followRepo.GetFollowers(resultMap.AuthorID)

		if err != nil {
			fmt.Printf("[ERROR] Failed to get followers for author %d: %v\n", resultMap.AuthorID, err)

			continue
		}

		fmt.Printf("[INFO] Author %d has %d followers\n", resultMap.AuthorID, len(authorFollowers))

		for _, follower := range authorFollowers {
			key := fmt.Sprintf("timeline:%d", follower.ID)
			if resultMap.IsAdd {
			
				err := rdb.ZAdd(ctx, key, redis.Z{
					Score:  float64(resultMap.Created),
					Member: resultMap.PostID,
				}).Err()
				if err != nil {
					fmt.Printf("[ERROR] Failed to add Post %d to timeline:%d: %v\n", resultMap.PostID, follower.ID, err)
				}
			} else {
				
				err := rdb.ZRem(ctx, key, resultMap.PostID).Err()
				if err != nil {
					fmt.Printf("[ERROR] Failed to remove Post %d from timeline:%d: %v\n", resultMap.PostID, follower.ID, err)
				}
			}
		}
		fmt.Println("[INFO] Finished processing queue item for PostID:", resultMap.PostID)

	}
}
