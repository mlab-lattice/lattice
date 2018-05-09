package terraform

import (
	"encoding/json"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type ApplicationLoadBalancer struct {
	Source string `json:"source"`

	Region string `json:"region"`

	LatticeID v1.LatticeID `json:"lattice_id"`
	SystemID  v1.SystemID  `json:"system_id"`
	VPCID     string       `json:"vpc_id"`
	SubnetIDs []string     `json:"subnet_ids"`

	Name                             string            `json:"name"`
	AutoscalingGroupSecurityGroupIDs map[string]string `json:"autoscaling_group_security_group_ids"`
	Ports                            map[int32]int32   `json:"ports"`
}

func (np *ApplicationLoadBalancer) MarshalJSON() ([]byte, error) {
	encoder := applicationLoadBalancerEncoder{
		Source: np.Source,

		Region: np.Region,

		LatticeID: string(np.LatticeID),
		SystemID:  string(np.SystemID),
		VPCID:     np.VPCID,
		SubnetIDs: strings.Join(np.SubnetIDs, ","),

		Name: np.Name,
		AutoscalingGroupSecurityGroupIDs: np.AutoscalingGroupSecurityGroupIDs,
		Ports: np.Ports,
	}
	return json.Marshal(&encoder)
}

type applicationLoadBalancerEncoder struct {
	Source string `json:"source"`

	Region string `json:"region"`

	LatticeID string `json:"lattice_id"`
	SystemID  string `json:"system_id"`
	VPCID     string `json:"vpc_id"`
	SubnetIDs string `json:"subnet_ids"`

	Name                             string            `json:"name"`
	AutoscalingGroupSecurityGroupIDs map[string]string `json:"autoscaling_group_security_group_ids"`
	Ports                            map[int32]int32   `json:"ports"`
}
