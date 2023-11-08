package base

type StreamNameFlag int

// redis streams
const (

	// ScheduleInstanceStream redis stream setting
	//key: stream name , value: consumer group name
	ScheduleInstanceStream = "s-scheduleinstance"
	VmActionStream         = "s-instance-action"
)

// stream consumer groups
const (
	// ScheduleInstanceStreamGroup consumer schedules instance while triggered by stream message
	ScheduleInstanceStreamGroup = "s-scheduleinstance-g"
	VmActionStreamGroup         = VmActionStream + "-g"
)

var streamGroupMap = map[string][]string{
	ScheduleInstanceStream: {ScheduleInstanceStreamGroup},
	VmActionStream:         {VmActionStreamGroup},
}

func GetStreamGroupMap() map[string][]string {
	return streamGroupMap
}
