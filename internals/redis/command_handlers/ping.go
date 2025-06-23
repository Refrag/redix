package commandhandlers

import commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"

func Ping(c *commandutilities.Context) {
	c.Conn.WriteString("PONG")
}
