package services

import (
	"testing"

	util "github.com/kubbee/ccloud-messaging-topology-operator/internal"
)

func Test_buildTopic(t *testing.T) {
	topic := &util.NewTopic{
		Tenant:     "odin",
		Namespace:  "test",
		Topic:      "contrato-integrador",
		Partitions: "1",
	}

	if _, err := BuildTopic(topic, "env-123zp6", "lkc-nvjp9z", nil); err != nil {
		t.Fail()
	}

}

func Test_retrieveTopic(t *testing.T) {
	topic := &util.ExistentTopic{
		Tenant: "odin",
		Domain: "cadastral",
		Topic:  "contrato-integrador-cdc-eventsource",
	}

	if _, err := RetrieveTopic(topic, "env-123zp6", "lkc-nvjp9z", nil); err != nil {
		t.Fail()
	}
}
