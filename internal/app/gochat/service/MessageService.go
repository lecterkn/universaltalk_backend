package service

import (
	"context"
	"encoding/json"
	"fmt"
	"lecter/goserver/internal/app/gochat/controller/response"
	"lecter/goserver/internal/app/gochat/enum/language"
	"lecter/goserver/internal/app/gochat/model"
	"lecter/goserver/internal/app/gochat/repository"
	"lecter/goserver/internal/app/gochat/service/output"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MessageService struct{}

var redisService = RedisService{
	Context: context.TODO(),
	Client: *redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}),
}
var messageDomainService = MessageDomainService{}
var messageRepository = repository.MessageRepository{}

func (MessageService) GetMessages(userId, channelId uuid.UUID, lastId *uuid.UUID, limit int, langCode *string) (*output.MessageOutput, *response.ErrorResponse) {
	// 言語を取得
	var lang *language.Language = nil
	if langCode != nil {
		langu, err := language.GetLanguageFromCode(*langCode)
		if err != nil {
			return nil, response.ValidationError("invalid language code")
		}
		lang = &langu
	}
	// チャンネルを取得
	channel, err := channelRepository.Select(channelId)
	if err != nil {
		return nil, response.NotFoundError("the channel does not exists")
	}
	// 権限確認
	if !isChannelReadable(*channel, userId) {
		return nil, response.ForbiddenError("permission error")
	}
	// 原文を取得
	if lang == nil {
		// メッセージを一覧取得
		models, error := messageDomainService.GetOriginalMessage(channelId, lastId, limit)
		if error != nil {
			return nil, error
		}
		messageOutput := output.MessageOutput{
			Messages: []output.MessageItem{},
		}
		for _, message := range models {
			messageOutput.Messages = append(messageOutput.Messages, output.MessageItem{
				Id:        message.Id,
				ChannelId: message.ChannelId,
				UserId:    message.UserId,
				Message:   message.Message,
				CreatedAt: message.CreatedAt,
				UpdatedAt: message.UpdatedAt,
			})
		}
		return &messageOutput, nil
		// 翻訳されたメッセージを取得
	} else {
		// メッセージを一覧取得
		models, error := messageDomainService.GetTranslatedMessage(channelId, lastId, limit, *lang)
		if error != nil {
			return nil, error
		}
		messageOutput := output.MessageOutput{
			Messages: []output.MessageItem{},
		}
		for _, message := range models {
			messageOutput.Messages = append(messageOutput.Messages, output.MessageItem{
				Id:        message.Id,
				ChannelId: message.ChannelId,
				UserId:    message.UserId,
				Message:   message.MessageContent,
				CreatedAt: message.CreatedAt,
				UpdatedAt: message.UpdatedAt,
			})
		}
		return &messageOutput, nil
	}
}

func (MessageService) CreateMessage(userId, channelId uuid.UUID, message string) (*model.MessageModel, *response.ErrorResponse) {
	channel, err := channelRepository.Select(channelId)
	if err != nil {
		return nil, response.NotFoundError("the channel does not exist")
	}
	if !isChannelWritable(*channel, userId) {
		return nil, response.ForbiddenError("permission error")
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, response.InternalError("failed to generate uuid")
	}
	model := &model.MessageModel{
		Id:        id,
		ChannelId: channel.Id,
		UserId:    userId,
		Message:   message,
		Deleted:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	model, err = messageRepository.Create(*model)
	if err != nil {
		return nil, response.InternalError("failed to create message")
	}
	// jsonに変換
	messageJson, err := json.Marshal(model)
	if err != nil {
		fmt.Println("failed to unmarshal message")
	}
	// redisにパブリッシュ
	_, err = redisService.Publish(Broadcast, RedisMessage{
		SrcUser: model.UserId,
		Event:   AddMessage,
		Message: string(messageJson),
	})
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("failed to publish message")
	}
	return model, nil
}

func (MessageService) UpdateMessage(userId, channelId, messageId uuid.UUID, message string) (*model.MessageModel, *response.ErrorResponse) {
	channel, err := channelRepository.Select(channelId)
	if err != nil {
		return nil, response.NotFoundError("the channel does not exist")
	}
	if !isChannelWritable(*channel, userId) {
		return nil, response.ForbiddenError("permission error")
	}
	model, err := messageRepository.Select(messageId)
	if err != nil {
		return nil, response.NotFoundError("the message does not exist")
	}
	model.Message = message
	model.UpdatedAt = time.Now()
	model, err = messageRepository.Update(*model)
	if err != nil {
		return nil, response.InternalError("failed to update message")
	}
	// jsonに変換
	messageJson, err := json.Marshal(model)
	if err != nil {
		fmt.Println("failed to unmarshal message")
	}
	// redisにパブリッシュ
	_, err = redisService.Publish(Broadcast, RedisMessage{
		SrcUser: model.UserId,
		Event:   UpdateMessage,
		Message: string(messageJson),
	})
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("failed to publish message")
	}
	return model, nil
}

func (MessageService) DeleteMessage(userId, channelId, messageId uuid.UUID) *response.ErrorResponse {
	channel, err := channelRepository.Select(channelId)
	if err != nil {
		return response.NotFoundError("the channel does not exist")
	}
	model, err := messageRepository.Select(messageId)
	if err != nil {
		return response.NotFoundError("the message does not exist")
	}
	if !isMessageDeletable(*channel, *model, userId) {
		return response.ForbiddenError("permission error")
	}
	model.Deleted = true
	model.UpdatedAt = time.Now()

	_, err = messageRepository.Update(*model)
	if err != nil {
		return response.InternalError("failed to delete message")
	}
	// jsonに変換
	messageJson, err := json.Marshal(model)
	if err != nil {
		fmt.Println("failed to unmarshal message")
	}
	// redisにパブリッシュ
	_, err = redisService.Publish(Broadcast, RedisMessage{
		SrcUser: model.UserId,
		Event:   DeleteMessage,
		Message: string(messageJson),
	})
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("failed to publish message")
	}
	return nil
}

func isChannelReadable(channel model.ChannelModel, userId uuid.UUID) bool {
	// チャンネルがプライベートの場合
	if channel.Permission == 2 && channel.OwnerId != userId {
		return false
	}
	return true
}

func isChannelWritable(channel model.ChannelModel, userId uuid.UUID) bool {
	// チャンネルが読み取り専用の場合
	if channel.Permission == 0 && channel.OwnerId != userId {
		return false
	}
	// チャンネルがプライベートの場合
	if channel.Permission == 2 && channel.OwnerId != userId {
		return false
	}
	return true
}

func isMessageDeletable(channel model.ChannelModel, message model.MessageModel, userId uuid.UUID) bool {
	if message.UserId != userId && channel.OwnerId != userId {
		return false
	}
	return true
}
