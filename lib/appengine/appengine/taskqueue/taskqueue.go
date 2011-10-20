// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

/*
Package taskqueue provides a client for App Engine's taskqueue service.
Using this service, applications may perform work outside a user's request.

A Task may be constucted manually; alternatively, since the most common
taskqueue operation is to add a single POST task, NewPOSTTask makes it easy.

	t := taskqueue.NewPOSTTask("/worker", url.Values{
		"key": {key},
	})
	taskqueue.Add(c, t, "") // add t to the default queue
*/
package taskqueue

// TODO: Bulk task adding/deleting, queue management.

import (
	"fmt"
	"http"
	"os"
	"time"
	"url"

	"appengine"
	"appengine_internal"
	"goprotobuf.googlecode.com/hg/proto"

	taskqueue_proto "appengine_internal/taskqueue"
)

// A Task represents a task to be executed.
type Task struct {
	// Path is the worker URL for the task.
	// If unset, it will default to /_ah/queue/<queue_name>.
	Path string

	// Payload is the data for the task.
	// This will be delivered as the HTTP request body.
	// It is only used when Method is POST, PUT or PULL.
	// http.EncodeQuery may be used to generate this for POST requests.
	Payload []byte

	// Additional HTTP headers to pass at the task's execution time.
	// To schedule the task to be run with an alternate app version
	// or backend, set the "Host" header.
	Header http.Header

	// Method is the HTTP method for the task ("GET", "POST", etc.),
	// or "PULL" if this is task is destined for a pull-based queue.
	// If empty, this defaults to "POST".
	Method string

	// A name for the task.
	// If empty, a name will be chosen.
	Name string

	// Delay is how far into the future this task should execute, in microseconds.
	Delay int64
}

// NewPOSTTask creates a Task that will POST to a path with the given form data.
func NewPOSTTask(path string, params url.Values) *Task {
	h := make(http.Header)
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	return &Task{
		Path:    path,
		Payload: []byte(params.Encode()),
		Header:  h,
		Method:  "POST",
	}
}

// Add adds the task to a named queue.
// An empty queue name means that the default queue will be used.
// Add returns an equivalent Task with defaults filled in, including setting
// the task's Name field to the chosen name if the original was empty.
func Add(c appengine.Context, task *Task, queueName string) (*Task, os.Error) {
	if queueName == "" {
		queueName = "default"
	}
	req := &taskqueue_proto.TaskQueueAddRequest{
		QueueName: []byte(queueName),
		TaskName:  []byte(task.Name),
		EtaUsec:   proto.Int64(time.Nanoseconds()/1e3 + task.Delay),
	}
	method := task.Method
	if method == "" {
		method = "POST"
	}
	if method == "PULL" {
		// Pull-based task
		req.Body = task.Payload
		req.Mode = taskqueue_proto.NewTaskQueueMode_Mode(taskqueue_proto.TaskQueueMode_PULL)
		// TODO: More fields will need to be set.
	} else {
		// HTTP-based task
		if v, ok := taskqueue_proto.TaskQueueAddRequest_RequestMethod_value[method]; ok {
			req.Method = taskqueue_proto.NewTaskQueueAddRequest_RequestMethod(
				taskqueue_proto.TaskQueueAddRequest_RequestMethod(v))
		} else {
			return nil, fmt.Errorf("taskqueue: bad method %q", method)
		}
		req.Url = []byte(task.Path)
		for k, vs := range task.Header {
			for _, v := range vs {
				req.Header = append(req.Header, &taskqueue_proto.TaskQueueAddRequest_Header{
					Key:   []byte(k),
					Value: []byte(v),
				})
			}
		}
		if method == "POST" || method == "PUT" {
			req.Body = task.Payload
		}
	}
	res := &taskqueue_proto.TaskQueueAddResponse{}
	if err := c.Call("taskqueue", "Add", req, res, nil); err != nil {
		return nil, err
	}
	resultTask := *task
	resultTask.Method = method
	if task.Name == "" {
		resultTask.Name = string(res.ChosenTaskName)
	}
	return &resultTask, nil
}

// Delete deletes a task from a named queue.
func Delete(c appengine.Context, task *Task, queueName string) os.Error {
	req := &taskqueue_proto.TaskQueueDeleteRequest{
		QueueName: []byte(queueName),
		TaskName:  [][]byte{[]byte(task.Name)},
	}
	res := &taskqueue_proto.TaskQueueDeleteResponse{}
	if err := c.Call("taskqueue", "Delete", req, res, nil); err != nil {
		return err
	}
	for _, ec := range res.Result {
		if ec != taskqueue_proto.TaskQueueServiceError_OK {
			return &appengine_internal.APIError{
				Service: "taskqueue",
				Code:    int32(ec),
			}
		}
	}
	return nil
}

// LeaseTasks leases tasks from a queue.
// leaseTime is in seconds.
// The number of tasks fetched will be at most maxTasks.
func LeaseTasks(c appengine.Context, maxTasks int, queueName string, leaseTime int) ([]*Task, os.Error) {
	req := &taskqueue_proto.TaskQueueQueryAndOwnTasksRequest{
		QueueName:    []byte(queueName),
		LeaseSeconds: proto.Float64(float64(leaseTime)),
		MaxTasks:     proto.Int64(int64(maxTasks)),
	}
	res := &taskqueue_proto.TaskQueueQueryAndOwnTasksResponse{}
	if err := c.Call("taskqueue", "QueryAndOwnTasks", req, res, nil); err != nil {
		return nil, err
	}
	tasks := make([]*Task, len(res.Task))
	for i, t := range res.Task {
		// TODO: Handle eta_usec, retry_count.
		tasks[i] = &Task{
			Payload: t.Body,
			Name:    string(t.TaskName),
			Method:  "PULL",
		}
	}
	return tasks, nil
}

// Purge removes all tasks from a queue.
func Purge(c appengine.Context, queueName string) os.Error {
	req := &taskqueue_proto.TaskQueuePurgeQueueRequest{
		QueueName: []byte(queueName),
	}
	res := &taskqueue_proto.TaskQueuePurgeQueueResponse{}
	return c.Call("taskqueue", "PurgeQueue", req, res, nil)
}

func init() {
	appengine_internal.RegisterErrorCodeMap("taskqueue", taskqueue_proto.TaskQueueServiceError_ErrorCode_name)
}
