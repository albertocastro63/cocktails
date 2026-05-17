package store

import (
	"errors"

	"github.com/almc/cocktails/internal/model"
)

var ErrDuplicate = errors.New("duplicate")

type RecipeStore interface {
	Create(recipe *model.Recipe) error
	GetByID(id string) (*model.Recipe, error)
	List(page, limit int) ([]*model.Recipe, int, error)
	Search(query string, page, limit int) ([]*model.Recipe, int, error)
	Random() (*model.Recipe, error)
	Update(recipe *model.Recipe) error
	Delete(id string) error
	ExistsByName(name string) (bool, error)
	ListAll() ([]*model.Recipe, error)
	ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)
	ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error)
}

type UserStore interface {
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Count() (int, error)
	List() ([]*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	GetByEmail(email string) (*model.User, error)
}
