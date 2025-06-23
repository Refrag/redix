//go:build linux || darwin

package main

import (
	log "log/slog"
	"os"

	"github.com/Refrag/redix/internals/config"
	"github.com/Refrag/redix/internals/datastore/contract"
	"github.com/Refrag/redix/internals/datastore/engines/filesystem"
	"github.com/Refrag/redix/internals/datastore/engines/postgresql"
	"github.com/Refrag/redix/internals/redis"
)

var (
	cfg *config.Config
)

const (
	minArgsCount = 2
)

func main() {
	if len(os.Args) < minArgsCount {
		log.Error("you must specify the configuration file as an argument")
		os.Exit(1)
	}

	var err error

	log.Info("=> registering engines ...")
	filesystem.Register()
	postgresql.Register()

	log.Info("=> loading the configs ...")

	cfg, err = config.Unmarshal(os.Args[1])
	if err != nil {
		log.Error("unable to load the config file due to: ", "error", err.Error())
		os.Exit(1)
	}

	db, err := contract.Open(cfg.Engine.Driver, cfg.Engine.DSN)
	if err != nil {
		log.Error("failed to open database connection due to: ", "error", err.Error())
		os.Exit(1)
	}

	redis.ListenAndServe(cfg, db)
}
