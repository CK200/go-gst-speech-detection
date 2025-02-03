package constants

const (
	SampleRate     = 44100.0 // Sample rate in Hz
	FFTWindowSize  = 2048    // FFT window size (46ms window at 44100Hz)
	SpeechLowFreq  = 250.0   // Lower frequency threshold for speech detection
	SpeechHighFreq = 4000.0  // Upper frequency threshold for speech detection
)
