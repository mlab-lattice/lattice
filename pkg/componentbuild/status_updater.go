package componentbuild

type Phase string

const (
	PhasePullingGitRepository = "pulling git repository"
	PhaseBuildingDockerImage  = "building docker image"
	PhasePushingDockerImage   = "pushing docker image"
)

type PhaseState string

const (
	PhaseStateInProgress = "in progress"
	PhaseStateFailed     = "failed"
)

type Progress struct {
	Phase Phase      `json:"phase"`
	State PhaseState `json:"state"`
}

type ProgressUpdater interface {
	UpdateProgress(Progress) error
}
