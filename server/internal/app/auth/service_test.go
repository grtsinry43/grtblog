package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

type fakeIdentityRepo struct {
	nextID          int64
	usersByID       map[int64]*identity.User
	usersByUsername map[string]*identity.User
	usersByEmail    map[string]*identity.User
}

func newFakeIdentityRepo(seed ...*identity.User) *fakeIdentityRepo {
	repo := &fakeIdentityRepo{
		nextID:          1,
		usersByID:       make(map[int64]*identity.User),
		usersByUsername: make(map[string]*identity.User),
		usersByEmail:    make(map[string]*identity.User),
	}
	for _, u := range seed {
		if u == nil {
			continue
		}
		user := *u
		if user.ID <= 0 {
			user.ID = repo.nextID
			repo.nextID++
		}
		if user.ID >= repo.nextID {
			repo.nextID = user.ID + 1
		}
		repo.usersByID[user.ID] = &user
		repo.usersByUsername[user.Username] = &user
		if user.Email != "" {
			repo.usersByEmail[user.Email] = &user
		}
	}
	return repo
}

func (f *fakeIdentityRepo) Create(_ context.Context, user *identity.User) error {
	if user == nil {
		return errors.New("nil user")
	}
	if user.Username == "" {
		return identity.ErrUserExists
	}
	if _, ok := f.usersByUsername[user.Username]; ok {
		return identity.ErrUserExists
	}
	if user.Email != "" {
		if _, ok := f.usersByEmail[user.Email]; ok {
			return identity.ErrUserExists
		}
	}
	created := *user
	created.ID = f.nextID
	f.nextID++
	now := time.Now()
	created.CreatedAt = now
	created.UpdatedAt = now
	f.usersByID[created.ID] = &created
	f.usersByUsername[created.Username] = &created
	if created.Email != "" {
		f.usersByEmail[created.Email] = &created
	}
	*user = created
	return nil
}

func (f *fakeIdentityRepo) FindByID(_ context.Context, id int64) (*identity.User, error) {
	if u, ok := f.usersByID[id]; ok {
		dup := *u
		return &dup, nil
	}
	return nil, identity.ErrUserNotFound
}

func (f *fakeIdentityRepo) FindByUsername(_ context.Context, username string) (*identity.User, error) {
	if u, ok := f.usersByUsername[username]; ok {
		dup := *u
		return &dup, nil
	}
	return nil, identity.ErrUserNotFound
}

func (f *fakeIdentityRepo) FindByEmail(_ context.Context, email string) (*identity.User, error) {
	if u, ok := f.usersByEmail[email]; ok {
		dup := *u
		return &dup, nil
	}
	return nil, identity.ErrUserNotFound
}

func (f *fakeIdentityRepo) FindByCredential(ctx context.Context, credential string) (*identity.User, error) {
	if u, err := f.FindByUsername(ctx, credential); err == nil {
		return u, nil
	}
	if u, err := f.FindByEmail(ctx, credential); err == nil {
		return u, nil
	}
	return nil, identity.ErrInvalidCredentials
}

func (f *fakeIdentityRepo) UpdateProfile(_ context.Context, userID int64, nickname, avatar, email string) (*identity.User, error) {
	u, ok := f.usersByID[userID]
	if !ok {
		return nil, identity.ErrUserNotFound
	}
	updated := *u
	if nickname != "" {
		updated.Nickname = nickname
	}
	if avatar != "" {
		updated.Avatar = avatar
	}
	if email != "" {
		updated.Email = email
	}
	updated.UpdatedAt = time.Now()
	f.usersByID[userID] = &updated
	f.usersByUsername[updated.Username] = &updated
	if updated.Email != "" {
		f.usersByEmail[updated.Email] = &updated
	}
	return &updated, nil
}

func (f *fakeIdentityRepo) UpdatePassword(_ context.Context, userID int64, hashed string) error {
	u, ok := f.usersByID[userID]
	if !ok {
		return identity.ErrUserNotFound
	}
	updated := *u
	updated.Password = hashed
	updated.UpdatedAt = time.Now()
	f.usersByID[userID] = &updated
	return nil
}

