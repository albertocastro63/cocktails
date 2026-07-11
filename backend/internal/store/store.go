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
	SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error)
	SearchByBaseSpirit(baseSpirit string, page, limit int) ([]*model.Recipe, int, error)
	SearchByBaseSpiritAndIngredients(baseSpirit string, ingredients []string, page, limit int) ([]*model.Recipe, int, error)
	Random() (*model.Recipe, error)
	Update(recipe *model.Recipe) error
	Delete(id string) error
	// SetRelated sets recipeID's related set to relatedIDs (normalized: deduped,
	// self dropped, non-existent dropped) and reconciles the symmetric reverse
	// relation on each counterpart. Non-transactional: returns any write error.
	SetRelated(recipeID string, relatedIDs []string) error
	ExistsByName(name string) (bool, error)
	ListAll() ([]*model.Recipe, error)
	ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)
	ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error)
}

type FavoriteStore interface {
	Add(userID, recipeID string) error
	Remove(userID, recipeID string) error
	IsFavorite(userID, recipeID string) (bool, error)
	ListByUser(userID string) ([]*model.Favorite, error)
	CountByRecipe(recipeID string) (int, error)
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
