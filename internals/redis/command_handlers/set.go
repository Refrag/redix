package commandhandlers

import (
	"bytes"
	"errors"
	log "log/slog"
	"strconv"
	"time"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

const (
	SetArgumentsMinCount = 2
)

// parseSetOptions parses the set options.
func parseSetOptions(c *commandutilities.Context, writeOpts *contract.WriteInput) error {
	if c.Argc <= SetArgumentsMinCount {
		return nil
	}

	for i := SetArgumentsMinCount; i < len(c.Argv); i++ {
		chr := string(bytes.ToLower(c.Argv[i]))
		switch chr {
		case "ex":
			if i+1 >= len(c.Argv) {
				return errors.New("EX requires a value")
			}
			n, err := strconv.ParseInt(string(c.Argv[i+1]), 10, 64)
			if err != nil {
				return err
			}
			writeOpts.TTL = time.Second * time.Duration(n)
			i++
		case "keepttl":
			writeOpts.KeepTTL = true
		case "nx":
			writeOpts.OnlyIfNotExists = true
		}
	}
	return nil
}

// Set sets a key-value pair in the database.
func Set(c *commandutilities.Context) {
	if c.Argc < SetArgumentsMinCount {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	writeOpts := contract.WriteInput{
		Key:   c.AbsoluteKeyPath(c.Argv[0]),
		Value: c.Argv[1],
	}

	if err := parseSetOptions(c, &writeOpts); err != nil {
		c.Conn.WriteError("Err " + err.Error())
		return
	}

	if c.Cfg.Server.Redis.AsyncWrites {
		go (func() {
			if _, err := c.Engine.Write(&writeOpts); err != nil {
				log.Error("[FATAL]", "error", err.Error())
			}
		})()
	} else {
		if _, err := c.Engine.Write(&writeOpts); err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}
	}

	c.Conn.WriteString("OK")
}
