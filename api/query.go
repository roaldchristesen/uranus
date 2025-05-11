package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func QueryHandler(c *gin.Context) {

	modeStr := c.Query("mode")
	fmt.Println("query mode:", modeStr)

	switch modeStr {
	case "event":
		QueryEvent(c)
		break

	case "venue":
		// QueryVenue(c)
		break

	case "space":
		// QuerySpace(c)
		break

	case "organization":
		// QueryOrganization(c)
		break

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("unknown mode: %s", modeStr),
		})
	}

}
