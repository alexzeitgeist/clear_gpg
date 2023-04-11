# GPG Agent Cache Clearer

This repository contains a Go program that clears the GPG Agent cache whenever the user locks their screen. It does this by listening to D-Bus signals for lock events and invoking commands to reset the GPG Agent when the screen is locked.

