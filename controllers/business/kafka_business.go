package business

import (
	"errors"

	"github.com/kubbee/ccloud-messaging-topology-operator/cross"
	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
	corev1 "k8s.io/api/core/v1"
)

func GetConfigMap(connCredsKafka util.ConnectionCredentials, connCredsSR util.ConnectionCredentials, name string, namespace string, topic string) (*corev1.ConfigMap, error) {

	kafkaKey := cross.ValidateConnCredsData(connCredsKafka.Data("ApiKey"))
	kafkaSecret := cross.ValidateConnCredsData(connCredsKafka.Data("ApiSecret"))
	kafkaEndpoint := cross.ValidateConnCredsData(connCredsKafka.Data("KafkaEndpoint"))

	if kafkaKey == string("") || kafkaSecret == string("") || kafkaEndpoint == string("") {
		return &corev1.ConfigMap{}, errors.New("error to recovery kafka data")
	}

	sRKey := cross.ValidateConnCredsData(connCredsSR.Data("ApiKey"))
	sRSecret := cross.ValidateConnCredsData(connCredsSR.Data("ApiSecret"))
	sREndpoint := cross.ValidateConnCredsData(connCredsSR.Data("SREndpoint"))

	if sRKey == string("") || sRSecret == string("") || sREndpoint == string("") {
		return &corev1.ConfigMap{}, errors.New("error to recovery schema registry data")
	}

	configMapKafka := &util.ConfigMapKafka{
		TopicName:               topic,
		SchemaRegistryURL:       sREndpoint,
		SchemaRegistryApiKey:    sRKey,
		SchemaRegistryApiSecret: sRSecret,
		KafkaURL:                kafkaEndpoint,
		KafkaApiKey:             kafkaKey,
		KafkaApiSecret:          kafkaSecret,
	}

	return cross.CreateConfigMap(configMapKafka, name, namespace), nil
}
