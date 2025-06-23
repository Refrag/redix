//go:build linux || darwin

package filesystem

import "github.com/Refrag/redix/internals/datastore/contract"

// Global consts.
const (
	Name = "filesystem"
)

// init registers the engine.
func init() {
	contract.Register(Name, &Engine{})
}
