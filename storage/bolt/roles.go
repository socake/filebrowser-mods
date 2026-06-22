package bolt

import (
	"errors"

	"github.com/asdine/storm/v3"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

type rolesBackend struct {
	db *storm.DB
}

func (st rolesBackend) GetBy(i interface{}) (*users.Role, error) {
	role := &users.Role{}

	var arg string
	switch i.(type) {
	case uint:
		arg = "ID"
	case string:
		arg = "Name"
	default:
		return nil, fberrors.ErrInvalidDataType
	}

	err := st.db.One(arg, i, role)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, fberrors.ErrNotExist
		}
		return nil, err
	}

	return role, nil
}

func (st rolesBackend) Gets() ([]*users.Role, error) {
	var roles []*users.Role
	err := st.db.All(&roles)
	if errors.Is(err, storm.ErrNotFound) {
		return roles, fberrors.ErrNotExist
	}

	return roles, err
}

func (st rolesBackend) Save(role *users.Role) error {
	err := st.db.Save(role)
	if errors.Is(err, storm.ErrAlreadyExists) {
		return fberrors.ErrExist
	}
	return err
}

func (st rolesBackend) DeleteByID(id uint) error {
	return st.db.DeleteStruct(&users.Role{ID: id})
}
