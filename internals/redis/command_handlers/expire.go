package commandhandlers

import (
	"strconv"
	"time"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

// Expire sets a timeout on key. After the timeout has expired, the key will automatically be deleted.
// EXPIRE key seconds
func Expire(c *commandutilities.Context) {
	if c.Argc < 2 {
		c.Conn.WriteError("ERR wrong number of arguments for 'expire' command")
		return
	}

	key := c.AbsoluteKeyPath(c.Argv[0])
	seconds, err := strconv.ParseInt(string(c.Argv[1]), 10, 64)
	if err != nil {
		c.Conn.WriteError("ERR value is not an integer or out of range")
		return
	}

	if seconds <= 0 {
		c.Conn.WriteError("ERR invalid expire time in 'expire' command")
		return
	}

	ttl := time.Duration(seconds) * time.Second

	// First, try to set TTL on a regular string key
	readOutput, err := c.Engine.Read(&contract.ReadInput{Key: key})
	if err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	if readOutput.Exists {
		// Key exists as a string key, update its TTL
		_, err := c.Engine.Write(&contract.WriteInput{
			Key:     key,
			Value:   readOutput.Value,
			TTL:     ttl,
			KeepTTL: false,
		})
		if err != nil {
			c.Conn.WriteError("ERR " + err.Error())
			return
		}
		c.Conn.WriteInt(1)
		return
	}

	// Key doesn't exist as a string, try as a set
	result, err := c.Engine.ExpireSet(key, ttl)
	if err != nil {
		c.Conn.WriteError("ERR " + err.Error())
		return
	}

	c.Conn.WriteInt(result)
}
