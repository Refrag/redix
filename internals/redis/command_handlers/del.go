package commandhandlers

import (
	log "log/slog"
	"path/filepath"
	"strings"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func deleteSingleKey(c *commandutilities.Context, keyPattern string) error {
	_, err := c.Engine.Write(&contract.WriteInput{
		Key:   c.AbsoluteKeyPath([]byte(keyPattern)),
		Value: nil,
	})
	return err
}

func deleteWildcardKeys(c *commandutilities.Context, keyPattern string, deletedCount *int) error {
	keyPattern = strings.TrimLeft(keyPattern, "/")
	prefix := strings.Split(keyPattern, "*")[0]

	return c.Engine.Iterate(&contract.IteratorOpts{
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

			if !matched {
				return nil
			}

			_, err = c.Engine.Write(&contract.WriteInput{
				Key:   ro.Key,
				Value: nil,
			})
			if err != nil {
				return err
			}
			*deletedCount++
			return nil
		},
	})
}

func deleteKey(c *commandutilities.Context, keyPattern string, deletedCount *int) error {
	if !strings.Contains(keyPattern, "*") {
		return deleteSingleKey(c, keyPattern)
	}
	return deleteWildcardKeys(c, keyPattern, deletedCount)
}

func Del(c *commandutilities.Context) {
	if c.Argc < 1 {
		c.Conn.WriteError("Err invalid arguments specified")
		return
	}

	deletedCount := 0

	if c.Cfg.Server.Redis.AsyncWrites {
		go (func() {
			for i := range c.Argv {
				keyPattern := string(c.Argv[i])
				if err := deleteKey(c, keyPattern, &deletedCount); err != nil {
					log.Error("[FATAL] DEL error:", "error", err.Error())
				}
			}
		})()
		c.Conn.WriteString("OK")
		return
	}

	for i := range c.Argv {
		keyPattern := string(c.Argv[i])
		if err := deleteKey(c, keyPattern, &deletedCount); err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}
	}

	c.Conn.WriteInt(deletedCount)
}
