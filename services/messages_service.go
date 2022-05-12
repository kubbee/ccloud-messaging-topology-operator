package services

const (
	ERROR_ENVIRONMENT_NOT_MATCH = "The environment in KafkaCluster Resource doesn't match with the environments on the Confluent Cloud"
	ERROR_SETENVIRONMENT        = "Error to setup the environment"
	ERROR_GET_KAFKA_CLUSTER     = "Error to retrive confluent cloud kafka clusters"
	ERROR_GET_SCHEMA_REGISTRY   = "Error to retrive schema registry"
	ERROR_NEW_TOPIC             = "Error to create the topic on Confluent Cloud"
)
