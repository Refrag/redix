package commandhandlers

import (
	log "log/slog"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Incr(c *commandutilities.Context) {
	if c.Argc < 1 {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	delta := []byte("1")
	if c.Argc > 1 {
		delta = c.Argv[1]
	}

	if c.Cfg.Server.Redis.AsyncWrites {
		go (func() {
			if _, err := c.Engine.Write(&contract.WriteInput{
				Key:       c.AbsoluteKeyPath(c.Argv[0]),
				Value:     delta,
				Increment: true,
			}); err != nil {
				log.Error("[FATAL]", "error", err.Error())
			}
		})()

		c.Conn.WriteNull()
		return
	}

	ret, err := c.Engine.Write(&contract.WriteInput{
		Key:       c.AbsoluteKeyPath(c.Argv[0]),
		Value:     delta,
		Increment: true,
	})

	if err != nil {
		c.Conn.WriteError("Err " + err.Error())
		return
	}

	c.Conn.WriteBulk(ret.Value)
}
