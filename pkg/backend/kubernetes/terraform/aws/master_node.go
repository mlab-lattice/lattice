package aws

type MasterNodeEtcdVolumeAttachment struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ClusterID  string `json:"cluster_id"`
	Name       string `json:"name"`
	InstanceID string `json:"instance_id"`
	DeviceName string `json:"device_name"`
}

func NewMasterNodeEtcdVolumeAttachment(
	moduleRoot,
	region,
	clusterID,
	name,
	instanceID,
	deviceName string,
) *MasterNodeEtcdVolumeAttachment {
	return &MasterNodeEtcdVolumeAttachment{
		Source: moduleRoot + modulePathMasterNodeEtcdVolumeAttachment,

		Region: region,

		ClusterID:  clusterID,
		Name:       name,
		InstanceID: instanceID,
		DeviceName: deviceName,
	}
}

type MasterNodeDNS struct {
	Source string `json:"source"`

	Region string `json:"region"`

	Name                 string `json:"name"`
	Route53PrivateZoneID string `json:"route53_private_zone_id"`
	InstancePrivateIP    string `json:"instance_private_ip"`
}

func NewMasterNodeDNS(
	moduleRoot,
	region,
	name,
	route53PrivateZoneID,
	instancePrivateIP string,
) *MasterNodeDNS {
	return &MasterNodeDNS{
		Source: moduleRoot + modulePathMasterNodeDNS,

		Region: region,

		Name:                 name,
		Route53PrivateZoneID: route53PrivateZoneID,
		InstancePrivateIP:    instancePrivateIP,
	}
}
