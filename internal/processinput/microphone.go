// Package processinput handles audio input processing from the microphone.
package processinput

import (
	"fmt"
	"main/constants"
	"main/pkg/objects"
	"math"
	"math/cmplx"

	"github.com/go-gst/go-gst/examples"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"gonum.org/v1/gonum/dsp/fourier"
)

// Variables for FFT processing
var (
	fft           = fourier.NewFFT(constants.FFTWindowSize)      // FFT instance
	hammingWindow = createHammingWindow(constants.FFTWindowSize) // Hamming window for smoothing
	sampleBuffer  = make([]int16, 0, constants.FFTWindowSize*2)  // Buffer to hold audio samples
)

// createHammingWindow generates a Hamming window of the specified size.
func createHammingWindow(size int) []float64 {
	window := make([]float64, size)
	for i := 0; i < size; i++ {
		// Calculate Hamming window value
		window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(size-1))
	}
	return window
}

// createMicrophonePipeline sets up the GStreamer pipeline for microphone input.
func createMicrophonePipeline() (*gst.Pipeline, error) {
	gst.Init(nil) // Initialize GStreamer

	pipeline, err := gst.NewPipeline("") // Create a new GStreamer pipeline
	if err != nil {
		return nil, err
	}

	// Create elements: pulsesrc (microphone input), capsfilter, audioconvert, and appsink
	pulsesrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	capsfilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}

	// Set desired audio format (16-bit signed LE, mono, 44100 Hz)
	caps := gst.NewCapsFromString(
		"audio/x-raw, format=S16LE, layout=interleaved, channels=1, rate=44100",
	)
	capsfilter.SetProperty("caps", caps) // Apply caps to the filter

	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	sink, err := app.NewAppSink() // Create an app sink for processing audio samples
	if err != nil {
		return nil, err
	}

	// Add all elements to the pipeline
	pipeline.AddMany(pulsesrc, capsfilter, audioconvert, sink.Element)

	// Link elements: pulsesrc -> capsfilter -> audioconvert -> appsink
	if gst.ElementLinkMany(pulsesrc, capsfilter, audioconvert, sink.Element) != nil {
		return nil, fmt.Errorf("failed to link elements")
	}

	// Configure appsink caps to match desired format
	sink.SetCaps(gst.NewCapsFromString(
		"audio/x-raw, format=S16LE, layout=interleaved, channels=1",
	))

	// Set callbacks for processing audio samples from the sink
	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			sample := sink.PullSample() // Pull a new sample from the sink
			if sample == nil {
				return gst.FlowEOS // End of stream
			}

			buffer := sample.GetBuffer() // Get the buffer from the sample
			if buffer == nil {
				return gst.FlowError // Error if buffer is nil
			}
			defer buffer.Unmap() // Ensure buffer is unmapped after processing

			// Get new samples and add to buffer
			newSamples := buffer.Map(gst.MapRead).AsInt16LESlice()
			sampleBuffer = append(sampleBuffer, newSamples...) // Append new samples to the buffer

			objects.MicroPhoneChannel <- newSamples // Send samples to the MicroPhoneChannel

			// Process complete windows
			for len(sampleBuffer) >= constants.FFTWindowSize {
				// Extract window samples and apply Hamming window
				window := make([]float64, constants.FFTWindowSize)
				for i := 0; i < constants.FFTWindowSize; i++ {
					window[i] = float64(sampleBuffer[i]) * hammingWindow[i]
				}

				// Remove processed samples from buffer
				sampleBuffer = sampleBuffer[constants.FFTWindowSize:]

				// Perform FFT on the windowed samples
				fftCoefs := fft.Coefficients(nil, window)

				// Calculate frequency characteristics
				binWidth := constants.SampleRate / float64(constants.FFTWindowSize) // Calculate bin width
				startBin := int(math.Floor(constants.SpeechLowFreq / binWidth))     // Start bin for speech frequency
				endBin := int(math.Ceil(constants.SpeechHighFreq / binWidth))       // End bin for speech frequency

				// Calculate speech band energy
				var speechEnergy float64
				for i := startBin; i <= endBin && i < len(fftCoefs); i++ {
					speechEnergy += cmplx.Abs(fftCoefs[i]) // Accumulate energy in the speech band
				}

				// Dynamic threshold calculation (could be improved)
				totalEnergy := 0.0
				for _, c := range fftCoefs {
					totalEnergy += cmplx.Abs(c) // Calculate total energy
				}
				threshold := totalEnergy * 0.3 // 30% of total energy in speech band

				// Detect speech based on energy levels
				if speechEnergy > threshold && speechEnergy > 4000 {
					// Speech detected
					fmt.Println("Speech detected - Energy:", speechEnergy)
					objects.SpeakingFlag.Store(true) // Set speaking flag to true
				} else {
					// Silence detected
					fmt.Println("Silence - Energy:", speechEnergy)
					objects.SpeakingFlag.Store(false) // Set speaking flag to false
				}
			}

			return gst.FlowOK // Indicate successful processing
		},
	})

	return pipeline, nil // Return the created pipeline
}

// RunMicrophoneGeneration starts the microphone input processing.
func RunMicrophoneGeneration() {
	examples.Run(func() error {
		pipeline, err := createMicrophonePipeline() // Create the microphone pipeline
		if err != nil {
			return err
		}
		objects.MicrophonePipeline = pipeline // Store the pipeline in the objects package
		return mainLoop(pipeline)             // Start the main loop for processing
	})
}
