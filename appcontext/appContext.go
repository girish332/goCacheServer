package appcontext

import (
	"cacheServer/db"
)

// Context struct contains database client and db timeout.
type Context struct {
	DatabaseClient db.DatabaseClient
	DBTimeout      int
}

// NewContext constructor for appcontext struct.
func NewContext(db db.DatabaseClient, timeout int) *Context {
	return &Context{DatabaseClient: db, DBTimeout: timeout}
}
