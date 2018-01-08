package aws

type ApplicationLoadBalancer struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ClusterID string `json:"cluster_id"`
	SystemID  string `json:"system_id"`
	VPCID     string `json:"vpc_id"`
	SubnetIDs string `json:"subnet_ids"`

	Name                    string `json:"name"`
	AutoscalingGroupName    string `json:"autoscaling_group_name"`
	NodePoolSecurityGroupID string `json:"node_pool_security_group_id"`

	Ports map[int32]int32 `json:"ports"`
}

func NewApplicationLoadBalancerModule(
	moduleRoot, region, clusterID, systemID, vpcID, subnetIDs,
	name, autoscalingGroupName, nodePoolSecurityGroupID string,
	ports map[int32]int32,
) *ApplicationLoadBalancer {
	return &ApplicationLoadBalancer{
		Source: moduleRoot + modulePathApplicationLoadBalancer,

		Region: region,

		ClusterID: clusterID,
		SystemID:  systemID,
		VPCID:     vpcID,
		SubnetIDs: subnetIDs,

		Name:                    name,
		AutoscalingGroupName:    autoscalingGroupName,
		NodePoolSecurityGroupID: nodePoolSecurityGroupID,

		Ports: ports,
	}
}
