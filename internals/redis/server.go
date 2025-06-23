package redis

import (
	"fmt"
	"log"

	"github.com/Refrag/redix/internals/config"
	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
	"github.com/tidwall/redcon"
)

// ListenAndServe start a redis server
func ListenAndServe(cfg *config.Config, engine contract.Engine) error {
	commandutilities.HandleFunc("CLIENTCOUNT", func(c *commandutilities.Context) {
		c.Conn.WriteAny(commandutilities.GetConnCounter())
	})

	fmt.Println("=> started listening on", cfg.Server.Redis.ListenAddr, "...")
	return redcon.ListenAndServe(cfg.Server.Redis.ListenAddr,
		func(conn redcon.Conn, cmd redcon.Command) {
			handleCommand(conn, cmd, engine, cfg)
		},
		func(conn redcon.Conn) bool {
			return accept(cfg, conn)
		},
		func(conn redcon.Conn, err error) {
			closed(conn, err)
		},
	)
}

func accept(cfg *config.Config, conn redcon.Conn) bool {
	if cfg.Server.Redis.MaxConns > 0 && cfg.Server.Redis.MaxConns <= commandutilities.GetConnCounter() {
		log.Println("max connections reached!")
		return false // reject connection
	}

	commandutilities.IncrementConnCounter()

	conn.SetContext(map[string]interface{}{
		"namespace": "/0/",
	})
	return true // accept connection
}

func handleCommand(conn redcon.Conn, cmd redcon.Command, engine contract.Engine, cfg *config.Config) {
	ctxPointer := commandutilities.NewContext(conn, engine, cfg, cmd.Args[1:], len(cmd.Args)-1)

	commandutilities.Call(string(cmd.Args[0]), ctxPointer)
}

func closed(conn redcon.Conn, err error) {
	commandutilities.DecrementConnCounter()
}
