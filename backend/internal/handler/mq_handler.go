package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/mq"
)

// SendMessageHandler handles POST /mq/send
// Simulates sending a task to the queue
func SendMessageHandler(c *gin.Context) {
	message := c.DefaultQuery("msg", fmt.Sprintf("Task at %s", time.Now().Format(time.RFC3339)))

	err := mq.SendTask(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Message sent", "message": message})
}
