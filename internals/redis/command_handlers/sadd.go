package commandhandlers

import (
	"time"

	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

const (
	SAddArgumentsMinCount = 2
)

func parseSAddOptions(c *commandutilities.Context, ttl *time.Duration) error {
	// Redis SADD doesn't have additional options like EX, but we might extend it later
	// For now, keep it simple and let TTL be handled by EXPIRE command separately
	*ttl = 0
	return nil
}

// SADD key member [member ...]
func SAdd(c *commandutilities.Context) {
	if c.Argc < SAddArgumentsMinCount {
		c.Conn.WriteError("ERR wrong number of arguments for 'sadd' command")
		return
	}

	key := c.AbsoluteKeyPath(c.Argv[0])
	members := make([][]byte, 0, c.Argc-1)

	// Collect all members to add (skip duplicates in the same command)
	memberSet := make(map[string]bool)
	for i := 1; i < c.Argc; i++ {
		member := string(c.Argv[i])
		if !memberSet[member] {
			memberSet[member] = true
			members = append(members, c.Argv[i])
		}
	}

	if len(members) == 0 {
		c.Conn.WriteInt(0)
		return
	}

	var ttl time.Duration
	if err := parseSAddOptions(c, &ttl); err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	addedCount, err := c.Engine.SAdd(key, members, ttl)
	if err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	c.Conn.WriteInt(addedCount)
}
