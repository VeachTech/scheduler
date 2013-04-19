/**
Copyright (c) 2013, Ryan Veach
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of the <organization> nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
**/

package scheduler

import (
	"sort"
	"time"
)

// ID type you can drop into a struct to fulfill the GetID()
// method of the Job interface
type ID uint

// Matches the GetID function of the Job interface, all it does is return the underlying uint
func (id ID) GetID() uint {
	return uint(id)
}

// Any object that implements this interface can be added to the scheduler
type Job interface {
	NextRunTime(time.Time) time.Time
	Run()
	GetID() uint
}

// This is the struct that allows you to communicate with the underlying
// scheduler. With it you can Add, Update, or Remove a Job aswell as shutting down the scheduler
type Scheduler struct {
	addOrUpdate chan Job
	remove      chan uint
}

// You MUST use this to create a new Scheduler
func New(buffered int) *Scheduler {
	var newSched *Scheduler
	newSched = &Scheduler{addOrUpdate: make(chan Job, buffered), remove: make(chan uint, buffered)}
	// Create worker
	go scheduleWorker(newSched.addOrUpdate, newSched.remove)
	return newSched
}

func (s *Scheduler) AddUpdateJob(job Job) {
	s.addOrUpdate <- job
}

func (s *Scheduler) RemoveJob(id uint) {
	s.remove <- id
}

func (s *Scheduler) ShutDown() {
	close(s.addOrUpdate)
	close(s.remove)
}

// Will sleep/loop constantly, add/update/remove/execute jobs and finish when addOrUpdate channel closes
func scheduleWorker(addOrUpdate <-chan Job, remove <-chan uint) {
	var jobData worker
	jobData.jobMap = make(map[uint]Job)
	for {
		jobData.updateTimesList()
		select {
		case <-time.After(jobData.closestTime().Sub(time.Now())):
			jobData.run()
		case job, ok := <-addOrUpdate:
			if !ok {
				return
			}
			jobData.addUpdate(job)
		case id := <-remove:
			jobData.remove(id)
		}
	}
}

// internal type for the scheduler
type jobTime struct {
	id       uint
	nextTime time.Time
}

type jobTimes []*jobTime

// These are used to implement the sorting Interface
func (s jobTimes) Len() int      { return len(s) }
func (s jobTimes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Will sort by the Job run time
type byTime struct{ jobTimes }

func (s byTime) Less(i, j int) bool { return s.jobTimes[i].nextTime.Before(s.jobTimes[j].nextTime) }

type worker struct {
	jobMap    map[uint]Job
	timesList jobTimes
}

func (w *worker) addUpdate(job Job) {
	w.jobMap[job.GetID()] = job
}

func (w *worker) remove(id uint) {
	delete(w.jobMap, id)
}

func (w *worker) updateTimesList() {
	var now = roundDownToSecond(time.Now().Add(time.Second))
	w.timesList = make(jobTimes, 0, len(w.jobMap))
	for _, job := range w.jobMap {
		w.timesList = append(w.timesList, &jobTime{job.GetID(), job.NextRunTime(now)})
	}
	sort.Sort(byTime{w.timesList})
}

func (w *worker) closestTime() time.Time {
	if len(w.timesList) == 0 {
		return time.Now().AddDate(0, 0, 1)
	}
	return w.timesList[0].nextTime
}

func (w *worker) run() {
	for _, job := range w.timesList {
		if job.nextTime.Before(time.Now()) {
			go w.jobMap[job.id].Run()
		} else {
			break
		}
	}
}

func roundDownToSecond(t time.Time) time.Time {
	return t.Add(-(time.Duration(t.Nanosecond()) * time.Nanosecond))
}
