// Package processinput handles the audio input processing for tone generation.
package processinput

import (
	"fmt"
	"main/pkg/objects"
	"os"
	"strings"

	"github.com/go-gst/go-gst/examples"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

// createPipeline initializes a GStreamer pipeline for audio processing from a specified file path.
func createPipeline(audioFilePath string) (*gst.Pipeline, error) {
	// Initialize GStreamer
	gst.Init(nil)

	// Create a new GStreamer pipeline
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, err
	}

	// Create elements: filesrc (source), decodebin (for decoding), audioconvert (for format conversion), and appsink (for output)
	filesrc, err := gst.NewElement("filesrc")
	if err != nil {
		return nil, err
	}
	filesrc.SetProperty("location", audioFilePath) // Set the location of the audio file

	decodebin, err := gst.NewElement("decodebin")
	if err != nil {
		return nil, err
	}

	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	sink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}

	// Add all elements to the pipeline
	pipeline.AddMany(filesrc, decodebin, audioconvert, sink.Element)

	// Link filesrc to decodebin
	if gst.ElementLinkMany(filesrc, decodebin) != nil {
		return nil, fmt.Errorf("failed to link filesrc to decodebin")
	}

	// Link audioconvert to appsink
	if gst.ElementLinkMany(audioconvert, sink.Element) != nil {
		return nil, fmt.Errorf("failed to link audioconvert to appsink")
	}

	// Handle dynamic pad addition from decodebin
	decodebin.Connect("pad-added", func(self *gst.Element, newPad *gst.Pad) {
		// Check if the new pad is for audio
		caps := newPad.GetCurrentCaps()
		if caps == nil {
			fmt.Println("Pad has no caps")
			return
		}
		structure := caps.GetStructureAt(0)
		if structure == nil {
			fmt.Println("Pad has no structure")
			return
		}
		name := structure.Name()
		if !strings.HasPrefix(name, "audio/") {
			fmt.Println("Skipping non-audio pad")
			return
		}

		// Get audioconvert's sink pad and link it to the new pad
		sinkPad := audioconvert.GetStaticPad("sink")
		if sinkPad == nil {
			fmt.Println("Failed to get audioconvert's sink pad")
			return
		}

		if linkRet := newPad.Link(sinkPad); linkRet != gst.PadLinkOK {
			fmt.Printf("Pad linking failed: %s\n", linkRet)
		}
	})

	// Configure appsink caps for desired audio format
	sink.SetCaps(gst.NewCapsFromString(
		"audio/x-raw, format=S16LE, layout=interleaved, channels=1",
	))

	// Set up sample processing callback for appsink
	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			sample := sink.PullSample() // Pull a sample from the sink
			if sample == nil {
				return gst.FlowEOS // End of stream if no sample is available
			}

			buffer := sample.GetBuffer() // Get the buffer from the sample
			if buffer == nil {
				return gst.FlowError // Return error if buffer is nil
			}

			// Map buffer and process samples
			samples := buffer.Map(gst.MapRead).AsInt16LESlice()
			defer buffer.Unmap() // Ensure buffer is unmapped after processing

			// Send samples to the ToneChannel for further processing
			objects.ToneChannel <- samples

			// RMS calculation (commented out for now)
			// var square float64
			// for _, i := range samples {
			// 	square += float64(i * i)
			// }
			// rms := math.Sqrt(square / float64(len(samples)))
			// fmt.Println("tone rms:", rms)

			return gst.FlowOK // Indicate successful processing
		},
	})

	return pipeline, nil // Return the created pipeline
}

// RunToneGeneration starts the tone generation process using the specified audio file.
func RunToneGeneration() {
	audioFilePath := "data/file_example_MP3_5MG.mp3" // Default audio file path
	if len(os.Args) > 1 {
		audioFilePath = os.Args[1] // Override with command line argument if provided
	}
	examples.Run(func() error {
		pipeline, err := createPipeline(audioFilePath) // Create the audio processing pipeline
		if err != nil {
			return err
		}
		objects.TonePipeline = pipeline // Store the pipeline in the objects package
		return mainLoop(pipeline)       // Start the main loop for processing
	})
}
