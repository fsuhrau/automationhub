package accessstore

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// Nodes lists all nodes.
func (s *Store) Nodes(ctx context.Context) ([]models.Node, error) {
	var nodes []models.Node
	err := s.db.WithContext(ctx).Find(&nodes).Error
	return nodes, storeerr.Map(err, "nodes")
}

// Node loads the node with the id.
func (s *Store) Node(ctx context.Context, id uint) (*models.Node, error) {
	var node models.Node
	if err := s.db.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, storeerr.Map(err, fmt.Sprintf("node %d", id))
	}
	return &node, nil
}

// CreateNode stores the node and its first token in one transaction.
func (s *Store) CreateNode(ctx context.Context, node *models.Node, token string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(node).Error; err != nil {
			return err
		}
		return replaceToken(tx, node, token)
	})
	return storeerr.Map(err, "node "+node.Name)
}

// ReplaceNodeToken replaces the tokens of the node with the token.
func (s *Store) ReplaceNodeToken(ctx context.Context, node *models.Node, token string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceToken(tx, node, token)
	})
	return storeerr.Map(err, fmt.Sprintf("node %d", node.ID))
}

func replaceToken(tx *gorm.DB, node *models.Node, token string) error {
	if err := tx.Where("node_id = ?", node.ID).Delete(&models.AccessToken{}).Error; err != nil {
		return err
	}
	return tx.Create(&models.AccessToken{NodeID: &node.ID, Token: token, Name: node.Name}).Error
}

// DeleteNode deletes the node and its tokens in one transaction.
func (s *Store) DeleteNode(ctx context.Context, id uint) error {
	what := fmt.Sprintf("node %d", id)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var node models.Node
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if err := tx.Where("node_id = ?", id).Delete(&models.AccessToken{}).Error; err != nil {
			return err
		}
		return tx.Delete(&node).Error
	})
	return storeerr.Map(err, what)
}
