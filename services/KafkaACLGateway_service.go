package services

import (
	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

// BuildACL call the service to create KafkaACL
func BuildACL(r *util.ACLRequest, logger *logr.Logger) (bool, error) {
	return createACL(r, logger)
}
