package commands

import (
	commandhandlers "github.com/Refrag/redix/internals/redis/command_handlers"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

// Initialize initializes the core commands.
func RegisterHandlers() {
	// PING
	commandutilities.HandleFunc("ping", func(c *commandutilities.Context) {
		commandhandlers.Ping(c)
	})

	// QUIT
	commandutilities.HandleFunc("quit", func(c *commandutilities.Context) {
		commandhandlers.Quit(c)
	})

	// SELECT <DB index>
	commandutilities.HandleFunc("select", func(c *commandutilities.Context) {
		commandhandlers.Select(c)
	})

	// GET <key> [DELETE]
	commandutilities.HandleFunc("get", func(c *commandutilities.Context) {
		commandhandlers.Get(c)
	})

	// GETDEL <key> =
	// same as: GET <key> DELETE
	commandutilities.HandleFunc("getdel", func(c *commandutilities.Context) {
		commandhandlers.GetDel(c)
	})

	// SET <key> <value> [EX seconds | KEEPTTL] [NX]
	commandutilities.HandleFunc("set", func(c *commandutilities.Context) {
		commandhandlers.Set(c)
	})

	// TTL <key>
	commandutilities.HandleFunc("ttl", func(c *commandutilities.Context) {
		commandhandlers.Ttl(c)
	})

	// INCR <key> [<delta>]
	commandutilities.HandleFunc("incr", func(c *commandutilities.Context) {
		commandhandlers.Incr(c)
	})

	// INCRBY <key> <delta>
	commandutilities.HandleFunc("incrby", func(c *commandutilities.Context) {
		commandutilities.Call("incr", c)
	})

	// DEL key [key ...]
	commandutilities.HandleFunc("del", func(c *commandutilities.Context) {
		commandhandlers.Del(c)
	})

	// HGETALL <prefix>
	commandutilities.HandleFunc("hgetall", func(c *commandutilities.Context) {
		commandhandlers.HGetAll(c)
	})

	// FLUSHALL
	commandutilities.HandleFunc("flushall", func(c *commandutilities.Context) {
		commandhandlers.FlushAll(c)
	})

	// FLUSHDB
	commandutilities.HandleFunc("flushdb", func(c *commandutilities.Context) {
		commandhandlers.FlushDb(c)
	})

	// PUBLISH
	commandutilities.HandleFunc("publish", func(c *commandutilities.Context) {
		commandhandlers.Publish(c)
	})

	commandutilities.HandleFunc("subscribe", func(c *commandutilities.Context) {
		commandhandlers.Subscribe(c)
	})

	commandutilities.HandleFunc("CLIENTCOUNT", func(c *commandutilities.Context) {
		commandhandlers.ClientCount(c)
	})
}
