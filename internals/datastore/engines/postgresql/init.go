package postgresql

import "github.com/Refrag/redix/internals/datastore/contract"

// Global consts
const (
	Name = "postgresql"
)

func init() {
	contract.Register(Name, &Engine{})
}
