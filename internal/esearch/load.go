package esearch

import (
	"context"
	"github.com/aalug/job-finder-go/internal/db/sqlc"
	"log"
	"sync"
)

func LoadJobsFromDB(ctx context.Context, store db.Store) (context.Context, error) {
	const (
		concurrency = 5
	)

	var (
		jobs      []Job
		waitGroup = new(sync.WaitGroup)
		workQueue = make(chan Job)
		mutex     = &sync.Mutex{}
	)

	// Fetch jobs from the database
	jobsFromDB, err := store.ListAllJobsForES(ctx)
	if err != nil {
		return nil, err
	}

	// Populate the work queue with movies from the database.
	go func() {
		for _, job := range jobsFromDB {
			skills, err := store.ListAllJobSkillsByJobID(ctx, job.ID)
			if err != nil {
				panic(err)
			}
			j := Job{
				ID:           job.ID,
				Title:        job.Title,
				Industry:     job.Industry,
				CompanyName:  job.CompanyName,
				Description:  job.Description,
				Location:     job.Location,
				SalaryMin:    job.SalaryMin,
				SalaryMax:    job.SalaryMax,
				Requirements: job.Requirements,
				JobSkills:    skills,
			}
			workQueue <- j
		}
		close(workQueue)
	}()

	for i := 0; i < concurrency; i++ {
		waitGroup.Add(1)
		go func(workQueue chan Job, waitGroup *sync.WaitGroup) {
			for job := range workQueue {
				mutex.Lock()
				jobs = append(jobs, job)
				mutex.Unlock()
			}
			waitGroup.Done()
		}(workQueue, waitGroup)
	}

	waitGroup.Wait()

	log.Printf("Jobs loaded from the database: %d\n", len(jobs))
	return context.WithValue(ctx, JobKey, jobs), nil
}
