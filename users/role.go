package users

// Role is a permission template. Assigning a role to a user copies the role's
// Permissions into the user's Perm field. The role itself never participates in
// authorization decisions; user.Perm remains the single source of truth.
type Role struct {
	ID          uint        `storm:"id,increment" json:"id"`
	Name        string      `storm:"unique" json:"name"`
	Description string      `json:"description"`
	Permissions Permissions `json:"permissions"`
	Scope       string      `json:"scope"`
	// IsPreset marks the four built-in roles. Preset roles cannot be deleted.
	IsPreset bool `json:"isPreset"`
	// Sort is an optional ordering hint for the UI.
	Sort int `json:"sort"`
}

// RolesStoreBackend is the interface to implement for a roles storage.
type RolesStoreBackend interface {
	GetBy(interface{}) (*Role, error)
	Gets() ([]*Role, error)
	Save(*Role) error
	DeleteByID(uint) error
}

// RolesStorage is a roles storage.
type RolesStorage struct {
	back RolesStoreBackend
}

// NewRolesStorage creates a roles storage from a backend.
func NewRolesStorage(back RolesStoreBackend) *RolesStorage {
	return &RolesStorage{back: back}
}

// Get fetches a role by its uint ID or string name.
func (s *RolesStorage) Get(id interface{}) (*Role, error) {
	return s.back.GetBy(id)
}

// GetByID fetches a role by its ID.
func (s *RolesStorage) GetByID(id uint) (*Role, error) {
	return s.back.GetBy(id)
}

// Gets returns all roles.
func (s *RolesStorage) Gets() ([]*Role, error) {
	return s.back.Gets()
}

// Save saves a role (creates or updates).
func (s *RolesStorage) Save(role *Role) error {
	return s.back.Save(role)
}

// Delete removes a role by its ID.
func (s *RolesStorage) Delete(id uint) error {
	return s.back.DeleteByID(id)
}

// PresetRoles returns the four built-in roles used to seed an empty database.
func PresetRoles() []*Role {
	return []*Role{
		{
			Name:        "管理员",
			Description: "拥有全部权限",
			Scope:       "/",
			IsPreset:    true,
			Sort:        1,
			Permissions: Permissions{
				Admin:    true,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
		{
			Name:        "编辑者",
			Description: "可创建、重命名、修改、删除、分享、下载",
			Scope:       "/",
			IsPreset:    true,
			Sort:        2,
			Permissions: Permissions{
				Admin:    false,
				Execute:  false,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
		{
			Name:        "查看者",
			Description: "仅可下载",
			Scope:       "/",
			IsPreset:    true,
			Sort:        3,
			Permissions: Permissions{
				Download: true,
			},
		},
		{
			Name:        "访客",
			Description: "无任何权限",
			Scope:       "/",
			IsPreset:    true,
			Sort:        4,
			Permissions: Permissions{},
		},
	}
}
