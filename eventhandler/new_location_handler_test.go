package eventhandler_test

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-service-template/domain"
	"go-service-template/eventhandler"
	"testing"
)

var (
	location = domain.Location{Name: "New Name"}
)

type NewLocationHandlerSuite struct {
	handler *eventhandler.NewLocationEventHandler
	suite.Suite
}

func (s *NewLocationHandlerSuite) SetupSuite() {
	s.handler = eventhandler.CreateNewLocationHandler()
}

func TestNewLocationHandlerSuite(t *testing.T) {
	suite.Run(t, new(NewLocationHandlerSuite))
}

func (s *NewLocationHandlerSuite) Test_Process_Success() {
	locationBytes, err := json.Marshal(location)
	if err != nil {
		s.FailNow("could not marshal Location")
	}

	testMsg := message.NewMessage(uuid.NewString(), locationBytes)

	assert.Nil(s.T(), s.handler.Process(testMsg))
}
