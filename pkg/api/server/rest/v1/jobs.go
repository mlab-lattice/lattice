package v1

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"

	"github.com/gin-gonic/gin"
)

const jobIdentifier = "job_id"

var (
	jobsPath                   = fmt.Sprintf(v1rest.JobsPathFormat, systemIdentifierPathComponent)
	jobIdentifierPathComponent = fmt.Sprintf(":%v", jobIdentifier)
	jobPath                    = fmt.Sprintf(v1rest.JobPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)
	jobLogPath                 = fmt.Sprintf(v1rest.JobLogsPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)
)

func (api *LatticeAPI) setupJobsEndpoints() {

	// run-job
	api.router.POST(jobsPath, api.handleRunJob)

	// list-jobs
	api.router.GET(jobsPath, api.handleListJobs)

	// get-job
	api.router.GET(jobPath, api.handleGetJob)

	// get-job-logs
	api.router.GET(jobLogPath, api.handleGetJobLogs)

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

	job, err := api.backend.Systems().Jobs(systemID).Run(req.Path, req.Command, req.Environment)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidPath:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
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
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
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
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidJobID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, job)
}

// handleGetJobLogs handler for get-job-logs
// @ID get-job-logs
// @Summary Get job logs
// @Description Retrieves/Streams logs for job
// @Router /systems/{system}/jobs/{id}/logs  [get]
// @Security ApiKeyAuth
// @Tags jobs
// @Param system path string true "System ID"
// @Param id path string true "Job ID"
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
func (api *LatticeAPI) handleGetJobLogs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))

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

	log, err := api.backend.Systems().Jobs(systemID).Logs(jobID, sidecar, logOptions)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidJobID, v1.ErrorCodeInvalidInstance:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
