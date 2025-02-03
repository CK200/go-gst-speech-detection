// This example shows how to use the appsink element for audio processing.
package main

import (
	"main/internal/processinput"  // Importing the processinput package for handling audio input
	"main/internal/processoutput" // Importing the processoutput package for handling audio output
	"main/pkg/objects"            // Importing the objects package for shared data structures
	"os"                          // Importing the os package for operating system functionalities
	"os/signal"                   // Importing the signal package for handling OS signals

	"github.com/go-gst/go-gst/gst" // Importing the GStreamer library for multimedia processing
)

func main() {
	// Create channels for tone and microphone audio data with a buffer size of 10
	objects.ToneChannel = make(chan []int16, 10)
	objects.MicroPhoneChannel = make(chan []int16, 10)

	// Create a channel to listen for OS interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) // Notify the channel on interrupt signals (like Ctrl+C)

	// Start the tone generation in a separate goroutine
	go processinput.RunToneGeneration()
	// Start the microphone input processing in a separate goroutine
	go processinput.RunMicrophoneGeneration()
	// Start the output sink processing in a separate goroutine
	go processoutput.RunOutputSink()

	// Wait for an interrupt signal
	<-c

	// Send end-of-stream events to all pipelines to gracefully shut them down
	objects.TonePipeline.SendEvent(gst.NewEOSEvent())
	objects.MicrophonePipeline.SendEvent(gst.NewEOSEvent())
	objects.OutputPipeline.SendEvent(gst.NewEOSEvent())
}
