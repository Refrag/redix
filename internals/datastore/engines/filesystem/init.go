//go:build linux || darwin

package filesystem

import "github.com/Refrag/redix/internals/datastore/contract"

// Global consts.
const (
	Name = "filesystem"
)

// Register registers the filesystem engine.
func Register() {
	contract.Register(Name, &Engine{})
}
