package mocks

// external libraries

type AcknowledgementMock struct{}

func (m AcknowledgementMock) Success() bool {
	return true
}

func (m AcknowledgementMock) Acknowledgement() []byte {
	return []byte("ack")
}
