package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vnf-config/internal/dto"
	"vnf-config/internal/service"
)

type DefinitionController struct {
	service *service.DefinitionService
}

func NewDefinitionController() *DefinitionController {
	return &DefinitionController{service: service.NewDefinitionService()}
}

func (ctl *DefinitionController) ListDefinitions(c *gin.Context) {
	vnfID, _ := strconv.Atoi(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	modifiedOnly := c.DefaultQuery("modifiedOnly", "false") == "true"

	items, total, err := ctl.service.List(c, uint(vnfID), page, pageSize, modifiedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (ctl *DefinitionController) CreateDefinition(c *gin.Context) {
	vnfID, _ := strconv.Atoi(c.Param("id"))
	var req dto.DefinitionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	resp, err := ctl.service.Create(c, uint(vnfID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (ctl *DefinitionController) UpdateDefinition(c *gin.Context) {
	vnfID, _ := strconv.Atoi(c.Param("id"))
	defID, _ := strconv.Atoi(c.Param("defId"))
	var req dto.DefinitionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	resp, err := ctl.service.Update(c, uint(vnfID), uint(defID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ctl *DefinitionController) DeleteDefinition(c *gin.Context) {
	vnfID, _ := strconv.Atoi(c.Param("id"))
	defID, _ := strconv.Atoi(c.Param("defId"))
	if err := ctl.service.Delete(c, uint(vnfID), uint(defID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}


