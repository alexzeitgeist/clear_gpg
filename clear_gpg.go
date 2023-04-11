package main

import (
	"fmt"
	"os"
	"os/exec"

	dbus "github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Println("Error connecting to system bus:", err)
		return
	}

	matchRule := "type='signal',interface='org.freedesktop.login1.Session',member='Lock'"
	if err = conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule).Err; err != nil {
		fmt.Println("Error adding D-Bus match:", err)
		return
	}

	signals := make(chan *dbus.Signal, 10)
	conn.Signal(signals)

	for signal := range signals {
		if signal.Name == "org.freedesktop.login1.Session.Lock" {
			if err = clearAll(); err != nil {
				fmt.Println("Error clearing GPG agent:", err)
			}
		}
	}
}

func clearAll() error {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	defer devNull.Close()

	commands := []struct {
		name   string
		args   []string
		stdout *os.File
	}{
		{"gpg-connect-agent", []string{"SCD RESET", "/bye"}, devNull},
		{"gpg-connect-agent", []string{"reloadagent", "/bye"}, devNull},
		{"pkill", []string{"-HUP", "gpg-agent"}, devNull},
	}

	for _, cmd := range commands {
		if err = runCommand(cmd.name, cmd.args, cmd.stdout); err != nil {
			return fmt.Errorf("error running %s: %w", cmd.name, err)
		}
	}
	return nil
}

func runCommand(name string, args []string, stdout *os.File) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	return cmd.Run()
}
