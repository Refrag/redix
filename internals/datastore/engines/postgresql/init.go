package postgresql

import "github.com/Refrag/redix/internals/datastore/contract"

// Global consts.
const (
	Name = "postgresql"
)

// Register registers the PostgreSQL engine.
func Register() {
	contract.Register(Name, &Engine{})
}
