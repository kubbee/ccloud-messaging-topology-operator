package services

import (
	"testing"
)

func Test_existsTopicName(t *testing.T) {

	b, _ := existsTopicName("lkc-nvjp9z", "odin-cadastral-contrato-integrador-cdc-eventsource")

	if !b {
		t.Fail()
	}
}
