package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vnf-config/internal/service"
)

type VNFController struct {
	service *service.VNFService
}

func NewVNFController() *VNFController {
	return &VNFController{service: service.NewVNFService()}
}

func (ctl *VNFController) ListVNFInstances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	items, total, err := ctl.service.List(c, page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (ctl *VNFController) GetVNFInstance(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	resp, err := ctl.service.Get(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ctl *VNFController) DeleteVNFInstance(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := ctl.service.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}


