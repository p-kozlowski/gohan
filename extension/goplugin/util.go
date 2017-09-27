package goplugin

import "github.com/golang/mock/gomock"

var controllers map[gomock.TestReporter]*gomock.Controller

func NewController(testReporter gomock.TestReporter) *gomock.Controller {
	ctrl := gomock.NewController(testReporter)
	if controllers == nil {
		controllers = make(map[gomock.TestReporter]*gomock.Controller)
	}
	controllers[testReporter] = ctrl
	return ctrl
}

func Finish(testReporter gomock.TestReporter) {
	controllers[testReporter].Finish()
}
