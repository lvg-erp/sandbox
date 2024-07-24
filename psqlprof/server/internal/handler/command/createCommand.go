package command

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"psqlprof/server/internal/entity"
)

// CreateCommand @Summary Create command
// @Description Add new command to DB
// @Tags command
// @Accept json
// @Produce json
// @Param command body entity.CreateCommandReq true "Script and description for script"
// @Success 200 {object} entity.CreateCommandRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /command [post]

func (h *Handler) validateReqCommand(req entity.CreateCommandReq) error {
	if len(req.Script) == 0 {
		return errors.New("script must not be empty")
	}

	if len(req.Description) == 0 {
		return errors.New("description must not be empty")
	}

	return nil
}

func (h *Handler) CreateCommand(c *gin.Context) {

	var cd entity.CreateCommandReq
	if err := c.ShouldBindJSON(&cd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validateReqCommand(cd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	res, err := h.Service.CreateCommand(c.Request.Context(), &cd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, res)

}
