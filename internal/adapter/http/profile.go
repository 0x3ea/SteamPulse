package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/0x3ea/SteamPulse/internal/core"
)

// Transfrom:
// http request -> core call
// result of core call -> http response

type Handler struct {
	core *core.Service
}

func NewHandler(svc *core.Service) *Handler {
	return &Handler{core: svc}
}

// handle [GET] /api/profile/:id
// id is a 17-digit SteamID64
func (h *Handler) GetProfile(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.core.GetProfile(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found or private"})
		default:
			// Upstream Steam failure (bad key, network, rate limit...)
			c.JSON(http.StatusBadGateway, gin.H{"error": "steam lookup failed"})
		}
		return
	}
	c.JSON(http.StatusOK, profile)
}
