package settings

import (
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// nodeResponse is a node with its token; the token is only sent when the
// node is created and when a new one is issued.
type nodeResponse struct {
	models.Node
	Token string `json:"token"`
}

// withNode writes a node with its token, or the error.
func (r reply) withNode(c *gin.Context, node *access.NodeToken, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(int(r), nodeResponse{Node: node.Node, Token: node.Token})
}

func (h *Handler) createNode(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if err := httpx.BindJSON(c, &request); err != nil {
		httpx.Fail(c, err)
		return
	}
	node, err := h.svc.CreateNode(c.Request.Context(), request.Name)
	created.withNode(c, node, err)
}

func (h *Handler) nodes(c *gin.Context) {
	nodes, err := h.svc.Nodes(c.Request.Context())
	ok.with(c, nodes, err)
}

// nodeID runs use with the node id of the path; a bad id is answered.
func nodeID(c *gin.Context, use func(id uint)) {
	id, err := httpx.ParseID(c, "node_id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	use(id)
}

func (h *Handler) node(c *gin.Context) {
	nodeID(c, func(id uint) {
		node, err := h.svc.Node(c.Request.Context(), id)
		ok.with(c, node, err)
	})
}

// issueNodeToken gives a node a new token, for setting it up again; the old
// token stops working.
func (h *Handler) issueNodeToken(c *gin.Context) {
	nodeID(c, func(id uint) {
		node, err := h.svc.IssueNodeToken(c.Request.Context(), id)
		ok.withNode(c, node, err)
	})
}

func (h *Handler) deleteNode(c *gin.Context) {
	nodeID(c, func(id uint) {
		done(c, h.svc.DeleteNode(c.Request.Context(), id))
	})
}
