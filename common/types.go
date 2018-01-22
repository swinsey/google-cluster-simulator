package common

const (
	TIME_DELAY = 0
)

type Machine struct {
	MachineID  int64
	PlatformID string
	Cpus       float64
	Mem        float64
	UsedCpus   float64
	UsedMem    float64
}

type Job struct {
	MissingInfo     int64
	JobID           int64 `csv:"job_id"`
	User            string
	SchedulingClass int64
	JobName         string
	LogicalJobName  string
	TaskNum         int64 `csv:"task_num"`
	SubmitTime      int64 `csv:"submit_time"`
	StartTime       int64
	EndTime         int64
	Duration        int64 `csv:"duration"`

	Share      float64
	index      int
	deleted    bool
	CpuUsed    float64
	MemUsed    float64
	TaskDone   int64
	TaskSubmit int64
	taskQueue  map[int64]*Task
}

func NewJob(job Job) *Job {
	return &Job{
		JobID:      job.JobID,
		SubmitTime: job.SubmitTime,
		StartTime:  job.StartTime,
		EndTime:    job.EndTime,
		TaskNum:    job.TaskNum,
		Duration:   job.Duration,
		Share:      job.Share,

		taskQueue: make(map[int64]*Task),
	}
}

func (job *Job) Done() bool {
	return job.TaskDone == job.TaskNum
}

type TaskStatus int

const (
	TASK_STATUS_FINISHED TaskStatus = iota
	TASK_STATUS_RUNNING
	TASK_STATUS_STAGING
	TASK_STATUS_KILLED
)

type Task struct {
	MissingInfo                  int64
	JobID                        int64   `csv:"job_id"`
	TaskIndex                    int64   `csv:"task_index"`
	MachineID                    int64
	User                         string
	SchedulingClass              int64
	Priority                     int64
	CpuRequest                   float64 `csv:"cpu_request"`
	MemoryRequest                float64 `csv:"memory_request"`
	DiskSpaceRequest             float64
	DifferentMachinesRestriction bool
	SubmitTime                   int64   `csv:"submit_time"`
	StartTime                    int64
	EndTime                      int64
	Duration                     int64   `csv:"duration"`

	Status        TaskStatus
	CpuSlack      float64
	MemSlack      float64
	Oversubscribe bool
}

func NewTask(task Task) *Task {
	return &Task{
		JobID:            task.JobID,
		TaskIndex:        task.TaskIndex,
		CpuRequest:       task.CpuRequest,
		MemoryRequest:    task.MemoryRequest,
		DiskSpaceRequest: task.DiskSpaceRequest,
		SubmitTime:       task.SubmitTime,
		EndTime:          task.EndTime,
		StartTime:        task.StartTime,
		Duration:         task.Duration,
	}
}

func TaskID(task *Task) int64 {
	return task.JobID*1333 + task.TaskIndex
}

func GetTaskID(jobId, taskIndex int64) int64 {
	return jobId*1333 + taskIndex
}

type MachineEventType int

const (
	MACHINE_ADD    MachineEventType = iota
	MACHINE_REMOVE
	MACHINE_UPDATE
)

type TaskEventType int

const (
	TASK_SUBMIT         TaskEventType = iota
	TASK_SCHEDULE
	TASK_EVICT
	TASK_FAIL
	TASK_FINISH
	TASK_KILL
	TASK_LOST
	TASK_UPDATE_PENDING
	TASK_UPDATE_RUNNING
)

type EventOrigin int

const (
	EVENT_MACHINE EventOrigin = iota
	EVENT_JOB
	EVENT_TASK
)

type Event struct {
	Id               string
	Time             int64
	Machine          *Machine
	Job              *Job
	Task             *Task
	EventOrigin      EventOrigin
	TaskEventType    TaskEventType
	MachineEventType MachineEventType
}

type EventChans map[EventOrigin]chan *Event

type EventSlice [] *Event

func (e EventSlice) Len() int {
	return len(e)
}

func (e EventSlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e EventSlice) Less(i, j int) bool {
	return e[j].Time > e[i].Time
}

