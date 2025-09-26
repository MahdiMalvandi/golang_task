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

func StartWorker(rdb *redis.Client, db *gorm.DB) {
	ctx := context.Background()
	queueKey := "post_queue"
	followRepo := repositories.NewFollowRepository(db)
	fmt.Println("worker started")
	_ = followRepo
	for {
		fmt.Println("worker starts to work")

		result, err := rdb.BLPop(ctx, 0*time.Second, queueKey).Result()
		if err != nil {
			fmt.Printf("Worker error: %v\n", err)
			continue
		}
		var resultMap map[string]uint
		json.Unmarshal([]byte(result[1]), &resultMap)

		authorFollowers, err := followRepo.GetFollowers(resultMap["author_id"])

		if err != nil {
			fmt.Printf("Worker error: %v\n", err)
			continue
		}

		for _, follower := range authorFollowers {
			err := rdb.ZAdd(ctx, fmt.Sprintf("timeline:%d", follower.ID), redis.Z{Score: float64(time.Now().Unix()), Member: resultMap["post_id"]}).Err()
			if err != nil {
				fmt.Printf("Worker error: %v\n", err)
				continue
			}
			fmt.Printf("timeline:%d", follower.ID)

		}

	}
}


