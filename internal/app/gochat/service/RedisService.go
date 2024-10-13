package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisChannel string
type RedisMessageEvent string

// RedisChannel
const (
	Broadcast = RedisChannel("broadcast")
)

// RedisMessageEvent
const (
	AddMessage    = RedisMessageEvent("add_message")
	UpdateMessage = RedisMessageEvent("update_message")
	DeleteMessage = RedisMessageEvent("delete_message")
)

func (rc RedisChannel) ToString() string {
	return string(rc)
}

type RedisMessage struct {
	SrcUser uuid.UUID         `json:"src"`
	Event   RedisMessageEvent `json:"event"`
	Message string            `json:"message"`
}

func (rm RedisMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(rm)
}

type RedisService struct {
	Context context.Context
	Client  redis.Client
}

func (rs RedisService) Publish(channel RedisChannel, message RedisMessage) (int64, error) {
	return rs.Client.Publish(rs.Context, channel.ToString(), message).Result()
}
