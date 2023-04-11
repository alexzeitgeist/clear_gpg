package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/godbus/dbus/v5"
)

const dbusRule = "type='signal',interface='org.freedesktop.login1.Session',member='Lock'"

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Error connecting to system bus:", err)
	}

	if err = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, dbusRule).Err; err != nil {
		log.Fatal("Error adding D-Bus match:", err)
	}

	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)

	for sig := range ch {
		if sig.Name == "org.freedesktop.login1.Session.Lock" {
			if err = clearAll(); err != nil {
				log.Println("Error clearing GPG agent:", err)
			} else {
				log.Println("Successfully cleared keys in the gpg-agent")
			}
		}
	}
}

func clearAll() error {
	cmds := []struct {
		name string
		args []string
	}{
		{"gpg-connect-agent", []string{"SCD RESET", "/bye"}},
		{"gpg-connect-agent", []string{"reloadagent", "/bye"}},
		{"pkill", []string{"-HUP", "gpg-agent"}},
	}

	for _, c := range cmds {
		if err := runCommand(c.name, c.args); err != nil {
			return fmt.Errorf("error running %s: %w", c.name, err)
		}
	}
	return nil
}

func runCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = io.Discard
	return cmd.Run()
}
