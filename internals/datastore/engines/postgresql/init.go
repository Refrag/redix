package postgresql

import "github.com/Refrag/redix/internals/datastore/contract"

// Global consts.
const (
	Name = "postgresql"
)

// init registers the engine.
func init() {
	contract.Register(Name, &Engine{})
}
