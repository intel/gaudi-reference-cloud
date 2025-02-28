package mock

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// FakeEventRecorder is a mock implementation of record.EventRecorder
type FakeEventRecorder struct {
	Events []string
}

func (f FakeEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	f.Events = append(f.Events, fmt.Sprintf("Event: %s %s %s", eventtype, reason, message))
}

func (f FakeEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	f.Events = append(f.Events, fmt.Sprintf("Eventf: %s %s %s", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
}

func (f FakeEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	f.Events = append(f.Events, fmt.Sprintf("AnnotatedEventf: %s %s %s", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
}
