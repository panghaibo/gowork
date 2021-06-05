package anet

//some global poll api
var (
	MainPollApi *EventLoopApi
)

//initialize poll
func init() {
	var err error
	MainPollApi, err = GetEventApi(1024)
	if err != nil {
		panic(err)
	}
}