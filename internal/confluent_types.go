package internal

type KafkaReferenceSecret struct {
	ClusterId     string `json:"clusterId"`
	EnvironmentId string `json:"environmentId"`
	Tenant        string `json:"tenant"`
}

type NewTopic struct {
	Tenant     string `json:"tenant"`
	Namespace  string `json:"namespace"`
	Topic      string `json:"topic"`
	Partitions string `json:"partitions"`
}

type ClusterKafka struct {
	Id           string `json:"id"`            //"id": "lkc-57qx6n",
	Name         string `json:"name"`          //"name": "demo-kafka",
	Type         string `json:"type"`          //"type": "BASIC",
	Ingress      int16  `json:"ingress"`       //"ingress": 100,
	Egress       int16  `json:"egress"`        //"egress": 100,
	Storage      string `json:"storage"`       //"storage": "5 TB",
	Provider     string `json:"provider"`      //"provider": "aws",
	Region       string `json:"region"`        //"region": "us-east-2",
	Availability string `json:"availability"`  //"availability": "single-zone",
	Status       string `json:"status"`        //"status": "UP",
	Endpoint     string `json:"endpoint"`      //"endpoint": "SASL_SSL://pkc-ymrq7.us-east-2.aws.confluent.cloud:9092",
	RestEndpoint string `json:"rest_endpoint"` //"rest_endpoint": "https://pkc-ymrq7.us-east-2.aws.confluent.cloud:443"
}

type ConfigMapKafka struct {
	TopicName               string
	SchemaRegistryURL       string
	SchemaRegistryApiKey    string
	SchemaRegistryApiSecret string
	KafkaURL                string
	KafkaApiKey             string
	KafkaApiSecret          string
}
