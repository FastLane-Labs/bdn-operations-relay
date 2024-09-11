package service

import (
	"fmt"

	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
)

const (
	notificationChannelSize = 10000

	SubscriptionTypeIntent SubscriptionType = "intent"
)

var (
	validSubscriptionTypes = map[SubscriptionType]struct{}{
		SubscriptionTypeIntent: {},
	}

	validSubscriptionTypeList = []SubscriptionType{
		SubscriptionTypeIntent,
	}
)

type Subscription struct {
	ID                  string
	NotificationChannel chan interface{}
	Type                SubscriptionType
	conn                *jsonrpc2.Conn
}

type SubscriptionType string

type SubscriptionManager struct {
	intentsSubscriptions *hashmap.Map[string, []Subscription]
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		intentsSubscriptions: hashmap.New[string, []Subscription](),
	}
}

func (s *SubscriptionManager) Subscribe(remoteAddress string, subscriptionType SubscriptionType, conn *jsonrpc2.Conn) (*Subscription, error) {
	_, valid := validSubscriptionTypes[subscriptionType]
	if !valid {
		return nil, fmt.Errorf("invalid 'subscription_type' param: '%s', valid values are: %v", subscriptionType, validSubscriptionTypeList)
	}

	subs, exists := s.intentsSubscriptions.Get(remoteAddress)
	if exists {
		for i := range subs {
			if subs[i].Type == subscriptionType {
				return nil, fmt.Errorf("subscription already exists for type: %s, id: %s", subscriptionType, subs[i].ID)
			}
		}
	}

	sub := Subscription{
		ID:                  uuid.New().String(),
		NotificationChannel: make(chan interface{}, notificationChannelSize),
		Type:                subscriptionType,
		conn:                conn,
	}
	subs = append(subs, sub)
	s.intentsSubscriptions.Set(remoteAddress, subs)

	return &sub, nil
}

func (s *SubscriptionManager) Unsubscribe(remoteAddress, subscriptionID string) error {
	exists := false
	subs, _ := s.intentsSubscriptions.Get(remoteAddress)
	for i := range subs {
		if subs[i].ID == subscriptionID {
			close(subs[i].NotificationChannel)
			subs = append(subs[:i], subs[i+1:]...)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("subscription not found for id: %s", subscriptionID)
	}

	s.intentsSubscriptions.Set(remoteAddress, subs)

	return nil
}

func (s *SubscriptionManager) Notify(n interface{}) {
	var subType SubscriptionType

	switch n.(type) {
	case *sdk.OnIntentsNotification:
		subType = SubscriptionTypeIntent
	default:
		return
	}

	s.intentsSubscriptions.Range(func(key string, value []Subscription) bool {
		for _, subscription := range value {
			if subscription.Type == subType {
				select {
				case subscription.NotificationChannel <- n:
				default:
					logger.Warn("notification channel for subscription is full, dropping notification")
				}
			}
		}
		return true
	})
}

func (s *SubscriptionManager) Close() {
	s.intentsSubscriptions.Range(func(key string, value []Subscription) bool {
		for _, subscription := range value {
			_ = subscription.conn.Close()
		}
		return true
	})
}
