// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package gochat

import (
	"github.com/google/wire"
	"lecter/goserver/internal/app/gochat/application/service"
	"lecter/goserver/internal/app/gochat/application/service/authorization"
	"lecter/goserver/internal/app/gochat/infrastructure/db"
	"lecter/goserver/internal/app/gochat/infrastructure/repository/implements"
	"lecter/goserver/internal/app/gochat/presentation/controller"
)

// Injectors from wire.go:

func InitializeControllerSet() *ControllerSet {
	gormDB := db.Database()
	channelRepositoryImpl := implements.NewChannelRepositoryImpl(gormDB)
	channelService := service.NewChannelService(channelRepositoryImpl)
	channelController := controller.NewChannelController(channelService)
	channelLanguageRepositoryImpl := implements.NewChannelLanguageRepositoryImpl(gormDB)
	channelLanguageService := service.NewChannelLanguageService(channelRepositoryImpl, channelLanguageRepositoryImpl)
	channelLanguageController := controller.NewChannelLanguageController(channelLanguageService)
	messageRepositoryImpl := implements.NewMessageRepositoryImpl(gormDB)
	messageDomainService := service.NewMessageDomainService(messageRepositoryImpl)
	redisService := service.NewRedisService()
	messageService := service.NewMessageService(messageRepositoryImpl, channelRepositoryImpl, messageDomainService, redisService)
	messageController := controller.NewMessageController(messageService)
	userRepositoryImpl := implements.NewUserRepositoryImpl(gormDB)
	userService := service.NewUserService(userRepositoryImpl)
	userProfileRepositoryImpl := implements.NewUserProfileRepositoryImpl(gormDB)
	userProfileService := service.NewUserProfileService(userProfileRepositoryImpl)
	userController := controller.NewUserController(userService, userProfileService)
	userProfileController := controller.NewUserProfileController(userProfileService)
	versionService := service.NewVersionService()
	versionController := controller.NewVersionController(versionService)
	gochatControllerSet := &ControllerSet{
		ChannelController:         channelController,
		ChannelLanguageController: channelLanguageController,
		MessageController:         messageController,
		UserController:            userController,
		UserProfileController:     userProfileController,
		VersionController:         versionController,
	}
	return gochatControllerSet
}

func InitializeJwtAuthorizationService() *authorization.JwtAuthorizationService {
	gormDB := db.Database()
	userRepositoryImpl := implements.NewUserRepositoryImpl(gormDB)
	jwtAuthorizationService := authorization.NewJwtAuthorizationService(userRepositoryImpl)
	return jwtAuthorizationService
}

// wire.go:

var databaseSet = wire.NewSet(db.Database)

var repositorySet = wire.NewSet(implements.NewUserRepositoryImpl, implements.NewUserProfileRepositoryImpl, implements.NewMessageRepositoryImpl, implements.NewChannelRepositoryImpl, implements.NewChannelLanguageRepositoryImpl)

var serviceSet = wire.NewSet(service.NewChannelLanguageService, service.NewChannelService, service.NewMessageDomainService, service.NewMessageService, service.NewRedisService, service.NewUserProfileService, service.NewUserService, service.NewVersionService, authorization.NewJwtAuthorizationService)

var controllerSet = wire.NewSet(controller.NewChannelController, controller.NewChannelLanguageController, controller.NewMessageController, controller.NewUserController, controller.NewUserProfileController, controller.NewVersionController)

type ControllerSet struct {
	ChannelController         controller.ChannelController
	ChannelLanguageController controller.ChannelLanguageController
	MessageController         controller.MessageController
	UserController            controller.UserController
	UserProfileController     controller.UserProfileController
	VersionController         controller.VersionController
}
