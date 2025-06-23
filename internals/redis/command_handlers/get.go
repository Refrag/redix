package commandhandlers

import (
	"strings"

	"github.com/Refrag/redix/internals/datastore/contract"
	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Get(c *commandutilities.Context) {
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
}
