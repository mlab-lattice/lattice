package v1

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"

	"github.com/gin-gonic/gin"
)

const (
	jobIdentifier    = "job_id"
	jobRunIdentifier = "job_run_id"
)

var (
	jobsPath                      = fmt.Sprintf(v1rest.JobsPathFormat, systemIdentifierPathComponent)
	jobIdentifierPathComponent    = fmt.Sprintf(":%v", jobIdentifier)
	jobPath                       = fmt.Sprintf(v1rest.JobPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)
	jobRunsPath                   = fmt.Sprintf(v1rest.JobRunsPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)
	jobRunIdentifierPathComponent = fmt.Sprintf(":%v", jobRunIdentifier)
	jobRunPath                    = fmt.Sprintf(v1rest.JobRunPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent, jobRunIdentifierPathComponent)
	jobRunLogPath                 = fmt.Sprintf(v1rest.JobRunLogsPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent, jobRunIdentifierPathComponent)
)

func (api *LatticeAPI) setupJobsEndpoints() {

	// run-job
	api.router.POST(jobsPath, api.handleRunJob)

	// list-jobs
	api.router.GET(jobsPath, api.handleListJobs)

	// get-job
	api.router.GET(jobPath, api.handleGetJob)

	// list-job-runs
	api.router.GET(jobRunsPath, api.handleListJobRuns)

	// get-job-run
	api.router.GET(jobRunPath, api.handleGetJobRun)

	// get-job-run-logs
	api.router.GET(jobRunLogPath, api.handleGetJobRunLogs)
}

// handleRunJob handler for run-job
// @ID run-job
// @Summary Run job
// @Description Runs a new job
// @Router /systems/{system}/jobs [post]
// @Security ApiKeyAuth
// @Tags jobs
// @Param system path string true "System ID"
// @Param jobRequest body rest.RunJobRequest true "Create build"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Job
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleRunJob(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	var req v1rest.RunJobRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	job, err := api.backend.Systems().Jobs(systemID).Run(req.Path, req.Command, req.Environment, req.NumRetries)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, job)
}

// handleListJobs handler for list-jobs
// @ID list-jobs
// @Summary Lists jobs
// @Description Lists all jobs
// @Router /systems/{system}/jobs [get]
// @Security ApiKeyAuth
// @Tags jobs
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Job
func (api *LatticeAPI) handleListJobs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	jobs, err := api.backend.Systems().Jobs(systemID).List()
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// handleGetJob handler for get-job
// @ID get-job
// @Summary Get job
// @Description Gets the job object
// @Router /systems/{system}/jobs/{id} [get]
// @Security ApiKeyAuth
// @Tags jobs
// @Param system path string true "System ID"
// @Param id path string true "Job ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Job
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetJob(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))

	job, err := api.backend.Systems().Jobs(systemID).Get(jobID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, job)
}

// handleListJobRuns handler for list-job-runs
// @ID list-job-runs
// @Summary Lists job runs
// @Description Lists all job runs
// @Router /systems/{system}/jobs/{job} [get]
// @Security ApiKeyAuth
// @Tags jobRuns
// @Param system path string true "System ID"
// @Param job path string true "Job ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.JobRun
func (api *LatticeAPI) handleListJobRuns(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))

	runs, err := api.backend.Systems().Jobs(systemID).Runs(jobID).List()
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, runs)
}

// handleGetJobRun handler for get-job-run
// @ID get-job-run
// @Summary Get job run
// @Description Gets the job run object
// @Router /systems/{system}/jobs/{job}/run/{id} [get]
// @Security ApiKeyAuth
// @Tags jobsRuns
// @Param system path string true "System ID"
// @Param job path string true "Job ID"
// @Param id path string true "Job Run ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.JobRun
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetJobRun(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))
	runID := v1.JobRunID(c.Param(jobRunIdentifier))

	run, err := api.backend.Systems().Jobs(systemID).Runs(jobID).Get(runID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, run)
}

// handleGetJobLogs handler for get-job-run-logs
// @ID get-job-run-logs
// @Summary Get job run logs
// @Description Retrieves/Streams logs for job
// @Router /systems/{system}/jobs/{job}/runs/{id}logs  [get]
// @Security ApiKeyAuth
// @Tags jobRuns
// @Param system path string true "System ID"
// @Param job path string true "Job ID"
// @Param id path string true "Job Run ID"
// @Param sidecar query string false "Sidecar"
// @Param follow query string bool "Follow"
// @Param previous query boolean false "Previous"
// @Param timestamps query boolean false "Timestamps"
// @Param tail query integer false "tail"
// @Param since query string false "Since"
// @Param sinceTime query string false "Since Time"
// @Accept  json
// @Produce  json
// @Success 200 {string} string "log stream"
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetJobRunLogs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))
	runID := v1.JobRunID(c.Param(jobRunIdentifier))

	sidecarQuery, sidecarSet := c.GetQuery("sidecar")
	var sidecar *string
	if sidecarSet {
		sidecar = &sidecarQuery
	}

	logOptions, err := requestedLogOptions(c)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	log, err := api.backend.Systems().Jobs(systemID).Runs(jobID).Logs(runID, sidecar, logOptions)
	if err != nil {
		handleError(c, err)
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
