// Package access is the use cases of who may use the hub: the users, the
// access tokens of a project, the personal tokens of a user and the nodes
// with their tokens. The tokens of a project are only reachable through the
// project that owns them.
package access

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gofrs/uuid"
)

// UserRepo reads users.
type UserRepo interface {
	// Users lists all users.
	Users(ctx context.Context) ([]models.User, error)
	// FindUser loads the stored user of a session, or returns an
	// apperr.NotFound error.
	FindUser(ctx context.Context, who authentication.User) (*models.User, error)
}

// ProjectTokenRepo stores the access tokens of projects. Lookups that find
// nothing return an apperr.NotFound error.
type ProjectTokenRepo interface {
	CreateToken(ctx context.Context, token *models.AccessToken) error
	ProjectTokens(ctx context.Context, projectID uint) ([]models.AccessToken, error)
	DeleteProjectToken(ctx context.Context, projectID, tokenID uint) error
}

// PersonalTokenRepo lists and deletes the tokens of a user; they are
// created with CreateToken.
type PersonalTokenRepo interface {
	PersonalTokens(ctx context.Context, userID uint) ([]models.AccessToken, error)
	DeletePersonalToken(ctx context.Context, userID, tokenID uint) error
}

// NodeRepo stores nodes and their tokens. A node has one token at a time.
type NodeRepo interface {
	Nodes(ctx context.Context) ([]models.Node, error)
	Node(ctx context.Context, id uint) (*models.Node, error)
	// CreateNode stores the node and its first token in one transaction.
	CreateNode(ctx context.Context, node *models.Node, token string) error
	// ReplaceNodeToken replaces the tokens of the node with the token.
	ReplaceNodeToken(ctx context.Context, node *models.Node, token string) error
	// DeleteNode deletes the node and its tokens in one transaction.
	DeleteNode(ctx context.Context, id uint) error
}

// Repo is everything the service needs from the storage.
type Repo interface {
	UserRepo
	ProjectTokenRepo
	PersonalTokenRepo
	NodeRepo
}

// NodeStatus tells whether a node is connected to this master.
type NodeStatus interface {
	GetStatus(nodeIdentifier manager.NodeIdentifier) (int, error)
}

// Deps are the dependencies of the Service. The generators and the clock
// default to the real ones.
type Deps struct {
	Repo        Repo
	Status      NodeStatus
	NewUUID     func() (string, error)
	NewPersonal func() (plain, stored, hint string, err error)
	Now         func() time.Time
}

// Service implements the access use cases.
type Service struct {
	repo        Repo
	status      NodeStatus
	newUUID     func() (string, error)
	newPersonal func() (plain, stored, hint string, err error)
	now         func() time.Time
}

// New returns a Service on the dependencies.
func New(deps Deps) *Service {
	if deps.NewUUID == nil {
		deps.NewUUID = newUUID
	}
	if deps.NewPersonal == nil {
		deps.NewPersonal = token.NewPersonal
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return &Service{repo: deps.Repo, status: deps.Status, newUUID: deps.NewUUID, newPersonal: deps.NewPersonal, now: deps.Now}
}

func newUUID() (string, error) {
	id, err := uuid.NewV4()
	return id.String(), err
}

// Users lists all users.
func (s *Service) Users(ctx context.Context) ([]models.User, error) {
	return s.repo.Users(ctx)
}
