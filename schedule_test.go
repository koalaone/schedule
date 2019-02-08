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
	"log"
	"testing"
	"time"
)

func EventSecond(param string) {
	log.Printf("second event value:%v\n", param)
}

func EventSecondWithWait(param string) {
	log.Printf("second event value:%v\n", param)

	time.Sleep(12 * time.Second)
}

func EventMinute(param string) {
	log.Printf("minute event value:%v\n", param)
}

func EventHour(param string) {
	log.Printf("hour event value:%v\n", param)
}

func EventDay(param string) {
	log.Printf("day event value:%v\n", param)
}

func EventAtDatetime(param string) {
	log.Printf("AtDatetime event value:%v\n", param)
}

func TestScheduler(t *testing.T) {
	err := EverySeconds(1).Do(EventSecond, "second")
	if err != nil {
		t.Errorf("test schedule EventSecond error:%v", err.Error())
		return
	}

	err = EverySeconds(2).Do(EventSecondWithWait, "second")
	if err != nil {
		t.Errorf("test schedule EventSecondWithWait error:%v", err.Error())
		return
	}

	err = EveryMinutes(1).Do(EventMinute, "minute")
	if err != nil {
		t.Errorf("test schedule EventMinute error:%v", err.Error())
		return
	}

	err = EveryHours(1).Do(EventHour, "hour")
	if err != nil {
		t.Errorf("test schedule EventHour error:%v", err.Error())
		return
	}

	err = EveryDays(1).Do(EventDay, "day")
	if err != nil {
		t.Errorf("test schedule EventDay error:%v", err.Error())
		return
	}

	err = AtDateTime(2018, time.December, 21, 16, 59, 10).
		Do(EventAtDatetime, "at_datetime")
	if err != nil {
		t.Errorf("test schedule AtDateTime error:%v", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	Start(ctx)
	defer func() {
		cancel()
	}()

	time.Sleep(3 * time.Minute)
}
