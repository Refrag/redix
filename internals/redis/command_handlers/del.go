package commandhandlers

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Del(c *commandutilities.Context) {
	if c.Argc < 1 {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	deletedCount := 0

	deleteKey := func(keyPattern string) error {
		// Check if the key contains wildcard character
		if strings.Contains(keyPattern, "*") {
			keyPattern = strings.TrimLeft(keyPattern, "/")

			prefix := strings.Split(keyPattern, "*")[0]
			err := c.Engine.Iterate(&contract.IteratorOpts{
				Prefix: c.AbsoluteKeyPath([]byte(prefix)),
				Callback: func(ro *contract.ReadOutput) error {
					keyToMatch := string(ro.Key)
					namespace, _ := c.SessionGet("namespace")
					if strings.HasPrefix(keyToMatch, namespace.(string)) {
						keyToMatch = strings.TrimPrefix(keyToMatch, namespace.(string))
					}

					matched, err := filepath.Match(keyPattern, keyToMatch)
					if err != nil {
						return err
					}

					if matched {
						_, err := c.Engine.Write(&contract.WriteInput{
							Key:   ro.Key,
							Value: nil,
						})
						if err != nil {
							return err
						}
						deletedCount++
					}
					return nil
				},
			})
			return err
		} else {
			_, err := c.Engine.Write(&contract.WriteInput{
				Key:   c.AbsoluteKeyPath([]byte(keyPattern)),
				Value: nil,
			})
			if err == nil {
				deletedCount++
			}
			return err
		}
	}

	if c.Cfg.Server.Redis.AsyncWrites {
		// NOTE: In async mode, deletedCount is not accurate since goroutines
		// complete after response is sent. This is a design trade-off for performance.
		go (func() {
			for i := range c.Argv {
				keyPattern := string(c.Argv[i])
				if err := deleteKey(keyPattern); err != nil {
					log.Println("[FATAL] DEL error:", err.Error())
				}
			}
		})()

		c.Conn.WriteString("OK")
		return
	}

	for i := range c.Argv {
		keyPattern := string(c.Argv[i])
		if err := deleteKey(keyPattern); err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}
	}

	c.Conn.WriteInt(deletedCount)
}
