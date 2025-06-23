package commandhandlers

import commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"

func Quit(c *commandutilities.Context) {
	c.Conn.WriteString("OK")
	c.Conn.Close()
}
