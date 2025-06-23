package commandhandlers

import (
	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Ttl(c *commandutilities.Context) {
	if c.Argc < 1 {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	ret, err := c.Engine.Read(&contract.ReadInput{
		Key: c.AbsoluteKeyPath(c.Argv[0]),
	})

	if err != nil {
		c.Conn.WriteError("Err " + err.Error())
		return
	}

	if !ret.Exists {
		c.Conn.WriteBulkString("-2")
		return
	}

	if ret.TTL == 0 {
		c.Conn.WriteBulkString("-1")
		return
	}

	c.Conn.WriteAny(ret.TTL.Milliseconds())
}
