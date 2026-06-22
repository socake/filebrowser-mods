package fbhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

func getRoleID(r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	i, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func decodeRole(r *http.Request) (*users.Role, error) {
	if r.Body == nil {
		return nil, fberrors.ErrEmptyRequest
	}
	role := &users.Role{}
	if err := json.NewDecoder(r.Body).Decode(role); err != nil {
		return nil, err
	}
	return role, nil
}

var rolesGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	roles, err := d.store.Roles.Gets()
	if err != nil && !errors.Is(err, fberrors.ErrNotExist) {
		return http.StatusInternalServerError, err
	}

	sort.Slice(roles, func(i, j int) bool {
		if roles[i].Sort != roles[j].Sort {
			return roles[i].Sort < roles[j].Sort
		}
		return roles[i].ID < roles[j].ID
	})

	return renderJSON(w, r, roles)
})

var roleGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getRoleID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	role, err := d.store.Roles.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, role)
})

var rolePostHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	role, err := decodeRole(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if role.Name == "" {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}

	// Creating a preset role through the API is not allowed; presets are only
	// seeded internally.
	role.ID = 0
	role.IsPreset = false

	if err := d.store.Roles.Save(role); err != nil {
		return errToStatus(err), err
	}

	w.Header().Set("Location", "/settings/roles/"+strconv.FormatUint(uint64(role.ID), 10))
	return http.StatusCreated, nil
})

var rolePutHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getRoleID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	role, err := decodeRole(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	existing, err := d.store.Roles.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	role.ID = id
	// Preset flag is immutable; preserve whatever it was.
	role.IsPreset = existing.IsPreset

	if err := d.store.Roles.Save(role); err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

var roleDeleteHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id, err := getRoleID(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	existing, err := d.store.Roles.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	if existing.IsPreset {
		return http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	if err := d.store.Roles.Delete(id); err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})
