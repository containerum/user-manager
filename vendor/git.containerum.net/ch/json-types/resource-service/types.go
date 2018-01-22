package resource

import (
	"time"

	dec "github.com/shopspring/decimal"
)

type User struct {
	ID        string      `json:"user_id"`
	Login     string      `json:"login"`
	Country   int         `json:"country"`
	Balance   dec.Decimal `json:"balance"`
	BillingID string      `json:"billing_id"`
	CreatedAt time.Time   `json:"created_at"`
}

type Tariff struct {
	ID        string      `json:"id"`
	Label     string      `json:"label"`
	Type      string      `json:"type"`
	Price     dec.Decimal `json:"price"`
	IsActive  bool        `json:"is_active"`
	IsPublic  bool        `json:"is_public"`
	BillingID string      `json:"billing_id"`
}

type NamespaceTariff struct {
	ID          string    `json:"id"`
	TariffID    string    `json:"tariff_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	CpuLimit         int         `json:"cpu_limit"`
	MemoryLimit      int         `json:"memory_limit"`
	Traffic          int         `json:"traffic"`
	TrafficPrice     dec.Decimal `json:"traffic_price"`
	ExternalServices int         `json:"external_services"`
	InternalServices int         `json:"internal_services"`

	VV *VolumeTariff `json:"VV,omitempty"`

	IsActive bool        `json:"is_active"`
	IsPublic bool        `json:"is_public"`
	Price    dec.Decimal `json:"price"`
}

type VolumeTariff struct {
	ID          string    `json:"id"`
	TariffID    string    `json:"tariff_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	StorageLimit  int  `json:"storage_limit"`
	ReplicasLimit int  `json:"replicas_limit"`
	IsPersistent  bool `json:"is_persistent"`

	IsActive bool        `json:"is_active"`
	IsPublic bool        `json:"is_public"`
	Price    dec.Decimal `json:"price"`
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
	UserID           string      `json:"user_id"`
	Access           AccessLevel `json:"access_level"`
	Limited          bool        `json:"limited"`
	NewAccess        AccessLevel `json:"new_access_level"`
	AccessChangeTime time.Time   `json:"access_level_change_time,omitempty"`
}

type Resource struct {
	ID         string    `json:"id"`
	CreateTime time.Time `json:"create_time"`
	Deleted    bool      `json:"deleted,omitempty"`
	DeleteTime time.Time `json:"delete_time,omitempty"`
	TariffID   string    `json:"tariff_id"`
	Label      string    `json:"label,omitempty"`
	AccessRecord

	Users []AccessRecord `json:"users,omitempty"`
}

type Volume struct {
	Resource

	Storage    int  `json:"storage"`
	Replicas   int  `json:"replicas"`
	Persistent bool `json:"persistent"`
	IsActive   bool `json:"active"`
}

type Namespace struct {
	Resource

	RAM           int `json:"ram"`
	CPU           int `json:"cpu"`
	MaxExtService int `json:"max_ext_service,omitempty"`
	MaxIntService int `json:"max_int_service,omitempty"`
	MaxTraffic    int `json:"max_traffic,omitempty"`

	Volumes []Volume `json:"volumes,omitempty"`
}

type ListAllResourcesParams struct {
	Page    int       `form:"page" binding:"gt=0"`
	PerPage int       `form:"per_page" binding:"gt=0"`
	Order   string    `form:"order" binding:"eq=ASC|eq=DESC"`
	After   time.Time `form:"after" binding:"omitempty,lt"`
	Deleted bool      `form:"deleted" binding:"omitempty"`
	Limited bool      `form:"limited" binding:"omitempty"`
}
