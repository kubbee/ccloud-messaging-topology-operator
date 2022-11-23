package services

import (
	"errors"

	"github.com/go-logr/logr"
)

func BuildKafkaReference(environmentId string, clusterName string, logger *logr.Logger) (string, error) {
	logger.Info("start::BuildKafkaReference")

	success, err := setEnvironment(environmentId)

	if err != nil {
		logger.Error(err, "error to select the environment")
		return "", err
	}

	if success {
		return getKafkaCluster(clusterName)
	} else {
		e := errors.New("there was an error to select the environment")
		logger.Error(e, "BuildKafkaReference::"+e.Error())
		return "", e
	}
}
