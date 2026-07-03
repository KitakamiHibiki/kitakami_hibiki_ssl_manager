package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/auth"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type AuthHandler struct {
	db     *store.DB
	secret string
}

func NewAuthHandler(db *store.DB, secret string) *AuthHandler {
	return &AuthHandler{db: db, secret: secret}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
		response.Error(c, 400, "email and password required")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "hash failed")
		return
	}

	username := strings.Split(req.Email, "@")[0]
	user := &store.User{
		Email:    req.Email,
		Username: username,
		Password: string(hashed),
		Role:     "user",
	}
	if err := h.db.CreateUser(user); err != nil {
		response.Error(c, 409, "email already registered")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.Email, h.secret)
	if err != nil {
		response.Error(c, 500, "token generation failed")
		return
	}

	response.OK(c, gin.H{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
		"email":    user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request")
		return
	}

	user, err := h.db.FindUserByEmail(req.Email)
	if err != nil {
		response.Error(c, 401, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Error(c, 401, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.Email, h.secret)
	if err != nil {
		response.Error(c, 500, "token generation failed")
		return
	}

	response.OK(c, gin.H{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
		"email":    user.Email,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	username := c.GetString("username")
	role := c.GetString("role")
	email := c.GetString("email")
	response.OK(c, gin.H{
		"username": username,
		"role":     role,
		"email":    email,
	})
}
