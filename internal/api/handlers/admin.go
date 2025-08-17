package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/services"
)

type AdminHandler struct {
	accountService      *services.AccountService
	emailAddressService *services.EmailAddressService
}

func NewAdminHandler(accountService *services.AccountService, emailAddressService *services.EmailAddressService) *AdminHandler {
	return &AdminHandler{accountService: accountService, emailAddressService: emailAddressService}
}

// GET /v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// pagination
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		// ignore errors - fallback to default
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	users, err := h.accountService.ListAccounts(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GET /v1/admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.accountService.GetAccount(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DELETE /v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.accountService.DeleteAccount(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /v1/admin/users/:id/revoke-key
func (h *AdminHandler) RevokeKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.accountService.RevokeAPIKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke key"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /v1/admin/users/:id/email-addresses
func (h *AdminHandler) ListEmailAddresses(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	// pagination
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	addrs, err := h.emailAddressService.ListByAccount(id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list email addresses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"email_addresses": addrs})
}

// GET /v1/admin/stats
func (h *AdminHandler) Stats(c *gin.Context) {
	stats, err := h.accountService.GetGlobalStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
