package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"regexp"

	"github.com/go-logr/logr"

	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

func GetEnvironments(logger *logr.Logger) (*[]util.Environment, error) {
	logger.Info("Start func GetEnvironments")
	cmd := exec.Command("/bin/confluent", "environment", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "Error to retrive clonfluent cloud environments")
		return &[]util.Environment{}, err
	} else {
		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		environments := []util.Environment{}
		json.Unmarshal([]byte(message), &environments)

		logger.Info("Existents Environments: " + message)

		return &environments, nil
	}
}

func SelectEnvironment(envs []util.Environment, logger *logr.Logger, kr *messagesv1alpha1.KafkaReference) (string, bool) {
	var envId string
	for i := 0; i < len(envs); i++ {
		if envs[i].Name == kr.Spec.Environment {
			envId = envs[i].Id
			logger.Info("func::SelectEnvironment, the id of environment is --> " + envId)
			break
		}
	}

	if envId == "" {
		logger.Error(errors.New("environment not match"), ERROR_ENVIRONMENT_NOT_MATCH)
	}

	return envId, SetEnvironment(envId, logger)
}

func SetEnvironment(environmentId string, logger *logr.Logger) bool {
	cmd := exec.Command("/bin/confluent", "environment", "use", environmentId)

	if err := cmd.Run(); err != nil {
		logger.Error(err, ERROR_SETENVIRONMENT)
		return false
	}

	return true
}

func GetKafkaCluster(logger *logr.Logger, kr *messagesv1alpha1.KafkaReference) (string, error) {
	cmd := exec.Command("/bin/confluent", "kafka", "cluster", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, ERROR_GET_KAFKA_CLUSTER)
		return string(""), err
	} else {
		var clusterId string

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		clusters := []util.ClusterKafka{}

		json.Unmarshal([]byte(message), &clusters)

		for i := 0; i < len(clusters); i++ {
			if kr.Spec.ClusterName == clusters[i].Name {
				clusterId = clusters[i].Id
				break
			}
		}
		logger.Info("Kafka cluster was selected")
		return clusterId, nil
	}
}

func CreateNewTopic(topic *util.NewTopic, logger *logr.Logger) (*util.TopicReference, error) {

	logger.Info("Strating to Creation name")
	topicName := topicNameGenerator(topic.Tenant, topic.Namespace, topic.Topic)

	if isOk, err := existsTopicName(topic.ClusterReference.ClusterId, topicName); err != nil {
		logger.Error(errors.New("error to create topic"), ERROR_NEW_TOPIC)
		return &util.TopicReference{}, err
	} else {

		topicReference := &util.TopicReference{}
		topicReference.Topic = topicName

		if !isOk {
			cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", topicName, "--partitions", topic.Partitions, "--cluster", topic.ClusterReference.ClusterId)

			if err := cmd.Run(); err != nil {
				logger.Error(err, "Confluent::NewTopic:: Error to run the command to create kafka topic")
				return &util.TopicReference{}, err
			} else {
				if reference, rError := newTopicConnection(topic.ClusterReference, topic.Namespace, topicReference, logger); rError != nil {
					logger.Info("Error:: error to get connection info")
					return topicReference, nil
				} else {
					return reference, nil
				}
			}
		} else {
			if reference, rError := newTopicConnection(topic.ClusterReference, topic.Namespace, topicReference, logger); rError != nil {
				logger.Info("Error:: error to get connection info")
				return topicReference, nil
			} else {
				return reference, nil
			}
		}
	}
}

func newTopicConnection(clusterReference util.ClusterReference, namespace string, topicReference *util.TopicReference, logger *logr.Logger) (*util.TopicReference, error) {

	if schemaRegistry, srError := GetSechemaRegistry(clusterReference.EnvironmentId, logger); srError != nil {
		logger.Error(srError, "error to obtain GetSechemaRegistry")
		return topicReference, srError
	} else {

		registry := util.SchemaRegistry{
			Name:                schemaRegistry.Name,
			ClusterId:           schemaRegistry.ClusterId,
			EndpointUrl:         schemaRegistry.EndpointUrl,
			UsedSchemas:         schemaRegistry.UsedSchemas,
			AvailableSchemas:    schemaRegistry.AvailableSchemas,
			GlobalCompatibility: schemaRegistry.GlobalCompatibility,
			Mode:                schemaRegistry.Mode,
			ServiceProvider:     schemaRegistry.ServiceProvider,
		}

		topicReference.SchemaRegistry = registry

		if registryApiKey, err := NewSchemaRegistryApiKey(registry.ClusterId, namespace, logger); err != nil {
			return topicReference, err
		} else {

			schemaRegistryApiKey := util.SchemaRegistryApiKey{
				Api:    registryApiKey.Api,
				Secret: registryApiKey.Secret,
			}

			topicReference.SchemaRegistryApiKey = schemaRegistryApiKey
		}
	}

	if clusterKafka, ckError := GetKafkaClusterSettings(clusterReference, logger); ckError != nil {
		logger.Error(ckError, "error to obtain GetKafkaClusterSettings")
		return topicReference, ckError
	} else {

		clusterKafkaSettings := util.ClusterKafka{
			Id:           clusterKafka.Id,
			Name:         clusterKafka.Name,
			Type:         clusterKafka.Type,
			Ingress:      clusterKafka.Ingress,
			Egress:       clusterKafka.Egress,
			Storage:      clusterKafka.Storage,
			Provider:     clusterKafka.Provider,
			Region:       clusterKafka.Region,
			Availability: clusterKafka.Availability,
			Status:       clusterKafka.Status,
			Endpoint:     clusterKafka.Endpoint,
			RestEndpoint: clusterKafka.RestEndpoint,
		}

		topicReference.ClusterKafka = clusterKafkaSettings

		if kafkaApiKey, kakError := NewKafkaApiKey(clusterReference.ClusterId, namespace, logger); kakError != nil {
			return topicReference, kakError
		} else {

			clusterkafkaApiKey := util.KafkaApiKey{
				Api:    kafkaApiKey.Api,
				Secret: kafkaApiKey.Secret,
			}

			topicReference.KafkaApiKey = clusterkafkaApiKey
		}
	}

	return topicReference, nil
}

