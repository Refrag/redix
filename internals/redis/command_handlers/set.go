package commandhandlers

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Set(c *commandutilities.Context) {
	if c.Argc < 2 {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	writeOpts := contract.WriteInput{
		Key:   c.AbsoluteKeyPath(c.Argv[0]),
		Value: c.Argv[1],
	}

	if c.Argc > 2 {
		// FIX: Added bounds checking to prevent panic when accessing i+1
		// Previously could crash if "ex" was the last argument
		for i := 0; i < len(c.Argv); i++ {
			chr := string(bytes.ToLower(c.Argv[i]))
			switch chr {
			case "ex":
				// Check if next argument exists before accessing it
				if i+1 >= len(c.Argv) {
					c.Conn.WriteError("Err EX requires a value")
					return
				}
				n, err := strconv.ParseInt(string(c.Argv[i+1]), 10, 64)
				if err != nil {
					c.Conn.WriteError("Err " + err.Error())
					return
				}
				writeOpts.TTL = time.Second * time.Duration(n)
				i++ // Skip the next argument since we consumed it
			case "keepttl":
				writeOpts.KeepTTL = true
			case "nx":
				writeOpts.OnlyIfNotExists = true
			}

		}
	}

	if c.Cfg.Server.Redis.AsyncWrites {
		go (func() {
			if _, err := c.Engine.Write(&writeOpts); err != nil {
				log.Println("[FATAL]", err.Error())
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
