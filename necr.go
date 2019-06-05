package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var cmd *exec.Cmd

/*
 * necr
 * Run a command and restart it when it or a child ends.
 *
 * TODO 
 * Enable/disable logging
 *     -> Maybe put log file into a userspace folder...
 * Possibly move for loop into its own method
 */
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: necr [command]")
		os.Exit(1)
	}

	commandString := os.Args[1]
	fmt.Println(commandString)

	// set up channel to receive child signals
	sigchldChannel := make(chan os.Signal, 1)
	signal.Notify(sigchldChannel, syscall.SIGCHLD)

	go sigintHandler()

	for {
		cmd = exec.Command(commandString)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		startTime := time.Now()
		appendLog(fmt.Sprintf("\nStarted {%s} at %s\n", commandString, startTime.String()))
		err := cmd.Start()
		if err != nil {
			appendLog("Failed to run")
			fmt.Println(err)
		}

		// main program enters sleep state while waiting for signal
		<-sigchldChannel
		sigchldHandler()
		cmd.Wait()

		endTime := time.Now()
		appendLog(fmt.Sprintf("Closed at %s\n", endTime.String()))
		appendLog(fmt.Sprintf("Total run time: %s\n", endTime.Sub(startTime).String()))

		appendLog("Process restarting in 10 seconds")
		for i := 0; i < 10; i++ {
			fmt.Print(".")
			time.Sleep(time.Second)
		}
		appendLog("\n")
	}
}

/**
 * Terminate the process group when a SIGCHLD is received
 */
func sigchldHandler() {
	pgid, _ := syscall.Getpgid(cmd.Process.Pid)
	err := syscall.Kill(-pgid, syscall.SIGTERM)
	if err != nil {
		fmt.Println(err)
	}
}

/**
 * If SIGINT (or SIGTERM) is sent to the main process, need to ensure
 * that the script and its children are also shut down.
 */
func sigintHandler() {
	sigintChannel := make(chan os.Signal, 1)
	signal.Notify(sigintChannel, syscall.SIGINT, syscall.SIGTERM)

	<-sigintChannel
	sigchldHandler()
	appendLog("Exiting\n")
	os.Exit(0)
}

func appendLog(s string) {
	f, err := os.OpenFile("necr.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening or creating log file.")
		return
	}

	fmt.Print(s)

	f.Write([]byte(s))
	f.Close()
}

