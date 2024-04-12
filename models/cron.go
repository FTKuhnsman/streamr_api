package models

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"streamr_api/common"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Scheduler struct {
	mu          sync.Mutex
	Jobs        map[string]*CronJob `json:"jobs"`
	Cron        *cron.Cron          `json:"-"`
	cronJobFile string
}

type CronJob struct {
	Name     string       `json:"name" example:"sample job"`
	Schedule string       `json:"schedule" example:"0/5 * * * * *"`
	Endpoint string       `json:"endpoint" example:"/api/v1/sample-endpoint"`
	Method   string       `json:"method" example:"GET"`
	Enabled  bool         `json:"enabled" example:"true"`
	EntryID  cron.EntryID `json:"-"`
}

func NewScheduler() *Scheduler {
	cron.WithSeconds()
	scheduler := Scheduler{
		mu:          sync.Mutex{},
		Jobs:        make(map[string]*CronJob),
		Cron:        cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger))),
		cronJobFile: common.GetStringEnvWithDefault("CRON_JOB_FILE", "cron_jobs.json"),
	}
	// Load existing jobs
	err := scheduler.LoadCronJobs()
	if err != nil {
		panic(err)
	}

	scheduler.Cron.Start()

	return &scheduler
}

func (s *Scheduler) LoadCronJobs() error {
	var jobs map[string]*CronJob

	// Check if the file exists
	_, err := os.Stat(s.cronJobFile)
	if os.IsNotExist(err) {
		return nil // Return an empty list if the file does not exist
	}

	// Read the file
	bytes, err := os.ReadFile(s.cronJobFile)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the jobs slice
	err = json.Unmarshal(bytes, &jobs)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.Jobs = jobs
	s.mu.Unlock()

	// Schedule the jobs if enabled
	s.mu.Lock()
	for _, job := range s.Jobs {
		if job.Enabled {
			err = s.ScheduleJob(job)
			if err != nil {
				log.Printf("Failed to schedule job %s: %v", job.Name, err)
			}
		}
	}
	s.mu.Unlock()

	return nil
}

// SaveCronJobs saves the provided cron jobs into a JSON file.
func (s *Scheduler) SaveCronJobs() error {
	s.mu.Lock()
	bytes, err := json.Marshal(s.Jobs)
	s.mu.Unlock()
	if err != nil {
		return err
	}

	// Write to the file
	return os.WriteFile(s.cronJobFile, bytes, 0644)
}

func (s *Scheduler) CreateCronJob(job *CronJob) error {
	// Add the job to the scheduler
	randID, err := common.GenerateRandomHexString(16)
	if err != nil {
		return errors.New("Failed to generate random ID")
	}

	s.mu.Lock()
	s.Jobs[randID] = job
	s.mu.Unlock()

	err = s.SaveCronJobs()
	if err != nil {
		return errors.New("Failed to save cron jobs")
	}

	// Schedule the job using the scheduler
	if job.Enabled {
		err = s.ScheduleJob(job)
		if err != nil {
			return errors.New("Failed to schedule job")
		}
	}

	return err
}

func (s *Scheduler) GetCronJobsCopy() map[string]CronJob {
	// create a copy of jobs map
	jobscopy := make(map[string]CronJob)
	s.mu.Lock()
	for k, v := range s.Jobs {
		jobscopy[k] = *v
	}
	s.mu.Unlock()

	return jobscopy
}

func (s *Scheduler) GetCronJob(id string) *CronJob {
	s.mu.Lock()
	job := s.Jobs[id]
	s.mu.Unlock()

	return job
}

func (s *Scheduler) ScheduleJob(job *CronJob) error {
	entryID, err := s.Cron.AddFunc(job.Schedule, func() {
		var err error
		switch job.Method {
		case "GET":
			_, err = http.Get("http://localhost:8080" + job.Endpoint)
		case "POST":
			_, err = http.Post("http://localhost:8080"+job.Endpoint, "application/json", nil)
		}
		if err != nil {
			log.Printf("cron failed to make request to %s: %v", job.Endpoint, err)
			return
		}
	})
	if err != nil {
		return err
	}
	job.Enabled = true
	job.EntryID = entryID
	return nil
}

func (s *Scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	job, ok := s.Jobs[id]
	s.mu.Unlock()
	if !ok {
		return errors.New("Job not found")
	}
	s.Cron.Remove(job.EntryID)

	s.mu.Lock()
	job.Enabled = false
	s.mu.Unlock()

	err := s.SaveCronJobs()
	if err != nil {
		return errors.New("Failed to save cron jobs")
	}
	return nil
}

func (s *Scheduler) DeleteJob(id string) error {
	s.mu.Lock()
	job, ok := s.Jobs[id]
	s.mu.Unlock()
	if !ok {
		return errors.New("Job not found")
	}

	s.Cron.Remove(job.EntryID)

	s.mu.Lock()
	delete(s.Jobs, id)
	s.mu.Unlock()

	err := s.SaveCronJobs()
	if err != nil {
		return errors.New("Failed to save cron jobs")
	}
	return nil
}

func (s *Scheduler) UpdateJob(id string, job *CronJob) error {
	err := s.RemoveJob(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.Jobs[id] = job
	s.mu.Unlock()

	err = s.SaveCronJobs()
	if err != nil {
		return err
	}

	err = s.ScheduleJob(job)
	if err != nil {
		return err
	}

	return nil
}
