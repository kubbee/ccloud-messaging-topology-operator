package services

import (
	"github.com/go-logr/logr"

	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

// BuildServiceAccount return the struct ServiceAccount
func BuildServiceAccount(sa, description string, logger *logr.Logger) (*util.ServiceAccount, error) {
	logger.Info("start::BuildServiceAccount")
	return createServiceAccount(sa, description, logger)
}
