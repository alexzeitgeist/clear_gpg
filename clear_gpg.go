package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Println("Error connecting to system bus:", err)
		return
	}

	rule := "type='signal',interface='org.freedesktop.login1.Session',member='Lock'"
	if err = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule).Err; err != nil {
		fmt.Println("Error adding D-Bus match:", err)
		return
	}

	ch := make(chan *dbus.Signal, 10)
	conn.Signal(ch)

	for sig := range ch {
		if sig.Name == "org.freedesktop.login1.Session.Lock" {
			if err = clearAll(); err != nil {
				fmt.Println("Error clearing GPG agent:", err)
			}
		}
	}
}

func clearAll() (err error) {
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	defer func() {
		closeErr := null.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close /dev/null: %w", closeErr)
		}
	}()

	cmds := []struct {
		name   string
		args   []string
		stdout *os.File
	}{
		{"gpg-connect-agent", []string{"SCD RESET", "/bye"}, null},
		{"gpg-connect-agent", []string{"reloadagent", "/bye"}, null},
		{"pkill", []string{"-HUP", "gpg-agent"}, null},
	}

	for _, c := range cmds {
		if err = runCommand(c.name, c.args, c.stdout); err != nil {
			return fmt.Errorf("error running %s: %w", c.name, err)
		}
	}
	return nil
}

func runCommand(name string, args []string, stdout *os.File) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	return cmd.Run()
}
