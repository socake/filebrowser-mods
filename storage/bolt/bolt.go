package bolt

import (
	"errors"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/auth"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)
	roleStore := users.NewRolesStorage(rolesBackend{db: db})

	err := save(db, "version", 2)
	if err != nil {
		return nil, err
	}

	if err := seedPresetRoles(roleStore); err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		Roles:    roleStore,
	}, nil
}

// seedPresetRoles inserts the built-in preset roles the first time the database
// is initialized (i.e. when no roles exist yet).
func seedPresetRoles(roleStore *users.RolesStorage) error {
	existing, err := roleStore.Gets()
	if err != nil && !errors.Is(err, fberrors.ErrNotExist) {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	for _, role := range users.PresetRoles() {
		if err := roleStore.Save(role); err != nil {
			return err
		}
	}
	return nil
}
