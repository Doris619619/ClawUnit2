// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// InstancesDao is the data access object for the table instances.
type InstancesDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  InstancesColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// InstancesColumns defines and stores column names for the table instances.
type InstancesColumns struct {
	Id                  string //
	OwnerUpn            string //
	Name                string //
	Description         string //
	Status              string //
	Image               string //
	StorageClass        string //
	MountPath           string //
	PodName             string //
	PodNamespace        string //
	PodIp               string //
	AccessToken         string //
	ApiKeyHash          string //
	CpuCores            string //
	MemoryGb            string //
	DiskGb              string //
	GpuCount            string //
	ContainerPort       string //
	GpuEnabled          string //
	CreatedAt           string //
	UpdatedAt           string //
	StartedAt           string //
	StoppedAt           string //
	ApiMode             string //
	ApiKey              string //
	BaseUrl             string //
	GatewayToken        string //
	Provider            string //
	ModelId             string //
	AllowPrivateNetwork string //
	PvcName             string //
}

// instancesColumns holds the columns for the table instances.
var instancesColumns = InstancesColumns{
	Id:                  "id",
	OwnerUpn:            "owner_upn",
	Name:                "name",
	Description:         "description",
	Status:              "status",
	Image:               "image",
	StorageClass:        "storage_class",
	MountPath:           "mount_path",
	PodName:             "pod_name",
	PodNamespace:        "pod_namespace",
	PodIp:               "pod_ip",
	AccessToken:         "access_token",
	ApiKeyHash:          "api_key_hash",
	CpuCores:            "cpu_cores",
	MemoryGb:            "memory_gb",
	DiskGb:              "disk_gb",
	GpuCount:            "gpu_count",
	ContainerPort:       "container_port",
	GpuEnabled:          "gpu_enabled",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	StartedAt:           "started_at",
	StoppedAt:           "stopped_at",
	ApiMode:             "api_mode",
	ApiKey:              "api_key",
	BaseUrl:             "base_url",
	GatewayToken:        "gateway_token",
	Provider:            "provider",
	ModelId:             "model_id",
	AllowPrivateNetwork: "allow_private_network",
	PvcName:             "pvc_name",
}

// NewInstancesDao creates and returns a new DAO object for table data access.
func NewInstancesDao(handlers ...gdb.ModelHandler) *InstancesDao {
	return &InstancesDao{
		group:    "default",
		table:    "instances",
		columns:  instancesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *InstancesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *InstancesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *InstancesDao) Columns() InstancesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *InstancesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *InstancesDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *InstancesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
