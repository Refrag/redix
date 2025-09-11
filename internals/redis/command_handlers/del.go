package commandhandlers

import (
	log "log/slog"
	"path/filepath"
	"strings"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func deleteSingleKey(c *commandutilities.Context, keyPattern string) error {
	key := c.AbsoluteKeyPath([]byte(keyPattern))

	// Try to delete as a string key first
	_, err := c.Engine.Write(&contract.WriteInput{
		Key:   key,
		Value: nil,
	})
	if err != nil {
		return err
	}

	// Also try to delete as a set key (this won't error if the set doesn't exist)
	if delErr := c.Engine.DelSet(key); delErr != nil {
		// Log but don't return error - we want DEL to succeed if either type was deleted
		log.Warn("Failed to delete set key:", "key", keyPattern, "error", delErr.Error())
	}

	return nil
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
		// Redis for Go actually executes NewIntCmd and not NewStringCmd for Del.
		// We can't know the amount of keys deleted asynchronously, so we just return 1.
		c.Conn.WriteInt(1)
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
