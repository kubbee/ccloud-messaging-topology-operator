package internal

type ClusterReference struct {
	ClusterId     string `json:"clusterId"`
	EnvironmentId string `json:"environmentId"`
}

type Environment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Cluster struct {
	Availability string `json:"availability"` //"availability": "single-zone",
	Id           string `json:"id"`           //"id": "lkc-nvywwv",
	Name         string `json:"name"`         //"name": "kubber2",
	Provider     string `json:"provider"`     //"provider": "aws",
	Region       string `json:"region"`       //"region": "us-east-2",
	Status       string `json:"status"`       //"status": "UP",
	Type         string `json:"type"`         //"type": "BASIC"
}

type KafkaReferenceSecret struct {
	ClusterId     string `json:"clusterId"`
	EnvironmentId string `json:"environmentId"`
	Tenant        string `json:"tenant"`
}
