package services

import (
	"errors"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

var message string = ""

func BuildTopic(topic *util.NewTopic, environmentId string, clusterId string, logger *logr.Logger) (*string, error) {
	if logger != nil {
		logger.Info("start::BuildTopic")
	}

	if _, err := setEnvironment(environmentId); err != nil {
		if logger != nil {
			logger.Error(err, "erro to selecet the environment")
		}

		return &message, err
	}

	return createNewTopic(topic.Tenant, topic.Topic, topic.Namespace, topic.Partitions, clusterId)
}

func RetrieveTopic(topic *util.ExistentTopic, environmentId string, clusterId string, logger *logr.Logger) (*string, error) {
	if logger != nil {
		logger.Info("start::RetriveTopic")
	}

	if _, err := setEnvironment(environmentId); err != nil {
		if logger != nil {
			logger.Error(err, "erro to selecet the environment")
		}
		return &message, err
	}

	name := topicNameGenerator(topic.Tenant, topic.Domain, topic.Topic)

	if ok, err := existsTopicName(clusterId, name); err != nil {
		return &name, err
	} else {
		if ok {
			return &name, err
		} else {
			return &name, errors.New("Topic not exists yet on the cluster id=" + clusterId)
		}
	}

}
