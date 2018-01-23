package resource

import (
	"time"

	"git.containerum.net/ch/json-types/misc"
)

type Kind string // constants KindNamespace, KindVolume, ... It`s recommended to use strings.ToLower before comparsion

const (
	KindNamespace  Kind = "namespace"
	KindVolume          = "volume"
	KindExtService      = "extservice"
	KindIntService      = "intservice"
	KindDomain          = "domain"
)

type PermissionStatus string // constants PermissionStatusOwner, PermissionStatusRead

const (
	PermissionStatusOwner      PermissionStatus = "owner"
	PermissionStatusRead                        = "read"
	PermissionStatusWrite                       = "write"
	PermissionStatusReadDelete                  = "readdelete"
	PermissionStatusNone                        = "none"
)

type Resource struct {
	ID         string          `json:"id" db:"id"`
	CreateTime string          `json:"create_time" db:"create_time"`
	Deleted    bool            `json:"deleted" db:"deleted"`
	DeleteTime misc.PqNullTime `json:"delete_time,omitempty" db:"delete_time"`
	TariffID   string          `json:"tariff_id" db:"tariff_id"`
}

type Namespace struct {
	Resource

	RAM                 int `json:"ram" db:"ram"` // megabytes
	CPU                 int `json:"cpu" db:"cpu"`
	MaxExternalServices int `json:"max_external_services" db:"max_ext_services"`
	MaxIntServices      int `json:"max_internal_services" db:"max_int_services"`
	MaxTraffic          int `json:"max_traffic" db:"max_traffic"` // megabytes per month
}

type Volume struct {
	Resource

	Active     bool `json:"active" db:"active"`
	Capacity   int  `json:"capacity" db:"capacity"` // gigabytes
	Replicas   int  `json:"replicas" db:"replicas"`
	Persistent bool `json:"is_persistent" db:"is_persistent"`
}

type Deployment struct {
	ID          string          `json:"id" db:"id"`
	NamespaceID string          `json:"namespace_id" db:"ns_id"`
	Name        string          `json:"name" db:"name"`
	RAM         int             `json:"ram" db:"ram"`
	CPU         int             `json:"cpu" db:"cpu"`
	CreateTime  time.Time       `json:"create_time" db:"create_time"`
	Deleted     bool            `json:"deleted" db:"deleted"`
	DeleteTime  misc.PqNullTime `json:"delete_time,omitempty" db:"delete_time"`
}

type AccessRecord struct {
	UserID                string           `json:"user_id" db:"user_id"`
	AccessLevel           PermissionStatus `json:"access_level" db:"access_level"`
	Limited               bool             `json:"limited" db:"limited"`
	AccessLevelChangeTime time.Time        `json:"access_level_change_time" db:"access_level_change_time"`
}

type PermissionRecord struct {
	ID            string    `json:"id" db:"id"`
	Kind          Kind      `json:"kind" db:"kind"`
	ResourceID    string    `json:"resource_id" db:"resource_id"`
	ResourceLabel string    `json:"resource_label" db:"resource_label"`
	OwnerUserID   string    `json:"owner_user_id" db:"owner_user_id"`
	CreateTime    time.Time `json:"create_time" db:"create_time"`

	AccessRecord
	NewAccessLevel PermissionStatus `json:"new_access_level,omitempty" db:"new_access_level"`
}

// Types below is not for storing in db

type NamespaceWithVolumes struct {
	Namespace
	Volumes []Volume `json:"volumes"`
}

type NamespaceWithAccesses struct {
	Namespace
	Users []AccessRecord `json:"users"`
}

type VolumeWithAccess struct {
	Volume
	AccessRecord
}

type VolumeWithUserAccesses struct {
	VolumeWithAccess
	Users []AccessRecord `json:"users"`
}
