package commandhandlers

import (
	"fmt"
	"strconv"

	commandutilities "github.com/Refrag/redix/internals/redis/command_utilities"
)

func Select(c *commandutilities.Context) {
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
}
