package scheduler

import (
	"log"
	"math/rand"
	"time"

	"github.com/MonkaKokosowa/watchalong-server/api"
	"github.com/MonkaKokosowa/watchalong-server/logger"
	"github.com/robfig/cron/v3"
)

func StartScheduler() {
	c := cron.New()
	_, err := c.AddFunc("CRON_TZ=Europe/Warsaw 0 0 * * 0", func() {
		logger.Info("Running cron job to update vote")

		// Get vote results
		winners, err := api.GetVoteResults()
		if err != nil {
			logger.Error("Error getting vote results: ", err)
		} else {
			// Add winners to queue
			for _, winner := range winners {
				if err := winner.AddMovieToQueue(); err != nil {
					logger.Error("Error adding winner to queue: ", err)
				}
			}
		}

		// Clear old vote
		if err := api.ClearCurrentVote(); err != nil {
			logger.Error("Error clearing current vote: ", err)
		}
		if err := api.ClearVotes(); err != nil {
			logger.Error("Error clearing votes: ", err)
		}

		// Get unwatched movies not in queue
		movies, err := api.GetUnwatchedMoviesNotInQueue()
		if err != nil {
			logger.Error("Error getting unwatched movies: ", err)
			return
		}

		// Get 5 random movies
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(movies), func(i, j int) { movies[i], movies[j] = movies[j], movies[i] })

		var movieIDs []int
		for i := 0; i < 5 && i < len(movies); i++ {
			movieIDs = append(movieIDs, movies[i].ID)
		}

		// Create new vote
		if err := api.CreateNewVote(movieIDs); err != nil {
			logger.Error("Error creating new vote: ", err)
		}
		logger.Info("Cron job finished")
	})
	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}
	c.Start()
}
