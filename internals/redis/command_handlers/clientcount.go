package commandhandlers

import (
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func ClientCount(c *commandutilities.Context) {
	c.Conn.WriteAny(commandutilities.GetConnCounter())
}