func (f *fakeIdentityRepo) ListOAuthBindings(_ context.Context, _ int64) ([]identity.UserOAuthBinding, error) {
	return nil, nil
}

func (f *fakeIdentityRepo) FindByOAuth(_ context.Context, _, _ string) (*identity.User, error) {
	return nil, identity.ErrUserNotFound
}

func (f *fakeIdentityRepo) BindOAuth(_ context.Context, _ identity.UserOAuth) error { return nil }

func (f *fakeIdentityRepo) BindOAuthByProvider(_ context.Context, _ identity.UserOAuth) error {
	return nil
}

func (f *fakeIdentityRepo) UnbindOAuth(_ context.Context, _ int64, _ string) error { return nil }

func (f *fakeIdentityRepo) CountUsers(_ context.Context) (int64, error) {
	return int64(len(f.usersByID)), nil
}

func (f *fakeIdentityRepo) ListAdmins(_ context.Context) ([]identity.User, error) {
	result := make([]identity.User, 0)
	for _, u := range f.usersByID {
		if u.IsAdmin {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (f *fakeIdentityRepo) CountActiveAdmins(_ context.Context) (int64, error) {
	var total int64
	for _, u := range f.usersByID {
		if u.IsAdmin && u.IsActive {
			total++
		}
	}
	return total, nil
}

func (f *fakeIdentityRepo) ListUsers(_ context.Context, _ identity.UserListOptions) ([]identity.User, int64, error) {
	return nil, int64(len(f.usersByID)), nil
}

func (f *fakeIdentityRepo) UpdateAdminUser(_ context.Context, userID int64, nickname, email string, isActive, isAdmin bool) (*identity.User, error) {
	u, ok := f.usersByID[userID]
	if !ok {
		return nil, identity.ErrUserNotFound
	}
	updated := *u
	if nickname != "" {
		updated.Nickname = nickname
	}
	if email != "" {
		updated.Email = email
	}
	updated.IsActive = isActive
	updated.IsAdmin = isAdmin
	updated.UpdatedAt = time.Now()
	f.usersByID[userID] = &updated
	f.usersByUsername[updated.Username] = &updated
	if updated.Email != "" {
		f.usersByEmail[updated.Email] = &updated
	}
	return &updated, nil
}

func TestRegisterOAuthUserDoesNotReuseExistingAdminOnUsernameConflict(t *testing.T) {
	repo := newFakeIdentityRepo(&identity.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		IsAdmin:  true,
		IsActive: true,
	})
	svc := &Service{users: repo}

	user, err := svc.registerOAuthUser(context.Background(), &ExternalIdentity{
		Provider:   "github",
		ProviderID: "123",
		Username:   "admin",
		Name:       "Another User",
	})
	if err != nil {
		t.Fatalf("registerOAuthUser returned error: %v", err)
	}
	if user.ID == 1 {
		t.Fatalf("expected new user id, got admin id=%d", user.ID)
	}
	if user.Username == "admin" {
		t.Fatalf("expected conflict-safe username, got %q", user.Username)
	}
	if user.IsAdmin {
		t.Fatalf("expected oauth user not admin")
	}
}

func TestFirstStringSupportsNumericID(t *testing.T) {
	raw := map[string]any{
		"id": float64(10086),
	}
	got := firstString(raw, "id")
	if got != "10086" {
		t.Fatalf("expected numeric id to be converted, got %q", got)
	}
}

func TestFetchExternalIdentityRejectsEmptyIdentity(t *testing.T) {
	_, err := fetchExternalIdentity(context.Background(), &identity.OAuthProvider{
		ProviderKey: "dummy",
	}, &oauth2.Token{})
	if !errors.Is(err, ErrInvalidOAuthIdentity) {
		t.Fatalf("expected ErrInvalidOAuthIdentity, got %v", err)
	}
}
