# go-gst-speech-detection

# Audio Processing

This project is an audio processing application built using Go and GStreamer. It captures audio from a microphone, processes it, and can also generate tones. The application uses channels for communication between different components and employs GStreamer for multimedia processing.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Audio Processing with FFT](#audio-processing-with-fft)
- [Running the Application](#running-the-application)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Features

- Capture audio from the microphone.
- Process audio input and detect speech.
- Generate audio tones from specified files.
- Output processed audio through GStreamer.

## Requirements

- Go 1.16 or later
- GStreamer 1.16 or later
- GStreamer Go bindings (`github.com/go-gst/go-gst`)

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/CK200/go-gst-speech-detection.git
   cd go-gst-speech-detection
   ```

2. **Install Go dependencies:**

   ```bash
   go mod tidy
   ```

3. **Install GStreamer:**

   Follow the installation instructions for GStreamer based on your operating system. You can find the instructions [here](https://gstreamer.freedesktop.org/download/).

## Configuration

### Constants

The application uses constants defined in the `constants/constants.go` file. You can modify the following constants to configure the application:

- `SampleRate`: The sample rate for audio processing (default: `44100.0` Hz).
- `FFTWindowSize`: The size of the FFT window (default: `2048`).
- `SpeechLowFreq`: The lower frequency threshold for speech detection (default: `300.0` Hz).
- `SpeechHighFreq`: The upper frequency threshold for speech detection (default: `4000.0` Hz).

### Audio File Path

The default audio file path for tone generation is set in the `internal/processinput/tone.go` file. You can change the path to point to your desired audio file:

go
audioFilePath := "data/file_example_MP3_5MG.mp3" // Default audio file path


You can also override this path by providing it as a command-line argument when running the application.


## Audio Processing with FFT

The application utilizes Fast Fourier Transform (FFT) to analyze audio signals in the frequency domain. This allows for effective speech detection and audio processing. Here’s a detailed explanation of how the application processes audio data using FFT:

### Overview of FFT

FFT is an efficient algorithm for computing the Discrete Fourier Transform (DFT) and its inverse. The DFT converts a sequence of time-domain samples into a sequence of frequency-domain components, allowing us to analyze the frequency content of the audio signal. This is particularly useful for applications such as speech recognition, audio analysis, and filtering.

### How the Application Uses FFT

1. **Audio Capture**:
   - The application captures audio input from the microphone using GStreamer. The audio samples are streamed in real-time and sent to a designated channel for processing.

2. **Windowing**:
   - To analyze the audio signal, the application collects samples into a buffer. Once the buffer reaches a specified size (defined by `FFTWindowSize`), the application extracts a segment of audio samples for processing.
   - A Hamming window is applied to the samples to reduce spectral leakage. This windowing function smooths the edges of the sample segment, which helps in obtaining a more accurate frequency representation.

3. **FFT Calculation**:
   - After applying the Hamming window, the application performs the FFT on the windowed samples. This transforms the time-domain samples into frequency-domain coefficients.
   - The FFT coefficients represent the amplitude and phase of the frequency components present in the audio signal.

4. **Frequency Analysis**:
   - The application calculates the energy in specific frequency bands to detect speech. It defines a range of frequencies (from `SpeechLowFreq` to `SpeechHighFreq`) that are relevant for speech detection.
   - The energy in these frequency bands is computed by summing the magnitudes of the FFT coefficients that fall within the defined range.

5. **Speech Detection**:
   - The application uses a dynamic threshold to determine whether speech is present. It calculates the total energy of the FFT coefficients and sets a threshold based on a percentage of this total energy.
   - If the energy in the speech frequency band exceeds the threshold, the application detects speech and sets a flag (`SpeakingFlag`) to indicate that speech is currently being detected.

6. **Continuous Processing**:
   - The application continuously processes incoming audio samples in real-time. As new samples are captured, they are added to the buffer, and the FFT analysis is performed on each complete window of samples.
   - This allows the application to respond to changes in the audio signal dynamically, making it suitable for real-time applications such as voice activation or speech recognition.

### Example of FFT Processing Code

Here’s a simplified snippet of how the FFT processing is implemented in the application:

```go
// Perform FFT on the windowed samples
fftCoefs := fft.Coefficients(nil, window)

// Calculate frequency characteristics
binWidth := constants.SampleRate / float64(constants.FFTWindowSize)
startBin := int(math.Floor(constants.SpeechLowFreq / binWidth))
endBin := int(math.Ceil(constants.SpeechHighFreq / binWidth))

// Calculate speech band energy
var speechEnergy float64
for i := startBin; i <= endBin && i < len(fftCoefs); i++ {
    speechEnergy += cmplx.Abs(fftCoefs[i]) // Accumulate energy in the speech band
}
```

### Conclusion

By leveraging FFT, the application can effectively analyze audio signals in real-time, enabling features such as speech detection and audio processing. This approach allows for a deeper understanding of the audio content and enhances the application's capabilities in handling audio data.

## Running the Application

1. **Run the application**:
   ```bash
   go run cmd/main.go
   ```

2. **To specify a different audio file for tone generation**:
   ```bash
   go run cmd/main.go path/to/your/audiofile.mp3
   ```

3. **Stop the application**:
   You can stop the application by pressing `Ctrl+C`. This will send an end-of-stream event to all pipelines and gracefully shut down the application.

## Project Structure

```
go-gst-speech-detection/
├── cmd/                     # Entry point of the application
│   └── main.go              # Main application logic
├── internal/                # Core logic of the application
│   ├── processinput/        # Audio input processing
│   │   ├── microphone.go    # Microphone input handling
│   │   └── tone.go          # Tone generation handling
│   └── processoutput/       # Audio output processing
│       └── output.go        # Output handling logic
├── pkg/                     # Shared packages
│   └── objects/             # Shared objects and channels
│       └── objects.go       # Channels and pipeline references
├── constants/               # Constants used throughout the application
│   └── constants.go         # Constant definitions
├── data/                    # Sample data (audio files, etc.)
├── go.mod                   # Go module file
└── go.sum                   # Go module checksum file
```


## Usage

### Audio Processing Flow

1. **Tone Generation**: The application can generate tones from specified audio files. The audio data is read from the file and sent to the `ToneChannel`.

2. **Microphone Input**: The application captures audio from the microphone and sends the audio samples to the `MicroPhoneChannel`.

3. **Output Processing**: The application processes the audio data from both channels and outputs the audio through the configured GStreamer pipeline.

### Example Commands

- To run the application with the default audio file:
  ```bash
  go run cmd/main.go
  ```

- To run the application with a custom audio file:
  ```bash
  go run cmd/main.go path/to/your/audiofile.mp3
  ```

## Contributing

Contributions are welcome! If you have suggestions for improvements or new features, please feel free to submit a pull request or open an issue.

### Steps to Contribute

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Push your changes to your forked repository.
5. Submit a pull request to the main repository.

## License

This project is licensed under ME. See the [LICENSE](LICENSE) file for details.
