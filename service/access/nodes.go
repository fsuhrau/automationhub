package access

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/service/internal/validate"
	"github.com/fsuhrau/automationhub/storage/models"
)

// NodeToken is a node with its token. The token is only handed out when the
// node is created and when a new one is issued.
type NodeToken struct {
	Node  models.Node
	Token string
}

// CreateNode creates a node with a new identifier and its first token.
func (s *Service) CreateNode(ctx context.Context, name string) (*NodeToken, error) {
	name, err := validate.TrimRequired(name, "the node needs a name")
	if err != nil {
		return nil, err
	}
	identifier, err := s.newUUID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, err, "create node identifier")
	}
	value, err := s.newUUID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, err, "create node token")
	}
	created := &NodeToken{Node: models.Node{Identifier: identifier, Name: name}, Token: value}
	if err := s.repo.CreateNode(ctx, &created.Node, value); err != nil {
		return nil, err
	}
	return created, nil
}

// Nodes lists the nodes with their connection status.
func (s *Service) Nodes(ctx context.Context) ([]models.Node, error) {
	nodes, err := s.repo.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	for i := range nodes {
		nodes[i].Status, _ = s.status.GetStatus(manager.NodeIdentifier(nodes[i].Identifier))
	}
	return nodes, nil
}

// Node returns the node with the id.
func (s *Service) Node(ctx context.Context, id uint) (*models.Node, error) {
	return s.repo.Node(ctx, id)
}

// IssueNodeToken gives a node a new token, for setting it up again; the old
// token stops working.
func (s *Service) IssueNodeToken(ctx context.Context, id uint) (*NodeToken, error) {
	node, err := s.repo.Node(ctx, id)
	if err != nil {
		return nil, err
	}
	value, err := s.newUUID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, err, "create node token")
	}
	if err := s.repo.ReplaceNodeToken(ctx, node, value); err != nil {
		return nil, err
	}
	return &NodeToken{Node: *node, Token: value}, nil
}

// DeleteNode removes a node and its tokens, so it can't connect anymore.
func (s *Service) DeleteNode(ctx context.Context, id uint) error {
	return s.repo.DeleteNode(ctx, id)
}
