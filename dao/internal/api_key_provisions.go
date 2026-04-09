// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ApiKeyProvisionsDao is the data access object for the table api_key_provisions.
type ApiKeyProvisionsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  ApiKeyProvisionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// ApiKeyProvisionsColumns defines and stores column names for the table api_key_provisions.
type ApiKeyProvisionsColumns struct {
	Id         string //
	InstanceId string //
	OwnerUpn   string //
	ApiKeyHash string //
	QuotaPool  string //
	CreatedAt  string //
	RevokedAt  string //
}

// apiKeyProvisionsColumns holds the columns for the table api_key_provisions.
var apiKeyProvisionsColumns = ApiKeyProvisionsColumns{
	Id:         "id",
	InstanceId: "instance_id",
	OwnerUpn:   "owner_upn",
	ApiKeyHash: "api_key_hash",
	QuotaPool:  "quota_pool",
	CreatedAt:  "created_at",
	RevokedAt:  "revoked_at",
}

// NewApiKeyProvisionsDao creates and returns a new DAO object for table data access.
func NewApiKeyProvisionsDao(handlers ...gdb.ModelHandler) *ApiKeyProvisionsDao {
	return &ApiKeyProvisionsDao{
		group:    "default",
		table:    "api_key_provisions",
		columns:  apiKeyProvisionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ApiKeyProvisionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ApiKeyProvisionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ApiKeyProvisionsDao) Columns() ApiKeyProvisionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ApiKeyProvisionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ApiKeyProvisionsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ApiKeyProvisionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
