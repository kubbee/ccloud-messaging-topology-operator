package services

import "testing"

func Test_createNewTopic(t *testing.T) {

	_, err := createNewTopic("odin", "contrato-integrador-cdc-eventsource", "cadastral", "", "lkc-nvjp9z")

	if err != nil {
		t.Fail()
	}

}
