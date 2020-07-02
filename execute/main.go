package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	enviroment = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

// Lets play with cmdline in linux
// ffmpeg -re -i url -c copy -flush_packets 0 -f mpegts "udp://127.0.0.1:1234?pkt_size=1316&buffer_size=65535"
func main() {
	comando := "ffmpeg -re -i url -c copy -flush_packets 0 -f mpegts udp://127.0.0.1:1234?pkt_size=1316&buffer_size=65535"
	cmd := exec.Command("ffmpeg")
	cmd.Args = strings.Fields(comando)
	cmd.Env = append(os.Environ(), enviroment)

	cmdline := ""
	for _, v := range cmd.Args {
		cmdline = cmdline + " " + v
	}
	done := make(chan error)
	fmt.Printf("EXEC: %s\n", cmdline)
	fmt.Println("Commandline executed async ...")
	cmd.Start()

	go func() {
		done <- cmd.Wait()
	}()

	// waiting to be killed
	select {
	case <-time.After(time.Hour * 900000): // beyond a century
		log.Println("Timeout hit..")
		return
	case err := <-done:
		if err != nil {
			log.Println("ffmpeg failed:", err)
		}

	fmt.Println("Program exited !!!")
}
