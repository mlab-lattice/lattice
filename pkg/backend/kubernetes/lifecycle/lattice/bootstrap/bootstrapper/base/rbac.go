package base

var (
	readVerbs                      = []string{"get", "watch", "list"}
	readAndCreateVerbs             = []string{"get", "watch", "list", "create"}
	readAndDeleteVerbs             = []string{"get", "watch", "list", "delete"}
	readAndUpdateVerbs             = []string{"get", "watch", "list", "update"}
	readCreateAndDeleteVerbs       = []string{"get", "watch", "list", "create", "delete"}
	readCreateUpdateAndDeleteVerbs = []string{"get", "watch", "list", "create", "update", "delete"}
)