func ParseMachineEvent(record []string) (*Event, error) {
	machine := &Machine{
		MachineID:  stringToInt64(record[1], 0),
		PlatformID: record[3],
		Cpus:       stringToFloat64(record[4], 0.0),
		Mem:        stringToFloat64(record[5], 0.0),
		UsedCpus:   0.0,
		UsedMem:    0.0,
	}

	return &Event{
		Time:             stringToInt64(record[0], 0),
		MachineEventType: MachineEventType(stringToInt64(record[2], 0)),
		Machine:          machine,
		EventOrigin:      EVENT_MACHINE,
	}, nil
}

func ParseTaskEvent(record []string) (*Event, error) {
	task := &Task{
		JobID:                        stringToInt64(record[3], 0),
		TaskIndex:                    stringToInt64(record[4], 0),
		MachineID:                    stringToInt64(record[5], 0),
		User:                         record[7],
		SchedulingClass:              stringToInt64(record[8], 0),
		Priority:                     stringToInt64(record[9], 0),
		CpuRequest:                   stringToFloat64(record[10], 0.0),
		MemoryRequest:                stringToFloat64(record[11], 0.0),
		DiskSpaceRequest:             stringToFloat64(record[12], 0.0),
		DifferentMachinesRestriction: stringToBool(record[13], false),
		Duration:                     stringToInt64(record[15], 0),
	}

	return &Event{
		Time:          stringToInt64(record[1], 0),
		TaskEventType: TaskEventType(stringToInt64(record[6], 0)),
		Task:          task,
		EventOrigin:   EVENT_TASK,
	}, nil
}

func ParseJobEvent(record []string) (*Event, error) {
	job := &Job{
		JobID:           stringToInt64(record[3], 0),
		User:            record[5],
		SchedulingClass: stringToInt64(record[6], 0),
		JobName:         record[7],
		LogicalJobName:  record[8],
		TaskNum:         stringToInt64(record[9],0),
	}

	return &Event{
		Time:          stringToInt64(record[1], 0),
		TaskEventType: TaskEventType(stringToInt64(record[4], 0)),
		Job:           job,
		EventOrigin:   EVENT_JOB,
	}, nil
}

type TaskUsage struct {
	StartTime int64 `csv:"start_time"`
	EndTime   int64 `csv:"end_time"`
	JobID     int64 `csv:"job_id"`
	TaskIndex int64 `csv:"task_index"`
	//MachineID                  int64 `csv:"machine_id"`
	CpuUsage    float64 `csv:"cpu_rate"`
	MemoryUsage float64 `csv:"canonical_memory_usage"`
	//AssignedMemoryUsage        float64 `csv:"assigned_memory_usage"`
	//UnmappedPageCache          float64 `csv:"unmapped_page_cache"`
	//TotalPageCache             float64 `csv:"total_page_cache"`
	//MaxMemoryUsage             float64 `csv:"maximum_memory_usage"`
	//DiskIOTime                 float64 `csv:"disk_io_time"`
	//LocalDiskSpaceUsage        float64 `csv:"local_disk_space_usage"`
	//MaxCpuRate                 float64 `csv:"maximum_cpu_rate"`
	//MaxDiskIOTime              float64 `csv:"maximum_disk_io_time"`
	//CyclePerInstruction        float64 `csv:"cycles_per_instruction"`
	//MemoryAccessPerInstruction float64 `csv:"memory_access_per_instruction"`
	//SamplePortion              float64 `csv:"sample_portion"`
	//AggregationType            bool `csv:"aggregation_type"`
	//SampleCpuUsage             float64 `csv:"sampled_cpu_usage"`
}

func NewTaskUsage(t TaskUsage) *TaskUsage {
	return &TaskUsage{
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
		JobID:       t.JobID,
		TaskIndex:   t.TaskIndex,
		CpuUsage:    t.CpuUsage,
		MemoryUsage: t.MemoryUsage,
	}
}

func ParseTaskUsage(record []string) TaskUsage {
	return TaskUsage{
		StartTime:   stringToInt64(record[4], 0),
		EndTime:     stringToInt64(record[5], 0),
		JobID:       stringToInt64(record[0], 0),
		TaskIndex:   stringToInt64(record[1], 0),
		CpuUsage:    stringToFloat64(record[2], 0.0),
		MemoryUsage: stringToFloat64(record[3], 0.0),
	}
}
