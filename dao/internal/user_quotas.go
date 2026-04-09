// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserQuotasDao is the data access object for the table user_quotas.
type UserQuotasDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserQuotasColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserQuotasColumns defines and stores column names for the table user_quotas.
type UserQuotasColumns struct {
	Id           string //
	OwnerUpn     string //
	MaxInstances string //
	MaxCpuCores  string //
	MaxMemoryGb  string //
	MaxStorageGb string //
	MaxGpuCount  string //
	CreatedAt    string //
	UpdatedAt    string //
}

// userQuotasColumns holds the columns for the table user_quotas.
var userQuotasColumns = UserQuotasColumns{
	Id:           "id",
	OwnerUpn:     "owner_upn",
	MaxInstances: "max_instances",
	MaxCpuCores:  "max_cpu_cores",
	MaxMemoryGb:  "max_memory_gb",
	MaxStorageGb: "max_storage_gb",
	MaxGpuCount:  "max_gpu_count",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewUserQuotasDao creates and returns a new DAO object for table data access.
func NewUserQuotasDao(handlers ...gdb.ModelHandler) *UserQuotasDao {
	return &UserQuotasDao{
		group:    "default",
		table:    "user_quotas",
		columns:  userQuotasColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserQuotasDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserQuotasDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserQuotasDao) Columns() UserQuotasColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserQuotasDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserQuotasDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserQuotasDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
