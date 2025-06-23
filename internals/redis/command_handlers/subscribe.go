package commandhandlers

import commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"

func Subscribe(c *commandutilities.Context) {
	if c.Argc != 1 {
		c.Conn.WriteError("ERR wrong number of arguments for 'subscribe' command")
		return
	}

	conn := c.Conn.Detach()
	defer conn.Close()

	conn.WriteArray(3)
	conn.WriteBulkString("subscribe")
	conn.WriteBulk(c.Argv[0])
	conn.WriteInt(1)
	conn.Flush()

	err := c.Engine.Subscribe(c.AbsoluteKeyPath([]byte("redix"), c.Argv[0]), func(msg []byte) error {
		conn.WriteArray(3)
		conn.WriteBulkString("message")
		conn.WriteBulk(c.Argv[0])
		conn.WriteBulk(msg)
		conn.Flush()
		return nil
	})

	if err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}
}
