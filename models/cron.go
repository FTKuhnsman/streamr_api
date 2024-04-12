package models

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"streamr_api/common"

	"github.com/robfig/cron/v3"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const cronJobFile = "cron_jobs.json"

type Scheduler struct {
	Jobs map[string]*CronJob `json:"jobs"`
	Cron *cron.Cron          `json:"-"`
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
		Jobs: make(map[string]*CronJob),
		Cron: cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger))),
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
	_, err := os.Stat(cronJobFile)
	if os.IsNotExist(err) {
		return nil // Return an empty list if the file does not exist
	}

	// Read the file
	bytes, err := os.ReadFile(cronJobFile)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the jobs slice
	err = json.Unmarshal(bytes, &jobs)
	if err != nil {
		return err
	}

	s.Jobs = jobs

	// Schedule the jobs if enabled
	for _, job := range s.Jobs {
		if job.Enabled {
			err = s.ScheduleJob(job)
			if err != nil {
				log.Printf("Failed to schedule job %s: %v", job.Name, err)
			}
		}
	}
	return nil
}

// SaveCronJobs saves the provided cron jobs into a JSON file.
func (s *Scheduler) SaveCronJobs() error {
	bytes, err := json.Marshal(s.Jobs)
	if err != nil {
		return err
	}

	// Write to the file
	return os.WriteFile(cronJobFile, bytes, 0644)
}

func (s *Scheduler) CreateCronJob(job *CronJob) error {
	// Add the job to the scheduler
	randID, err := common.GenerateRandomHexString(16)
	if err != nil {
		return errors.New("Failed to generate random ID")
	}

	s.Jobs[randID] = job

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

func (s *Scheduler) GetCronJobs() map[string]*CronJob {
	return s.Jobs
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
	job.EntryID = entryID
	return nil
}

func (s *Scheduler) RemoveJob(id string) error {
	job, ok := s.Jobs[id]
	if !ok {
		return errors.New("Job not found")
	}
	s.Cron.Remove(job.EntryID)
	job.Enabled = false
	err := s.SaveCronJobs()
	if err != nil {
		return errors.New("Failed to save cron jobs")
	}
	return nil
}

func (s *Scheduler) DeleteJob(id string) error {
	job, ok := s.Jobs[id]
	if !ok {
		return errors.New("Job not found")
	}
	s.Cron.Remove(job.EntryID)
	delete(s.Jobs, id)
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

	s.Jobs[id] = job

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
