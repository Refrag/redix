package commands

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Refrag/redix/internals/datastore/contract"
)

func init() {
	// PING
	HandleFunc("ping", func(c *Context) {
		c.Conn.WriteString("PONG")
	})

	// QUIT
	HandleFunc("quit", func(c *Context) {
		c.Conn.WriteString("OK")
		c.Conn.Close()
	})

	// SELECT <DB index>
	HandleFunc("select", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments supplied")
			return
		}

		i, err := strconv.Atoi(string(c.Argv[0]))
		if err != nil {
			c.Conn.WriteError("Err invalid DB index")
			return
		}

		c.SessionSet("namespace", fmt.Sprintf("/%d/", i))

		c.Conn.WriteString("OK")
	})

	// GET <key> [DELETE]
	HandleFunc("get", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		delete := false

		for i := 1; i < c.Argc; i++ {
			switch strings.ToLower(string(c.Argv[i])) {
			case "delete":
				delete = true
			}
		}

		ret, err := c.Engine.Read(&contract.ReadInput{
			Key:    c.AbsoluteKeyPath(c.Argv[0]),
			Delete: delete,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		if len(ret.Value) < 1 {
			c.Conn.WriteNull()
			return
		}

		c.Conn.WriteBulk(ret.Value)
	})

	// GETDEL <key> =
	// same as: GET <key> DELETE
	HandleFunc("getdel", func(c *Context) {
		if c.Argc != 1 {
			c.Conn.WriteError("Err invalid number of arguments specified")
			return
		}

		c.Argc++
		c.Argv = append(c.Argv, []byte("DELETE"))

		Call("get", c)
	})

	// SET <key> <value> [EX seconds | KEEPTTL] [NX]
	HandleFunc("set", func(c *Context) {
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
	})

	// TTL <key>
	HandleFunc("ttl", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		ret, err := c.Engine.Read(&contract.ReadInput{
			Key: c.AbsoluteKeyPath(c.Argv[0]),
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		if !ret.Exists {
			c.Conn.WriteBulkString("-2")
			return
		}

		if ret.TTL == 0 {
			c.Conn.WriteBulkString("-1")
			return
		}

		c.Conn.WriteAny(ret.TTL.Milliseconds())
	})

	// INCR <key> [<delta>]
	HandleFunc("incr", func(c *Context) {
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
					log.Println("[FATAL]", err.Error())
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
	})

	// INCRBY <key> <delta>
	HandleFunc("incrby", func(c *Context) {
		Call("incr", c)
	})

	// DEL key [key ...]
	HandleFunc("del", func(c *Context) {
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
	})

	// HGETALL <prefix>
	HandleFunc("hgetall", func(c *Context) {
		prefix := []byte("")

		if c.Argc > 0 {
			prefix = c.Argv[0]
		}

		result := map[string]string{}

		err := c.Engine.Iterate(&contract.IteratorOpts{
			Prefix: c.AbsoluteKeyPath(prefix),
			Callback: func(ro *contract.ReadOutput) error {
				endKey := strings.TrimPrefix(string(ro.Key), string(c.AbsoluteKeyPath(prefix)))
				result[endKey] = string(ro.Value)
				return nil
			},
		})

		if err != nil && err != contract.ErrStopIterator {
			c.Conn.WriteError("ERR " + err.Error())
		}

		c.Conn.WriteAny(result)
	})

	// FLUSHALL
	HandleFunc("flushall", func(c *Context) {
		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   nil,
			Value: nil,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})

	// FLUSHDB
	HandleFunc("flushdb", func(c *Context) {
		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   c.AbsoluteKeyPath(),
			Value: nil,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})

	// PUBLISH
	HandleFunc("publish", func(c *Context) {
		if c.Argc < 2 {
			c.Conn.WriteError("ERR wrong number of arguments for 'publish' command")
			return
		}

		if err := c.Engine.Publish(c.AbsoluteKeyPath([]byte("redix"), c.Argv[0]), c.Argv[1]); err != nil {
			// FIX: Removed invalid format string placeholder that wasn't being used
			// Was: "ERR %s " + err.Error() - the %s had no corresponding argument
			c.Conn.WriteError("ERR " + err.Error())
			return
		}

		c.Conn.WriteInt(0)
	})

	HandleFunc("subscribe", func(c *Context) {
		if c.Argc != 1 {
			c.Conn.WriteError("ERR wrong number of arguments for 'subscribe' command")
			return
		}

		conn := c.Conn.Detach()
		defer conn.Close()

		conn.WriteArray(3)
		conn.WriteBulkString("subscribe")
		conn.WriteBulk(c.Argv[0])
		conn.WriteInt(1)
		conn.Flush()

		err := c.Engine.Subscribe(c.AbsoluteKeyPath([]byte("redix"), c.Argv[0]), func(msg []byte) error {
			conn.WriteArray(3)
			conn.WriteBulkString("message")
			conn.WriteBulk(c.Argv[0])
			conn.WriteBulk(msg)
			conn.Flush()
			return nil
		})

		if err != nil {
			c.Conn.WriteError("ERR " + err.Error())
			return
		}
	})
}
