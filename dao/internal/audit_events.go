// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. Created at 2026-04-07 11:36:13
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AuditEventsDao is the data access object for the table audit_events.
type AuditEventsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AuditEventsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AuditEventsColumns defines and stores column names for the table audit_events.
type AuditEventsColumns struct {
	Id           string //
	ActorUpn     string //
	Action       string //
	ResourceType string //
	ResourceId   string //
	Details      string //
	IpAddress    string //
	CreatedAt    string //
}

// auditEventsColumns holds the columns for the table audit_events.
var auditEventsColumns = AuditEventsColumns{
	Id:           "id",
	ActorUpn:     "actor_upn",
	Action:       "action",
	ResourceType: "resource_type",
	ResourceId:   "resource_id",
	Details:      "details",
	IpAddress:    "ip_address",
	CreatedAt:    "created_at",
}

// NewAuditEventsDao creates and returns a new DAO object for table data access.
func NewAuditEventsDao(handlers ...gdb.ModelHandler) *AuditEventsDao {
	return &AuditEventsDao{
		group:    "default",
		table:    "audit_events",
		columns:  auditEventsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AuditEventsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AuditEventsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AuditEventsDao) Columns() AuditEventsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AuditEventsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AuditEventsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AuditEventsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
