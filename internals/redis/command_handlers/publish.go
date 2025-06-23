package commandhandlers

import commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"

func Publish(c *commandutilities.Context) {
	if c.Argc < 2 {
		c.Conn.WriteError("ERR wrong number of arguments for 'publish' command")
		return
	}

	if err := c.Engine.Publish(c.AbsoluteKeyPath([]byte("redix"), c.Argv[0]), c.Argv[1]); err != nil {
		// FIX: Removed invalid format string placeholder that wasn't being used
		// Was: "ERR %s " + err.Error() - the %s had no corresponding argument
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	c.Conn.WriteInt(0)
}
