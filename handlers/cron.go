package handlers

import (
	"net/http"

	"streamr_api/models"

	"github.com/gin-gonic/gin"
)

// Use this command to update swagger docs:
// swag init --parseDependency  --parseInternal --parseDepth 1  -g main.go

// CreateCronJob godoc
// @Summary      Create a new cron job
// @Description  Adds a new cron job to the scheduler and saves it to the storage.
// @Tags         CronJob
// @Accept       json
// @Produce      json
// @Param        cronJob  body      models.CronJob  true  "Create Cron Job"
// @Success      200  {object}  models.CronJob
// @Router       /cronjobs/create [post]
func CreateCronJob(s *models.Scheduler) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var job models.CronJob
		if err := c.BindJSON(&job); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err := s.CreateCronJob(&job)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, job)
	}
	return gin.HandlerFunc(fn)
}

// GetCronJobs godoc
// @Summary Get all cron jobs
// @Description Retrieves a list of all scheduled cron jobs.
// @Tags CronJob
// @Produce json
// @Success 200 {array} models.Scheduler.Jobs
// @Router /cronjobs [get]
func GetCronJobs(s *models.Scheduler) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		jobs := s.GetCronJobs()
		c.JSON(http.StatusOK, jobs)
	}
	return gin.HandlerFunc(fn)
}

// DisableCronJob godoc
// @Summary      Disable a cron job by ID
// @Description  Disables a cron job in the scheduler and saves it to the storage.
// @Tags         CronJob
// @Produce      json
// @Param        id  path      string  true  "Cron Job ID"
// @Success      200  {object}  models.CronJob
// @Router       /cronjobs/disable/{id} [post]
func DisableCronJob(s *models.Scheduler) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")
		err := s.RemoveJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, s.Jobs[id])
	}
	return gin.HandlerFunc(fn)
}

// EnableCronJob godoc
// @Summary      Enable a cron job by ID
// @Description  Enables a cron job in the scheduler and saves it to the storage.
// @Tags         CronJob
// @Produce      json
// @Param        id  path      string  true  "Cron Job ID"
// @Success      200  {object}  models.CronJob
// @Router       /cronjobs/enable/{id} [post]
func EnableCronJob(s *models.Scheduler) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")

		err := s.ScheduleJob(s.Jobs[id])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, s.Jobs[id])
	}
	return gin.HandlerFunc(fn)
}

// DeleteCronJob godoc
// @Summary      Delete a cron job by ID
// @Description  Deletes a cron job in the scheduler and saves the change to the storage.
// @Tags         CronJob
// @Produce      json
// @Param        id  path      string  true  "Cron Job ID"
// @Success      200  {object}  models.CronJob
// @Router       /cronjobs/delete/{id} [post]
func DeleteCronJob(s *models.Scheduler) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")
		err := s.DeleteJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, s.Jobs[id])
	}
	return gin.HandlerFunc(fn)
}
