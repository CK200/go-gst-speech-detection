package processinput

import (
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

func handleMessage(msg *gst.Message) error {
	switch msg.Type() {
	case gst.MessageEOS:
		return app.ErrEOS
	case gst.MessageError:
		return msg.ParseError()
	}
	return nil
}

func mainLoop(pipeline *gst.Pipeline) error {
	pipeline.SetState(gst.StatePlaying)
	defer pipeline.SetState(gst.StateNull)

	bus := pipeline.GetPipelineBus()
	for {
		msg := bus.TimedPop(gst.ClockTimeNone)
		if msg == nil {
			break
		}
		if err := handleMessage(msg); err != nil {
			return err
		}
	}
	return nil
}
