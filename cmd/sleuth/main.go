package main

import (
	"os"

	"github.com/giraffesyo/sleuth/internal/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	os.Stderr.WriteString(`
         .-""-.    ____  _     _____ _   _ _____ _   _ 
 _______/      \  / ___|| |   | ____| | | |_   _| | | |
|_______        ; \___ \| |   |  _| | | | | | | | |_| |
        \      /   ___) | |___| |___| |_| | | | |  _  |
         '-..-'   |____/|_____|_____|\___/  |_| |_| |_|
		`)
	cli.RootCmd.Execute()
}
