package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/deploy"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/store"
)

type NodeHandler struct {
	db *store.DB
}

func NewNodeHandler(db *store.DB) *NodeHandler {
	return &NodeHandler{db: db}
}

func (h *NodeHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}
	userID := c.GetUint("user_id")
	node, err := h.db.ListNodeByID(uint(id))
	if err != nil || node.UserID != userID {
		response.Error(c, 404, "node not found")
		return
	}
	if err := deploy.TestConnection(node); err != nil {
		response.Error(c, 400, "连接失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "连接成功"})
}

func (h *NodeHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	nodes, err := h.db.ListNodeByUser(userID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	if nodes == nil {
		nodes = []store.DeployNode{}
	}
	response.OK(c, nodes)
}

func (h *NodeHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		NodeType string `json:"node_type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		AuthType string `json:"auth_type"`
		Password string `json:"password"`
		SSHKey   string `json:"ssh_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || req.Host == "" || req.Username == "" {
		response.Error(c, 400, "name, host and username are required")
		return
	}
	if req.NodeType == "" {
		req.NodeType = "ssh"
	}
	if req.Port == 0 {
		req.Port = 22
	}
	if req.AuthType == "" {
		req.AuthType = "password"
	}
	node := &store.DeployNode{
		UserID:   c.GetUint("user_id"),
		Name:     req.Name,
		NodeType: req.NodeType,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		AuthType: req.AuthType,
		Password: req.Password,
		SSHKey:   req.SSHKey,
	}
	if err := h.db.CreateNode(node); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, node)
}

func (h *NodeHandler) Update(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		AuthType string `json:"auth_type"`
		Password string `json:"password"`
		SSHKey   string `json:"ssh_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.Error(c, 400, "invalid id")
		return
	}
	userID := c.GetUint("user_id")
	node, err := h.db.ListNodeByID(req.ID)
	if err != nil || node.UserID != userID {
		response.Error(c, 404, "node not found")
		return
	}
	if req.Name != "" {
		node.Name = req.Name
	}
	if req.Host != "" {
		node.Host = req.Host
	}
	if req.Port > 0 {
		node.Port = req.Port
	}
	if req.Username != "" {
		node.Username = req.Username
	}
	if req.AuthType != "" {
		node.AuthType = req.AuthType
	}
	if req.Password != "" {
		node.Password = req.Password
	}
	if req.SSHKey != "" {
		node.SSHKey = req.SSHKey
	}
	if err := h.db.UpdateNode(node); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, node)
}

func (h *NodeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}
	userID := c.GetUint("user_id")
	node, err := h.db.ListNodeByID(uint(id))
	if err != nil || node.UserID != userID {
		response.Error(c, 404, "node not found")
		return
	}
	if err := h.db.DeleteNode(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}
