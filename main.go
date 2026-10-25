package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	cf "drift/internals/config"
	"drift/internals/errs"
	"drift/internals/fs"
	id "drift/internals/identity"
	"drift/internals/toml"
	"drift/internals/wire"
)

func main() {

	userConfig, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	_, err = fs.CreateDir(userConfig, "drift")
	if err != nil {
		log.Fatal(err)
	}

	tomlConfigPath := filepath.Join(userConfig, "drift", "config.toml")

	var config cf.Config
	if err = toml.Init(tomlConfigPath, &config, cf.DefaultConfig, false); err != nil {
		var appErr *errs.Error
		if errors.As(err, &appErr) {
			if appErr.Is(errs.ErrDecodingToml) {
				fmt.Fprintf(os.Stderr, corruptConfigFileMsg, tomlConfigPath, err)
				os.Exit(1)
			}
		}
	}

	tomlIdentityPath := filepath.Join(userConfig, "drift", "identity.toml")

	var identity id.Keys
	if err = toml.Init(tomlIdentityPath, &identity, make(map[string]any), false); err != nil {
		var appErr *errs.Error
		if errors.As(err, &appErr) {
			if appErr.Is(errs.ErrDecodingToml) {
				fmt.Fprintf(os.Stderr, corruptIdentityFileMsg, tomlIdentityPath, err)
				os.Exit(1)
			}
		}
	}

	err = identity.Validate()
	if err != nil {
		if errors.Is(err, errs.ErrMissingKeys) {
			err = identity.Generate()
			if err != nil {
				log.Fatal(err)
			}

			err = toml.Write(tomlIdentityPath, identity)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Fprintf(os.Stderr, corruptIdentityFileMsg, tomlIdentityPath, err)
			os.Exit(1)
		}
	}

	port := strconv.Itoa(config.Server.Port)
	if port == "" {
		log.Fatal("invalid server port")
	}

	wire.Listen(config.Stream.MaxConnections, config.Stream.MaxMsgStreamSize, config.Stream.MaxFileStreamSize, port, config.Download)

}

var corruptConfigFileMsg = `
Configuration file is corrupted.

File: %s
Error: %s

To fix it:
  1. Edit the configuration file and correct the error.
  2. Restore the default configuration with:
     drift --default

Warning: --default overwrites your existing configuration.

Configuration template:
https://github.com/dario61k/drift/blob/main/config.example.toml

`

var corruptIdentityFileMsg = `
Idenity file is corrupted.

File: %s
Error: %s

To fix it:
  1. Edit the identity file and correct the error.
  2. Restore the default configuration with:
     drift --default --identity

Warning: --default --identity removes you from the Drift network; your peers will have to add you back.

Configuration template:
https://github.com/dario61k/drift/blob/main/identity.example.toml

`
