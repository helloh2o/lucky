package natsq

import (
	"errors"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/log"
	"github.com/nats-io/stan.go"
	"runtime"
	"time"
)

var ns stan.Conn

func InitClient(clusterId, natsUrl string) stan.Conn {
	var err error
	ns, err = stan.Connect(clusterId, uuid.New().String(), stan.NatsURL(natsUrl))
	if err != nil {
		if runtime.GOOS == "windows" {
			log.Error("nats client is not started. err:%v", err)
		} else {
			panic(err)
		}
	} else {
		log.Release("nats client is running ...")
		return ns
	}
	return nil
}

func InitOneClient(clusterId, natsUrl, clientId string) stan.Conn {
	var err error
	ns, err = stan.Connect(clusterId, clientId, stan.NatsURL(natsUrl), func(options *stan.Options) error {
		options.AllowCloseRetry = true
		return nil
	})
	if err != nil {
		if runtime.GOOS == "windows" {
			log.Error("nats client is not started. err:%v", err)
		} else {
			panic(err.Error() + "=>" + clientId)
		}
	} else {
		log.Release("nats client is running ...")
		return ns
	}
	return nil
}

// Subscribe 订阅
func Subscribe(subject string, callback func(m *stan.Msg)) {
	if ns == nil {
		return
	}
	// Simple Async Subscriber
	_, err := ns.Subscribe(subject, callback,
		stan.DeliverAllAvailable(),   // StartWithLastReceived()), StartAtSequence(22)
		stan.SetManualAckMode(),      // Manual Ack
		stan.AckWait(time.Second*30), // Wait Act
		stan.MaxInflight(2048))       // Speed Limit
	if err != nil {
		log.Error("Sub %s ,err %v", subject, err)
		return
	}
}

// SubscribeDurable 持久化订阅
func SubscribeDurable(subject, durableName string, callback func(m *stan.Msg)) *stan.Subscription {
	if ns == nil {
		return nil
	}
	// Simple Async Subscriber
	subscription, err := ns.Subscribe(subject, callback,
		stan.DurableName(durableName), // durable name
		stan.DeliverAllAvailable(),    // StartWithLastReceived()), StartAtSequence(22)
		stan.SetManualAckMode(),       // Manual Ack
		stan.MaxInflight(2048))        // Speed Limit
	if err != nil {
		log.Error("Sub %s ,err %v", subject, err)
		return nil
	}
	return &subscription
}

// Pub 发布
func Pub(subject string, msg []byte) error {
	if ns == nil {
		log.Error("pub error:%s", "nats client is nil")
		return errors.New("ns is nil")
	}
	// Simple Synchronous Publisher
	guid, err := ns.PublishAsync(subject, msg, func(s string, err error) {
		if err != nil {
			log.Error("pub error: subugot sub ack %s, err:%v", s, err)
		} else {
			log.Release("client act msg:%s", s)
		}
	})
	if err != nil {
		log.Error("pub error: %v", err)
		return err
	}
	log.Release("pub_message:%s msg:%s, guid:%s", subject, string(msg), guid)
	return nil
}
