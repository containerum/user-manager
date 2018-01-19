package resource

import (
	"time"

	dec "github.com/shopspring/decimal"
)

type User struct {
	ID        string      `json:"user_id,omitempty"`
	Login     string      `json:"login,omitempty"`
	Country   int         `json:"country,omitempty"`
	Balance   dec.Decimal `json:"balance,omitempty"`
	BillingID string      `json:"billing_id,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
}

type Tariff struct {
	ID        string      `json:"id,omitempty"`
	Label     string      `json:"label,omitempty"`
	Type      string      `json:"type,omitempty"`
	Price     dec.Decimal `json:"price,omitempty"`
	IsActive  bool        `json:"is_active,omitempty"`
	IsPublic  bool        `json:"is_public,omitempty"`
	BillingID string      `json:"billing_id,omitempty"`
}

type NamespaceTariff struct {
	ID          string    `json:"id,omitempty"`
	TariffID    string    `json:"tariff_id,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`

	CpuLimit         int         `json:"cpu_limit,omitempty"`
	MemoryLimit      int         `json:"memory_limit,omitempty"`
	Traffic          int         `json:"traffic,omitempty"`
	TrafficPrice     dec.Decimal `json:"traffic_price,omitempty"`
	ExternalServices int         `json:"external_services,omitempty"`
	InternalServices int         `json:"internal_services,omitempty"`

	VV *VolumeTariff `json:"VV,omitempty"`

	IsActive bool        `json:"is_active,omitempty"`
	IsPublic bool        `json:"is_public,omitempty"`
	Price    dec.Decimal `json:"price,omitempty"`
}

type VolumeTariff struct {
	ID          string    `json:"id,omitempty"`
	TariffID    string    `json:"tariff_id,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`

	StorageLimit  int  `json:"storage_limit,omitempty"`
	ReplicasLimit int  `json:"replicas_limit,omitempty"`
	IsPersistent  bool `json:"is_persistent,omitempty"`

	IsActive bool        `json:"is_active,omitempty"`
	IsPublic bool        `json:"is_public,omitempty"`
	Price    dec.Decimal `json:"price,omitempty"`
}

type Resource struct {
	ResourceID string `json:"resource_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	TariffID   string `json:"tariff_id,omitempty"`
	BillingID  string `json:"billing_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

type AccessLevel string // constants AOwner, etc.

const (
	AOwner      AccessLevel = "owner"
	AWrite      AccessLevel = "write"
	AReadDelete AccessLevel = "readdelete"
	ARead       AccessLevel = "read"
	ANone       AccessLevel = "none"
)

type AccessRecord struct {
	UserID           string      `json:"user_id,omitempty"`
	Access           AccessLevel `json:"access_level,omitempty"`
	Limited          bool        `json:"limited,omitempty"`
	NewAccess        AccessLevel `json:"new_access_level,omitempty"`
	AccessChangeTime time.Time   `json:"access_level_change_time,omitempty"`
}

type Volume struct {
	ID               string      `json:"id,omitempty"`
	CreateTime       time.Time   `json:"create_time,omitempty"`
	Deleted          bool        `json:"deleted,omitempty"`
	DeleteTime       time.Time   `json:"delete_time,omitempty"`
	UserID           string      `json:"user_id,omitempty"`
	TariffID         string      `json:"tariff_id,omitempty"`
	Label            string      `json:"label,omitempty"`
	Access           AccessLevel `json:"access,omitempty"`
	AccessChangeTime time.Time   `json:"access_change_time,omitempty"`
	Limited          bool        `json:"limited,omitempty"`
	NewAccess        AccessLevel `json:"new_access,omitempty"`

	Storage    int  `json:"storage,omitempty"`
	Replicas   int  `json:"replicas,omitempty"`
	Persistent bool `json:"persistent,omitempty"`

	Users []AccessRecord `json:"users,omitempty"`
}

type Namespace struct {
	ID               string      `json:"id,omitempty"`
	CreateTime       time.Time   `json:"create_time,omitempty"`
	Deleted          bool        `json:"deleted,omitempty"`
	DeleteTime       time.Time   `json:"delete_time,omitempty"`
	UserID           string      `json:"user_id,omitempty"`
	TariffID         string      `json:"tariff_id,omitempty"`
	Label            string      `json:"label,omitempty"`
	Access           AccessLevel `json:"access,omitempty"`
	AccessChangeTime time.Time   `json:"access_change_time,omitempty"`
	Limited          bool        `json:"limited,omitempty"`
	NewAccess        AccessLevel `json:"new_access,omitempty"`

	RAM           int `json:"ram,omitempty"`
	CPU           int `json:"cpu,omitempty"`
	MaxExtService int `json:"max_ext_service,omitempty"`
	MaxIntService int `json:"max_int_service,omitempty"`
	MaxTraffic    int `json:"max_traffic,omitempty"`

	Volumes []Volume `json:"volumes,omitempty"`

	Users []AccessRecord `json:"users,omitempty"`
}
