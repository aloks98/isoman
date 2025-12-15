package api

import (
	"net/http"
	"strconv"

	"linux-iso-manager/internal/service"

	"github.com/gin-gonic/gin"
)

// StatsHandlers holds references to stats service.
type StatsHandlers struct {
	statsService *service.StatsService
}

// NewStatsHandlers creates a new StatsHandlers instance.
func NewStatsHandlers(statsService *service.StatsService) *StatsHandlers {
	return &StatsHandlers{
		statsService: statsService,
	}
}

// GetStats returns aggregated statistics.
func (h *StatsHandlers) GetStats(c *gin.Context) {
	stats, err := h.statsService.GetStats()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to retrieve statistics")
		return
	}

	SuccessResponse(c, http.StatusOK, stats)
}

// GetDownloadTrends returns download trends over time.
func (h *StatsHandlers) GetDownloadTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "daily") // daily or weekly
	daysStr := c.DefaultQuery("days", "30")

	// Validate period
	if period != "daily" && period != "weekly" {
		period = "daily"
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	trends, err := h.statsService.GetDownloadTrends(period, days)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to retrieve download trends")
		return
	}

	SuccessResponse(c, http.StatusOK, trends)
}
