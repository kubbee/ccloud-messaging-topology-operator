package cross

import (
	"errors"

	"github.com/go-logr/logr"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"

	corev1 "k8s.io/api/core/v1"
)

func GetKafkaCredentials(secret *corev1.Secret, secretType int, logger *logr.Logger) util.ConnectionCredentials {
	switch secretType {

	case 1:
		logger.Info("Reading readCredentialsFromKubernetesSecret")
		return readCredentialsFromKubernetesSecret(secret)

	case 2:
		logger.Info("Reading readCredentialsKafkaFromKubernetesSecret")
		return readCredentialsKafkaFromKubernetesSecret(secret)

	case 3:
		logger.Info("Reading readCredentialsSchemaRegistryFromKubernetesSecret")
		return readCredentialsSchemaRegistryFromKubernetesSecret(secret)

	default:
		logger.Error(errors.New("internal error missing secretype"), "")
		return nil
	}
}

// readCredentialsFromKubernetesSecret read the cluster kafka credentials
func readCredentialsFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"tenant":        secret.Data["tenant"],
			"clusterId":     secret.Data["clusterId"],
			"environmentId": secret.Data["environmentId"],
		},
	}
}

// readCredentialsKafkaFromKubernetesSecret This function is responsible to ready the content of the secret and return for execute operations with the data
func readCredentialsKafkaFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"ApiKey":        secret.Data["ApiKey"],
			"ApiSecret":     secret.Data["ApiSecret"],
			"KafkaEndpoint": secret.Data["KafkaEndpoint"],
		},
	}
}

// readCredentialsSchemaRegistryFromKubernetesSecret This function is responsible to ready the content of the secret and return for execute operations with the data
func readCredentialsSchemaRegistryFromKubernetesSecret(secret *corev1.Secret) *util.ClusterCredentials {
	return &util.ClusterCredentials{
		DataContent: map[string][]byte{
			"ApiKey":     secret.Data["ApiKey"],
			"ApiSecret":  secret.Data["ApiSecret"],
			"SREndpoint": secret.Data["SREndpoint"],
		},
	}
}
