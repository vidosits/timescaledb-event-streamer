package integration

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/noctarius/timescaledb-event-streamer/internal/supporting/logging"
	inttest "github.com/noctarius/timescaledb-event-streamer/testsupport"
	"testing"
)

type kafkaConsumer struct {
	t         *testing.T
	ready     chan bool
	collected chan bool
	envelopes []inttest.Envelope
}

func NewKafkaConsumer(t *testing.T) (*kafkaConsumer, <-chan bool) {
	kc := &kafkaConsumer{
		t:         t,
		ready:     make(chan bool, 1),
		collected: make(chan bool, 1),
		envelopes: make([]inttest.Envelope, 0),
	}
	return kc, kc.ready
}

func (k *kafkaConsumer) Envelopes() []inttest.Envelope {
	return k.envelopes
}

func (k *kafkaConsumer) Collected() <-chan bool {
	return k.collected
}

func (k *kafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *kafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	kafkaLogger, err := logging.NewLogger("Test_Kafka_Sink")
	if err != nil {
		return err
	}

	k.ready <- true
	for {
		select {
		case message := <-claim.Messages():
			envelope := inttest.Envelope{}
			if err := json.Unmarshal(message.Value, &envelope); err != nil {
				k.t.Error(err)
			}
			kafkaLogger.Infof("EVENT: %+v", envelope)
			k.envelopes = append(k.envelopes, envelope)
			if len(k.envelopes) >= 10 {
				k.collected <- true
			}
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}