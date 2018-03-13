package block

import (
	"encoding/json"
)

type Reference struct {
	Path string `json:"path"`
}

func (r *Reference) UnmarshalJSON(data []byte) error {
	re := &referenceEncoder{}
	if err := json.Unmarshal(data, &re); err != nil {
		return err
	}

	r.Path = re.Reference
	return nil
}

func (r *Reference) MarshalJSON() ([]byte, error) {
	re := &referenceEncoder{
		Reference: r.Path,
	}

	return json.Marshal(&re)
}

type referenceEncoder struct {
	Reference string `json:"reference"`
}
