package commandhandlers

import (
	"strings"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func HGetAll(c *commandutilities.Context) {
	prefix := []byte("")

	if c.Argc > 0 {
		prefix = c.Argv[0]
	}

	result := map[string]string{}

	err := c.Engine.Iterate(&contract.IteratorOpts{
		Prefix: c.AbsoluteKeyPath(prefix),
		Callback: func(ro *contract.ReadOutput) error {
			endKey := strings.TrimPrefix(string(ro.Key), string(c.AbsoluteKeyPath(prefix)))
			result[endKey] = string(ro.Value)
			return nil
		},
	})

	if err != nil && err != contract.ErrStopIterator {
		c.Conn.WriteError("ERR " + err.Error())
	}

	c.Conn.WriteAny(result)
}
