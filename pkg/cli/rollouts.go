package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

var rolloutHeaders = []string{"ID", "State"}

func getRolloutValues(rollout *types.SystemRollout) []string {
	return []string{
		string(rollout.ID),
		string(rollout.State),
	}
}

func ShowRollout(rollout *types.SystemRollout, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getRolloutValues(rollout)
		ShowResource(rolloutHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(rollout)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowRollouts(rollouts []types.SystemRollout, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(rollouts))
		for i, b := range rollouts {
			values[i] = getRolloutValues(&b)
		}
		ListResources(rolloutHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(rollouts)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
