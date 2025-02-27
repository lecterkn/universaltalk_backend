package controller

import (
	"lecter/goserver/internal/app/gochat/application/service"

	"github.com/gin-gonic/gin"
)

type VersionController struct {
	VersionService service.VersionService
}

func NewVersionController(versionService service.VersionService) VersionController {
	return VersionController{
		VersionService: versionService,
	}
}

// @Summary バージョンを取得
// @Accept json
// @Produce json
// @Router /version [get]
func (vc VersionController) Index(ctx *gin.Context) {
	ctx.JSON(200, vc.VersionService.GetVersion())
}