func NewKafkaApiKey(clusterId string, description string, logger *logr.Logger) (*util.KafkaApiKey, error) {
	logger.Info("Creating Api-Key for Kafka Connection")

	cmd := exec.Command("/bin/confluent", "api-key", "create", "--resource", clusterId, "--description", description, "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "Confluent::NewKafkaApiKey:: Error to create the kafka api-key for the application")
		return &util.KafkaApiKey{}, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		kafkaApiKey := &util.KafkaApiKey{}
		json.Unmarshal([]byte(message), kafkaApiKey)

		if kafkaApiKey.Api != "" && kafkaApiKey.Secret != "" {
			return kafkaApiKey, nil
		} else {
			return &util.KafkaApiKey{}, errors.New("error to parse the api-key")
		}
	}
}

func NewSchemaRegistryApiKey(schemaRegistryId string, description string, logger *logr.Logger) (*util.KafkaApiKey, error) {
	logger.Info("Creating Api-Key for Kafka Connection")

	cmd := exec.Command("/bin/confluent", "api-key", "create", "--resource", schemaRegistryId, "--description", description, "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, "Confluent::NewSchemaRegistryApiKey:: Error to create the schema-registry api-key for the application")
		return &util.KafkaApiKey{}, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		kafkaApiKey := &util.KafkaApiKey{}
		json.Unmarshal([]byte(message), kafkaApiKey)

		if kafkaApiKey.Api != "" && kafkaApiKey.Secret != "" {
			return kafkaApiKey, nil
		} else {
			return &util.KafkaApiKey{}, errors.New("error to parse the api-key")
		}
	}
}

func GetSechemaRegistry(environmentId string, logger *logr.Logger) (*util.SchemaRegistry, error) {
	logger.Info("func GetSechemaRegistry started")
	cmd := exec.Command("/bin/confluent", "schema-registry", "cluster", "describe", "--environment", environmentId, "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		logger.Error(err, ERROR_GET_SCHEMA_REGISTRY)
		return &util.SchemaRegistry{}, err
	} else {
		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		registry := &util.SchemaRegistry{}
		json.Unmarshal([]byte(message), registry)

		if registry.EndpointUrl != "" {
			return registry, nil
		} else {
			return &util.SchemaRegistry{}, errors.New("error to parse schema registry settings")
		}
	}
}

func GetKafkaClusterSettings(clusterReference util.ClusterReference, logger *logr.Logger) (*util.ClusterKafka, error) {
	cmdUse := exec.Command("/bin/confluent", "kafka", "cluster", "use", clusterReference.ClusterId)

	if err := cmdUse.Run(); err != nil {
		logger.Error(err, "error to use Kafka Cluster")
		return &util.ClusterKafka{}, err
	} else {

		cmd := exec.Command("/bin/confluent", "kafka", "cluster", "describe", "--environment", clusterReference.EnvironmentId, "--output", "json")

		cmdOutput := &bytes.Buffer{}
		cmd.Stdout = cmdOutput

		if err := cmd.Run(); err != nil {
			logger.Error(err, "error to get the kafka cluster reference")
			return &util.ClusterKafka{}, err
		} else {
			output := cmdOutput.Bytes()
			message, _ := getOutput(output)

			cluster := &util.ClusterKafka{}
			json.Unmarshal([]byte(message), cluster)

			if cluster.Id != "" {
				return cluster, nil
			} else {
				return &util.ClusterKafka{}, errors.New("error to parse kafka cluster settings")
			}
		}
	}
}

/*
 * This function will generates the name of the kafka topic, if in the tenant name you pass something
 * different of blank and default, the name will be use the tenant do compose the toppic name;
 */
func topicNameGenerator(tenant string, namespace string, topicName string) string {
	separator := "-"
	x := string("")
	if tenant != x || tenant != string("default") {
		return tenant + separator + namespace + separator + topicName
	}
	return namespace + separator + topicName
}

/*
 * This function get the output console and converts to string
 */
func getOutput(outs []byte) (string, bool) {
	if len(outs) > 0 {
		return string(outs), true
	}
	return "", false
}

/*
 * This function verify the existence of a topic create with the name
 */
func existsTopicName(clusterId string, topicName string) (bool, error) {

	cmd := exec.Command("/bin/confluent", "kafka", "topic", "list", "--cluster", clusterId)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		//c.Log.Error(err, "Falied to verify the existence of topic")
		return false, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := getOutput(output)

		return regexp.MatchString("\\b"+topicName+"\\b", message)
	}
}
