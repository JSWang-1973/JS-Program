package main

import (
	"gerrit.o-ran-sc.org/r/scp/ric-app/kpimon/control"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGHUP)
	signal.Notify(sigs, os.Interrupt)
	signal.Notify(sigs, syscall.SIGTERM)

	c := control.NewControl()

	go func() {
		fmt.Println("wait signal...")
		sig := <-sigs
		fmt.Println("got signal:")
		fmt.Println(sig)
		done <- true
		c.Stop()
	}()

	c.Run()
//	c.Stop()

	fmt.Println("waiting signal")
	<-done
	fmt.Println("exiting")
}

