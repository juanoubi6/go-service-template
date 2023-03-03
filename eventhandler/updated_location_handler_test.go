package eventhandler_test

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-service-template/eventhandler"
	"testing"
)

type UpdatedLocationHandlerSuite struct {
	handler *eventhandler.UpdatedLocationEventHandler
	suite.Suite
}

func (s *UpdatedLocationHandlerSuite) SetupSuite() {
	s.handler = eventhandler.CreateUpdatedLocationHandler()
}

func TestUpdatedLocationHandlerSuite(t *testing.T) {
	suite.Run(t, new(UpdatedLocationHandlerSuite))
}

func (s *UpdatedLocationHandlerSuite) Test_Process_Success() {
	locationBytes, err := json.Marshal(location)
	if err != nil {
		s.FailNow("could not marshal Location")
	}

	testMsg := message.NewMessage(uuid.NewString(), locationBytes)

	assert.Nil(s.T(), s.handler.Process(testMsg))
}
