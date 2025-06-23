package commandhandlers

import (
	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func FlushDb(c *commandutilities.Context) {
	_, err := c.Engine.Write(&contract.WriteInput{
		Key:   c.AbsoluteKeyPath(),
		Value: nil,
	})

	if err != nil {
		c.Conn.WriteError("Err " + err.Error())
		return
	}

	c.Conn.WriteString("OK")
}
