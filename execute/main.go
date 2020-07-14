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
// ffmpeg -re -i url -c copy -flush_packets 0 -f mpegts 'udp://127.0.0.1:1234?pkt_size=1316&buffer_size=65535'
func main() {
	//comando := "muxer,-re,-i,rtmp://url,-c,copy,-f,mpegts,-mpegts_original_network_id,1,-mpegts_transport_stream_id,1,-mpegts_service_id,1,-mpegts_pmt_start_pid,129,-mpegts_start_pid,1024,-metadata,service_provider='TodoStreaming',-metadata,service_name='Channel 1',tcp://127.0.0.1:1500"
	comando := "muxer,-i,tcp://127.0.0.1:1501,-c,copy,-f,flv,rtmp://url"
	cmd := exec.Command("muxer")
	cmd.Args = strings.Split(comando, ",")
	cmd.Env = append(os.Environ(), enviroment)

	cmdline := ""
	for _, v := range cmd.Args {
		cmdline = cmdline + " " + v
	}
	fmt.Printf("EXEC: %s\n", cmdline)
	fmt.Println("Commandline executed async ...")

	done := make(chan error)
	cmd.Start()
	// timer := time.NewTimer(time.Hour * 900000) // beyond a century
	timer := time.NewTimer(time.Second * 120) // timeout in seconds
	go func() {
		done <- cmd.Wait()
	}()

	// waiting to be killed
	select {
	case <-timer.C: // timeout in seconds
		log.Println("Timeout hit...")
		cmd.Process.Signal(os.Interrupt) // SIGINT (-2)
		//cmd.Process.Signal(os.Kill)      // SIGKILL (-9)
		<-done // drain the channel
	case err := <-done:
		if err != nil {
			log.Println("ffmpeg failed:", err)
		}
		if !timer.Stop() { // stops timer from firing
			<-timer.C // drain the channel
		}
	}

	fmt.Println("Program exited !!!")
}
