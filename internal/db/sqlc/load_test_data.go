package db

import (
	"context"
	"github.com/aalug/go-gin-job-search/pkg/utils"
	"github.com/bxcodec/faker/v3"
	"log"
	"sync"
	"sync/atomic"
)

// LoadTestData loads the test data into the database
func (store *SQLStore) LoadTestData(ctx context.Context) {
	var wg sync.WaitGroup
	nOfJobsCreated := int32(0)
	jobTitles := append(utils.GenerateEngineerJobs(), utils.GenerateDeveloperJobs()...)

	// create fake companies
	for i := 0; i < 5; i++ {
		for _, industry := range utils.Industries {
			// Increment the WaitGroup counter for each goroutine
			wg.Add(1)

			go func(industry string) {
				// Decrement the WaitGroup counter when the goroutine is done
				defer wg.Done()

				idx := utils.RandomInt(0, int32(len(utils.Locations)-1))
				location := utils.Locations[idx]
				companyParams := CreateCompanyParams{
					Name:     faker.DomainName(),
					Industry: industry,
					Location: location,
				}
				company, err := store.CreateCompany(ctx, companyParams)
				if err != nil {
					log.Println(err)
					return
				}

				var jobsCreated int32
				var jobWg sync.WaitGroup

				// create jobs
				for j := 0; j < 3; j++ {
					// Increment the job WaitGroup counter
					jobWg.Add(1)

					go func() {
						// Decrement the job WaitGroup counter when the job is done
						defer jobWg.Done()

						idx := utils.RandomInt(0, int32(len(jobTitles)-1))
						jobTitle := jobTitles[idx]
						jobParams := CreateJobParams{
							Title:        jobTitle,
							Industry:     industry,
							CompanyID:    company.ID,
							Description:  jobTitle + " " + faker.Paragraph(),
							Location:     location,
							SalaryMin:    utils.RandomInt(0, 2000),
							SalaryMax:    utils.RandomInt(2001, 5000),
							Requirements: jobTitle + " " + faker.Paragraph(),
						}
						_, err := store.CreateJob(ctx, jobParams)
						if err != nil {
							log.Println(err)
						} else {
							atomic.AddInt32(&jobsCreated, 1)
						}
					}()
				}

				jobWg.Wait() // Wait for all jobs to finish
				atomic.AddInt32(&nOfJobsCreated, jobsCreated)
			}(industry)
		}
	}

	wg.Wait() // Wait for all companies to finish
	log.Printf("Created %d jobs", nOfJobsCreated)
}
