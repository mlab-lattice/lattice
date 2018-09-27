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

// swagger:operation POST /systems/{system}/jobs jobs RunJob
//
// Runs jobs
//
// This will run a new job with the provided command and environment
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - in: body
//         schema:
//           "$ref": "#/definitions/JobRequest"
//     responses:
//         default:
//           description: Job object
//           schema:
//             "$ref": "#/definitions/Job"

// handleRunJob handler for RunJob
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
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeInvalidPath:
			c.JSON(http.StatusNotFound, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, job)

}

// swagger:operation GET /systems/{system}/jobs jobs ListJobs
//
// Lists jobs
//
// Lists jobs for a system
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: job list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Job"
//

// handleListJobs handler for ListJobs
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

// swagger:operation GET /systems/{system}/jobs/{jobId} jobs GetJob
//
// Get job
//
// Gets jobs details
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Job ID
//         in: path
//         name: jobId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Job Object
//           schema:
//             "$ref": "#/definitions/Job"
//
// handleGetJob handler for GetJob
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

// swagger:operation GET /systems/{system}/jobs/{jobId}/logs jobs GetJobLogs
//
// Get job logs
//
// Returns a job log stream
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Job ID
//         in: path
//         name: jobId
//         required: true
//         type: string
//       - description: Sidecar
//         in: query
//         name: sidecar
//         required: false
//         type: string
//       - description: Follow
//         in: query
//         name: follow
//         required: false
//         type: boolean
//       - description: Previous
//         in: query
//         name: previous
//         required: false
//         type: boolean
//       - description: Timestamps
//         in: query
//         name: timestamps
//         required: false
//         type: boolean
//       - description: Tail
//         in: query
//         name: tail
//         required: false
//         type: int
//       - description: Since
//         in: query
//         name: since
//         required: false
//         type: string
//       - description: Since Time
//         in: query
//         name: sinceTime
//         required: false
//         type: string
//     responses:
//         '200':
//           description: log stream
//           schema:
//             type: string

// handleGetJobLogs handler for GetJobLogs
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
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidJobID:
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
