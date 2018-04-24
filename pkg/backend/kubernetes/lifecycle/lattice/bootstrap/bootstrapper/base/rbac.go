package base

var (
	ReadVerbs                      = []string{"get", "watch", "list"}
	ReadAndCreateVerbs             = []string{"get", "watch", "list", "create"}
	ReadAndDeleteVerbs             = []string{"get", "watch", "list", "delete"}
	ReadAndUpdateVerbs             = []string{"get", "watch", "list", "update"}
	ReadCreateAndDeleteVerbs       = []string{"get", "watch", "list", "create", "delete"}
	ReadCreateUpdateAndDeleteVerbs = []string{"get", "watch", "list", "create", "update", "delete"}
)
