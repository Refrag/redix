package commandhandlers

import (
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

// SMembers returns all members of a set.
// SMEMBERS key
func SMembers(c *commandutilities.Context) {
	if c.Argc < 1 {
		c.Conn.WriteError("ERR wrong number of arguments for 'smembers' command")
		return
	}

	key := c.AbsoluteKeyPath(c.Argv[0])

	members, err := c.Engine.SMembers(key)
	if err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	c.Conn.WriteArray(len(members))
	for _, member := range members {
		c.Conn.WriteBulk(member)
	}
}
