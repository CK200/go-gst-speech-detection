// Package processoutput handles the audio output processing.
package processoutput

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"main/pkg/objects"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/examples"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/go-gst/go-gst/gst/audio"
)

// createPipeline initializes a GStreamer pipeline for audio output processing.
func createPipeline() (*gst.Pipeline, error) {
	gst.Init(nil) // Initialize GStreamer

	// Create a new GStreamer pipeline
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, err // Return error if pipeline creation fails
	}

	// Create the necessary elements for the pipeline
	elems, err := gst.NewElementMany("appsrc", "audioconvert", "autoaudiosink")
	if err != nil {
		return nil, err // Return error if element creation fails
	}

	// Add the elements to the pipeline and link them together
	pipeline.AddMany(elems...)
	gst.ElementLinkMany(elems...)

	// Get the app source from the first element returned
	src := app.SrcFromElement(elems[0])

	// Specify the audio format for the appsrc element
	audioInfo := audio.NewInfo()
	audioInfo.SetFormat(audio.FormatS16LE, 44100, audio.ChannelPositionsFromMask(1, 4))

	// Set the caps for the appsrc element
	src.SetCaps(audioInfo.ToCaps())

	// Set the callbacks for the appsrc element to handle data requests
	src.SetCallbacks(&app.SourceCallbacks{
		NeedDataFunc: func(self *app.Source, _ uint) {
			// Check if there is data from the microphone channel
			if len(objects.MicroPhoneChannel) > 0 {
				micData := <-objects.MicroPhoneChannel // Retrieve microphone data
				micDataBytes := int16ToBytes(micData)  // Convert int16 slice to bytes

				// Create a new buffer for the microphone data
				micBuffer := gst.NewBufferWithSize(int64(len(micDataBytes)))
				micBuffer.Map(gst.MapWrite).WriteData(micDataBytes) // Write data to the buffer
				micBuffer.Unmap()                                   // Unmap the buffer after writing

				// Push the microphone buffer if speech is detected
				if objects.SpeakingFlag.Load() {
					self.PushBuffer(micBuffer)
					return
				}
			}

			// Check if there is data from the tone channel
			if len(objects.ToneChannel) > 0 {
				toneData := <-objects.ToneChannel       // Retrieve tone data
				toneDataBytes := int16ToBytes(toneData) // Convert int16 slice to bytes

				// Create a new buffer for the tone data
				toneBuffer := gst.NewBufferWithSize(int64(len(toneDataBytes)))
				toneBuffer.Map(gst.MapWrite).WriteData(toneDataBytes) // Write data to the buffer
				toneBuffer.Unmap()                                    // Unmap the buffer after writing

				// Push the tone buffer if no speech is detected
				if !objects.SpeakingFlag.Load() {
					self.PushBuffer(toneBuffer)
					return
				}
			}

			// If no data is available, push a dummy buffer
			fmt.Println("dummy buffer")
			dummyBuffer := gst.NewBufferWithSize(int64(64))
			self.PushBuffer(dummyBuffer) // Push the dummy buffer
		},
	})

	return pipeline, nil // Return the created pipeline
}

// handleMessage processes messages from the GStreamer pipeline.
func handleMessage(msg *gst.Message) error {
	switch msg.Type() {
	case gst.MessageEOS: // End of stream message
		return app.ErrEOS
	case gst.MessageError: // Error message
		gerr := msg.ParseError()
		if debug := gerr.DebugString(); debug != "" {
			fmt.Println(debug) // Print debug information if available
		}
		return gerr // Return the error
	}
	return nil // Return nil if no relevant message is handled
}

// int16ToBytes converts a slice of int16 values to a byte slice.
func int16ToBytes(int16Slice []int16) []byte {
	buf := new(bytes.Buffer) // Create a buffer to hold the byte data

	// Write each int16 value to the buffer
	for _, v := range int16Slice {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			fmt.Println("Error writing to buffer:", err) // Print error if writing fails
			return nil
		}
	}

	return buf.Bytes() // Return the byte slice
}

// mainLoop runs the GStreamer pipeline and handles messages.
func mainLoop(loop *glib.MainLoop, pipeline *gst.Pipeline) error {
	pipeline.Ref()         // Take a reference to the pipeline
	defer pipeline.Unref() // Ensure the pipeline is unreferenced when done

	pipeline.SetState(gst.StatePlaying) // Set the pipeline to playing state

	// Retrieve the bus from the pipeline and add a watch function
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		if err := handleMessage(msg); err != nil {
			fmt.Println(err) // Print error if message handling fails
			loop.Quit()      // Quit the main loop
			return false
		}
		return true // Continue watching for messages
	})

	loop.Run() // Run the main loop

	return nil // Return nil when done
}

// RunOutputSink starts the output sink processing.
func RunOutputSink() {
	examples.RunLoop(func(loop *glib.MainLoop) error {
		var pipeline *gst.Pipeline
		var err error
		if pipeline, err = createPipeline(); err != nil {
			return err // Return error if pipeline creation fails
		}
		objects.OutputPipeline = pipeline // Store the pipeline in the objects package
		return mainLoop(loop, pipeline)   // Start the main loop for processing
	})
}
