package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type UserHandler struct {
	db *store.DB
}

func NewUserHandler(db *store.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.db.ListUsers()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	var req struct {
		ID   uint   `json:"id"`
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.Error(c, 400, "invalid request, id required")
		return
	}
	if req.Role != "admin" && req.Role != "user" {
		response.Error(c, 400, "role must be admin or user")
		return
	}

	user, err := h.db.GetUserByID(req.ID)
	if err != nil {
		response.Error(c, 404, "user not found")
		return
	}

	user.Role = req.Role
	if err := h.db.UpdateUserRole(user); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	uid := c.GetUint("user_id")
	if uint(id) == uid {
		response.Error(c, 400, "cannot delete yourself")
		return
	}

	if err := h.db.DeleteUser(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}
