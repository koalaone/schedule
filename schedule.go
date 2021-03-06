/*
 *
 *
 *  * Copyright 2019 koalaone@163.com
 *  *
 *  * Licensed under the Apache License, Version 2.0 (the "License");
 *  * you may not use this file except in compliance with the License.
 *  * You may obtain a copy of the License at
 *  *
 *  *       http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS,
 *  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  * See the License for the specific language governing permissions and
 *  * limitations under the License.
 *
 */

package schedule

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"time"
)

var timeLocal = time.Local

func ChangeTimeLocation(newLocal *time.Location) {
	timeLocal = newLocal
}

// Task struct
type Task struct {
	isOnce   bool
	interval time.Duration
	running  bool
	lastRun  time.Time
	stop     bool
	gName    string
	gFunc    map[string]interface{}
	gParams  map[string]([]interface{})
}

// Scheduler struct
type Scheduler struct {
	running bool
	time    *time.Ticker
	tasks   []*Task
	sync.RWMutex
}

var schedule *Scheduler

func newScheduler() *Scheduler {
	if schedule == nil {
		schedule = &Scheduler{
			running: false,
			tasks:   make([]*Task, 0),
		}
	}

	return schedule
}

func GetScheduler() *Scheduler {
	if schedule == nil {
		newScheduler()
	}

	return schedule
}

func every(interval uint64, once bool) *Task {
	if interval <= 0 {
		interval = 1
	}
	newTask := &Task{
		isOnce:   once,
		interval: time.Duration(interval),
		lastRun:  time.Now(),
		stop:     false,
		gName:    "",
		gFunc:    make(map[string]interface{}, 0),
		gParams:  make(map[string]([]interface{}), 0),
	}

	if once {
		newTask.lastRun = time.Unix(int64(interval), 0)
	}

	if schedule == nil {
		newScheduler()
	}

	schedule.Add(newTask)

	return newTask
}

// Every Seconds Task
func EverySeconds(interval uint64) *Task {
	return every(interval, false)
}

// Every Minutes Task
func EveryMinutes(interval uint64) *Task {
	return every(interval*60, false)
}

// Every Hours Task
func EveryHours(interval uint64) *Task {
	return every(interval*60*60, false)
}

// Every Days Task
func EveryDays(interval uint64) *Task {
	return every(interval*60*60*24, false)
}

// Fixed execution at dateTime
func AtDateTime(year int, month time.Month, day, hour, minute, second int) *Task {
	return every(uint64(time.Date(year, month, day, hour, minute, second, 0, timeLocal).Unix()), true)
}

// Task execution
func (tk *Task) Do(taskFun interface{}, params ...interface{}) error {
	if taskFun == nil {
		return errors.New("param taskFun is nil")
	}

	typ := reflect.TypeOf(taskFun)
	if typ.Kind() != reflect.Func {
		return errors.New("param taskFun type error")
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(taskFun).Pointer()).Name()
	tk.gName = funcName
	tk.gFunc[funcName] = taskFun
	tk.gParams[funcName] = params

	return nil
}

func (tk *Scheduler) checkTaskStatus(isWait bool) bool {
retry:
	for _, taskItem := range tk.tasks {
		locTask := taskItem

		if locTask.running {
			if isWait {
				time.Sleep(5 * time.Millisecond)
				goto retry
			}

			return false
		}
	}

	return true
}

func (tk *Task) run(locNow time.Time) (result []reflect.Value, err error) {
	if tk.stop {
		return
	}

	if tk.isOnce && (locNow.Unix()-tk.lastRun.Unix() > 0) {
		return
	}

	if tk.running {
		return
	}

	tk.running = true
	defer func() {
		tk.running = false
	}()

	if !tk.isOnce {
		nextTime := tk.lastRun.Add(tk.interval * time.Second)
		if (locNow.Unix() - nextTime.Unix()) < 0 {
			return
		}
	} else {
		if (locNow.Unix() - tk.lastRun.Unix()) < 0 {
			return
		}
	}

	if tk.gFunc[tk.gName] == nil {
		return
	}

	f := reflect.ValueOf(tk.gFunc[tk.gName])
	params := tk.gParams[tk.gName]
	if len(params) != f.Type().NumIn() {
		err = errors.New(" param num not adapted ")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)

	if tk.isOnce {
		tk.lastRun = time.Unix(0, 0)
	}

	tk.lastRun = locNow

	return
}

// Add Task
func (tk *Scheduler) Add(value *Task) *Scheduler {
	if tk.running {
		return tk
	}

	if value == nil {
		return tk
	}

	tk.Lock()
	tk.tasks = append(tk.tasks, value)
	tk.Unlock()

	return tk
}

func (tk *Scheduler) runAll(locNow time.Time) {

	for _, taskItem := range tk.tasks {
		locTask := taskItem
		locTask.stop = false

		go func(task *Task) {
			_, _ = task.run(locNow)
		}(locTask)
	}

	return
}

func Stop() {
	if !schedule.running {
		return
	}

	schedule.Lock()
	schedule.running = false
	schedule.Unlock()

	for _, taskItem := range schedule.tasks {
		locTask := taskItem
		locTask.stop = true
	}
}

// schedule start
func Start(context context.Context) {
	if schedule == nil {
		newScheduler()
	}

	if schedule.running {
		return
	}

	schedule.Lock()
	schedule.running = true
	schedule.time = time.NewTicker(1 * time.Second)
	schedule.Unlock()

	go func() {
		for {
			select {
			case locNow := <-schedule.time.C:
				schedule.runAll(locNow)

			case <-context.Done():
				schedule.time.Stop()
				schedule.checkTaskStatus(true)
				return
			}
		}
	}()
}
