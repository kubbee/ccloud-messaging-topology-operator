package cross

import (
	"strings"

	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateConfigMap(configMapKafka *util.ConfigMapKafka, name string, namespace string) *corev1.ConfigMap {

	var immutable bool = true

	var labels = make(map[string]string)
	labels["name"] = name
	labels["owner"] = "ccloud-messaging-topology-operator"
	labels["controller"] = "kafkaproducer_controller"

	n := strings.ReplaceAll(name, "-", "_")

	mapper := make(map[string]string)
	mapper[n+"_topic_name"] = configMapKafka.TopicName
	mapper[n+"_schema_registry_url"] = configMapKafka.SchemaRegistryURL
	mapper[n+"_schema_registry_api"] = configMapKafka.SchemaRegistryApiKey
	mapper[n+"_schema_registry_secret"] = configMapKafka.SchemaRegistryApiSecret
	mapper[n+"_kafka_url"] = configMapKafka.KafkaURL
	mapper[n+"_kafka_api"] = configMapKafka.KafkaApiKey
	mapper[n+"_kafka_secret"] = configMapKafka.KafkaApiSecret

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data:      mapper,
		Immutable: &immutable,
	}
}
