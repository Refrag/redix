package commandhandlers

import (
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func GetDel(c *commandutilities.Context) {
	if c.Argc != 1 {
		c.Conn.WriteError("Err invalid number of arguments specified")
		return
	}

	c.Argc++
	c.Argv = append(c.Argv, []byte("DELETE"))

	commandutilities.Call("get", c)
}
