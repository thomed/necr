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
 * Have the command be a command line parameter?
 * Log start/stop times to file?
 *     -> Make times global and do this from sigchldHandler
 * Possibly move for loop into its own method
 * See if possible to get CPU time of process
 *     -> Might not be worth because most CPU time is not in cmd
 */
func main() {
	// set up channel to receive child signals
	sigchldChannel := make(chan os.Signal, 1)
	signal.Notify(sigchldChannel, syscall.SIGCHLD)

	go sigintHandler()

	for {
		cmd = exec.Command("./startserv.sh")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		startTime := time.Now()
		fmt.Println("\nStarted at ", startTime)
		err := cmd.Start()
		if err != nil {
			fmt.Println("Failed to run")
			fmt.Println(err)
		}

		// main program enters sleep state while waiting for signal
		<-sigchldChannel
		sigchldHandler()
		cmd.Wait()

		endTime := time.Now()
		fmt.Println("Closed at ", endTime)
		fmt.Println("Total run time: ", endTime.Sub(startTime).String())

		fmt.Print("Process restarting in 10 seconds")
		for i := 0; i < 10; i++ {
			fmt.Print(".")
			time.Sleep(time.Second)
		}
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
	os.Exit(0)
}

