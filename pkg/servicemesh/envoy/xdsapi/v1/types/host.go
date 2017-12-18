package types

type StaticHost struct {
	URL string `json:"url"`
}

type SDSHost struct {
	IPAddress string `json:"ip_address"`
	Port      int32  `json:"port"`
	// TODO: reexamine other fields
}
