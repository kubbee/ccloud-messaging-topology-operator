package services

import (
	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/ccloud-messaging-topology-operator/api/v1alpha1"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

func GetClusterReference(kr *messagesv1alpha1.KafkaReference, logger *logr.Logger) (*util.ClusterReference, error) {

	if environments, err := GetEnvironments(logger); err != nil {
		logger.Error(err, "error to get environmentId")
		return &util.ClusterReference{}, err
	} else {
		environmentId, selected := SelectEnvironment(*environments, logger, kr)
		//
		if selected {
			clusterId, err := GetKafkaCluster(logger, kr)
			//
			if err != nil {
				logger.Error(err, "error to get clusterId")
				return &util.ClusterReference{}, err
			}

			//
			return &util.ClusterReference{
				ClusterId:     clusterId,
				EnvironmentId: environmentId,
			}, nil
		}
		return &util.ClusterReference{}, nil
	}
}
