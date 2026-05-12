package handler_test

import (
	"errors"
	"strings"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
)

var errGeneric = errors.New("store error")

// stubRecipeStore

type stubRecipeStore struct {
	recipes map[string]*model.Recipe
	err     error
}

func newStubRecipeStore(rs ...*model.Recipe) *stubRecipeStore {
	s := &stubRecipeStore{recipes: map[string]*model.Recipe{}}
	for _, r := range rs {
		s.recipes[r.ID] = r
	}
	return s
}

func (s *stubRecipeStore) Create(r *model.Recipe) error {
	if s.err != nil {
		return s.err
	}
	s.recipes[r.ID] = r
	return nil
}

func (s *stubRecipeStore) GetByID(id string) (*model.Recipe, error) {
	if s.err != nil {
		return nil, s.err
	}
	r, ok := s.recipes[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (s *stubRecipeStore) List(page, limit int) ([]*model.Recipe, int, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	all := make([]*model.Recipe, 0, len(s.recipes))
	for _, r := range s.recipes {
		all = append(all, r)
	}
	return all, len(all), nil
}

func (s *stubRecipeStore) Search(q string, page, limit int) ([]*model.Recipe, int, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	var matches []*model.Recipe
	for _, r := range s.recipes {
		if strings.Contains(strings.ToLower(r.Name), strings.ToLower(q)) {
			matches = append(matches, r)
		}
	}
	return matches, len(matches), nil
}

func (s *stubRecipeStore) Random() (*model.Recipe, error) {
	if s.err != nil {
		return nil, s.err
	}
	for _, r := range s.recipes {
		return r, nil
	}
	return nil, nil
}

func (s *stubRecipeStore) Update(r *model.Recipe) error {
	if s.err != nil {
		return s.err
	}
	s.recipes[r.ID] = r
	return nil
}

func (s *stubRecipeStore) Delete(id string) error {
	if s.err != nil {
		return s.err
	}
	delete(s.recipes, id)
	return nil
}

func (s *stubRecipeStore) ExistsByName(name string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	for _, r := range s.recipes {
		if strings.EqualFold(r.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

// compile-time interface checks
var _ store.RecipeStore = (*stubRecipeStore)(nil)
var _ store.UserStore = (*stubUserStore)(nil)

type stubUserStore struct {
	byUsername map[string]*model.User
	byID       map[string]*model.User
	createErr  error
}

func newStubUserStore(users ...*model.User) *stubUserStore {
	s := &stubUserStore{
		byUsername: map[string]*model.User{},
		byID:       map[string]*model.User{},
	}
	for _, u := range users {
		s.byUsername[u.Username] = u
		s.byID[u.ID] = u
	}
	return s
}

func (s *stubUserStore) Create(u *model.User) error {
	if s.createErr != nil {
		return s.createErr
	}
	if _, exists := s.byUsername[u.Username]; exists {
		return store.ErrDuplicate
	}
	s.byUsername[u.Username] = u
	s.byID[u.ID] = u
	return nil
}

func (s *stubUserStore) GetByUsername(name string) (*model.User, error) {
	u, ok := s.byUsername[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (s *stubUserStore) GetByID(id string) (*model.User, error) {
	u, ok := s.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (s *stubUserStore) Count() (int, error) {
	return len(s.byUsername), nil
}

func (s *stubUserStore) List() ([]*model.User, error) {
	var users []*model.User
	for _, u := range s.byID {
		if !u.IsAdmin {
			users = append(users, u)
		}
	}
	if users == nil {
		users = []*model.User{}
	}
	return users, nil
}

func (s *stubUserStore) Update(u *model.User) error {
	if _, exists := s.byID[u.ID]; !exists {
		return errors.New("not found")
	}
	s.byID[u.ID] = u
	s.byUsername[u.Username] = u
	return nil
}

func (s *stubUserStore) Delete(id string) error {
	u, exists := s.byID[id]
	if !exists {
		return errors.New("not found")
	}
	delete(s.byID, id)
	delete(s.byUsername, u.Username)
	return nil
}

func (s *stubUserStore) GetByEmail(email string) (*model.User, error) {
	for _, u := range s.byID {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}
