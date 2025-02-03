// Package objects provides channels and pipelines for audio processing.
package objects

import (
	"sync/atomic" // Importing atomic for safe concurrent access to flags.

	"github.com/go-gst/go-gst/gst" // Importing GStreamer package for handling multimedia.
)

// ToneChannel is a channel for transmitting audio samples of tone generation.
var ToneChannel chan []int16

// MicroPhoneChannel is a channel for transmitting audio samples from the microphone.
var MicroPhoneChannel chan []int16

// SpeakingFlag is an atomic boolean flag indicating whether speech is detected.
var SpeakingFlag atomic.Bool

// TonePipeline is the GStreamer pipeline for tone generation.
var TonePipeline *gst.Pipeline

// MicrophonePipeline is the GStreamer pipeline for microphone input processing.
var MicrophonePipeline *gst.Pipeline

// OutputPipeline is the GStreamer pipeline for audio output processing.
var OutputPipeline *gst.Pipeline
