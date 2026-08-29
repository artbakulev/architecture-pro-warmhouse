package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"smarthome/proto/remotecontrol"
	"smarthome/services"

	"github.com/gin-gonic/gin"
)

type CommandHandler struct {
	CommandService *services.CommandService
}

func NewCommandHandler(commandService *services.CommandService) *CommandHandler {
	return &CommandHandler{CommandService: commandService}
}

func (h *CommandHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/sensors/:id/commands", h.SendCommand)
	router.GET("/commands/:id", h.GetCommandStatus)
}

type sendCommandRequest struct {
	Code    string                 `json:"code" binding:"required"`
	Payload map[string]interface{} `json:"payload"`
}

func (h *CommandHandler) SendCommand(c *gin.Context) {
	sensorID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	var request sendCommandRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.CommandService.Send(context.Background(), sensorID, request.Code, request.Payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, commandResponse(result))
}

func (h *CommandHandler) GetCommandStatus(c *gin.Context) {
	result, err := h.CommandService.GetResult(context.Background(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commandResponse(result))
}

func commandResponse(result *remotecontrol.CommandResult) gin.H {
	response := gin.H{
		"id":         result.Id,
		"status":     strings.ToLower(strings.TrimPrefix(result.Status.String(), "COMMAND_STATUS_")),
		"created_at": result.CreatedAt.AsTime(),
	}
	if result.FinishedAt != nil {
		response["finished_at"] = result.FinishedAt.AsTime()
	}
	return response
}
