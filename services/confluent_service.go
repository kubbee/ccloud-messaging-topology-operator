package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"

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
		} else {
			logger.Error(errors.New("environment not match"), ERROR_ENVIRONMENT_NOT_MATCH)
		}
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

		clusters := []util.Cluster{}

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

func getOutput(outs []byte) (string, bool) {
	if len(outs) > 0 {
		return string(outs), true
	}
	return "", false
}
