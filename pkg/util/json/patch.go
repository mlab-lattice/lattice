package json

// http://jsonpatch.com/
const (
	PatchOpAdd     = "add"
	PatchOpRemove  = "remove"
	PatchOpReplace = "replace"
)

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
