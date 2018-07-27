package kubernetes

const (
	UserMainContainerName      = "lattice-user-main"
	UserSidecarContainerPrefix = "lattice-user-sidecar-"
)

func UserSidecarContainerName(sidecar string) string {
	return UserSidecarContainerPrefix + sidecar
}
