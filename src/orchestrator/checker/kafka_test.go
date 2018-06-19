package checker

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

func TestNewKafka(t *testing.T) {
	k, err := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"]}`)), nil)
	if err != nil {
		t.Fatalf("Good config should not return err : %s", err.Error())
	}
	if k == nil {
		t.Fatal("Good config should return kafka checker")
	}
}

func TestNewKafkaError(t *testing.T) {
	k, err := newKafka(json.RawMessage([]byte(`{`)), nil)
	if err == nil {
		t.Fatal("Bad config should return err")
	}
	if k != nil {
		t.Fatal("Bad config should not return kafka checker")
	}
}

func TestKafka_StartNewConsumerError(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test"}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		return nil, errors.New("error")
	}
	if err := k.Start(); err == nil {
		t.Fatal("new consumer error must return error")
	}
}

type fakeErrorReporter struct{}

func (fakeErrorReporter) Errorf(string, ...interface{}) {}

func TestKafka_StartTopicError(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test"}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		return mocks.NewConsumer(fakeErrorReporter{}, nil), nil
	}
	if err := k.Start(); err == nil {
		t.Fatal("topic error must return error")
	}
}

func TestKafka_StartConsumePartitionError(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test"}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(fakeErrorReporter{}, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		return mock, nil
	}
	if err := k.Start(); err == nil {
		t.Fatal("consume partition error must return error")
	}
}

func TestKafka_CheckNoCheck(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test"}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		consumerMock := mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte("data")})
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 0 {
		t.Fatal("No check should return no success")
	}
	if len(err) != 1 {
		t.Fatal("No check should return 1 error")
	}
}

func TestKafka_CheckGood(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":[""],"text":"data"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		consumerMock := mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte("data")})
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 1 {
		t.Fatal("Good check should return 1 success")
	}
	if len(err) != 0 {
		t.Fatal("Good check should return 0 error")
	}
}

func TestKafka_CheckBad(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":[""],"text":"data"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		consumerMock := mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte("data2")})
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 0 {
		t.Fatal("Bad check should return no success")
	}
	if len(err) != 1 {
		t.Fatal("Bad check should return 1 error")
	}
	if err[0].Details()["value"] != "data2" {
		t.Fatal("Bad check should report bad data to understand why the check is bad")
	}
}

func TestKafka_CheckGoodWithReplace(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":["_toRemove_"],"text":"data"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		consumerMock := mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte("data_toRemove_")})
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 1 {
		t.Fatal("Good check should return 1 success")
	}
	if success[0].Details()["success"] != "1" {
		t.Fatal("Succes should report description of check")
	}
	if len(err) != 0 {
		t.Fatal("Good check should return 0 error")
	}
}

func TestKafka_CheckBadNoMessage(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":[""],"text":"data"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 0 {
		t.Fatal("No message check should return no success")
	}
	if len(err) != 1 {
		t.Fatal("No message check should return 1 error")
	}
}

func TestKafka_CheckSimpleRemovalOfDynamicValues(t *testing.T) {
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":["\"applicationID\":[^,]+,","\"applicationName\":[^,]+\""],"text":"{\"devEUI\":\"3c0a1f3811e5c56b\",}"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0}
		mock.SetTopicMetadata(metadata)
		consumerMock := mock.ExpectConsumePartition("test", 0, sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte(`{"devEUI":"3c0a1f3811e5c56b","applicationID":"19","applicationName":"kafka"}`)})
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	if len(success) != 1 {
		t.Fatal("Good check should return 1 success")
	}
	if success[0].Details()["success"] != "1" {
		t.Fatal("Succes should report description of check")
	}
	if len(err) != 0 {
		t.Fatal("Good check should return 0 error")
	}
}

func TestKafka_MultiplePartitions(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("When the consumer and the partition are closed, try to send another message must panic")
		}
	}()
	k, _ := newKafka(json.RawMessage([]byte(`{"address": ["127.0.0.1:9092"], "topic": "test", "checks": [{"description": "1","remove":["\"applicationID\":[^,]+,","\"applicationName\":[^,]+\""],"text":"{\"devEUI\":\"3c0a1f3811e5c56b\",}"}]}`)), nil)
	k.(*kafka).newConsumer = func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
		mock := mocks.NewConsumer(t, nil)
		metadata := make(map[string][]int32)
		metadata["test"] = []int32{0, 1, 2, 3, 4}
		mock.SetTopicMetadata(metadata)
		for i := 0; i < 5; i++ {
			consumerMock := mock.ExpectConsumePartition("test", int32(i), sarama.OffsetNewest)
			consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte(`{"devEUI":"3c0a1f3811e5c56b","applicationID":"19","applicationName":"kafka"}`)})
		}
		return mock, nil
	}
	k.Start()
	time.Sleep(10 * time.Millisecond)
	success, err := k.Check()
	for i := 0; i < 5; i++ {
		consumerMock := k.(*kafka).kafkaConsumer.(*mocks.Consumer).ExpectConsumePartition("test", int32(i), sarama.OffsetNewest)
		consumerMock.YieldMessage(&sarama.ConsumerMessage{Value: []byte(`{"devEUI":"3c0a1f3811e5c56b","applicationID":"19","applicationName":"kafka"}`)})
	}
	if len(success) != 5 {
		t.Fatal("Good check should return 5 success")
	}
	if success[0].Details()["success"] != "1" {
		t.Fatal("Succes should report description of check")
	}
	if len(err) != 0 {
		t.Fatal("Good check should return 0 error")
	}
}
