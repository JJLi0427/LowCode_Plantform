package appplugin

import (
	"fmt"
	"onlinetools/core/control"
	"path"
	"strings"

	"github.com/robfig/cron/v3"
)

func NewScheduler() *TaskScheduler {
	return &TaskScheduler{cron: cron.New(), records: make(map[string][]cron.EntryID)}
}

type TaskScheduler struct {
	records map[string][]cron.EntryID
	cron    *cron.Cron

	endtasks map[string][]cron.Job
}

func (t *TaskScheduler) AddTask(ctl *control.Control) {
	if ids, ok := t.records[ctl.Name]; ok {
		for _, id := range ids {
			t.cron.Remove(id)
		}
	}

	if _, ok := t.endtasks[ctl.Name]; ok {
		t.endtasks[ctl.Name] = nil
	}

	var entryids []cron.EntryID
	for i, task := range ctl.Backtasks {

		entry := &exeEntry{
			cmdenvs: append(task.Envs, ctl.GetEnvs()...),
			cmdroot: task.Workdir,
			cmd:     task.Cmd,
			ishell:  task.Inline_shell,
			runner:	 task.GetAppService(),
			add:     task.Add,
			copy:    task.Copy,
			trace:   task.Trace,
			period:  task.Period,
			args:    task.Args,
			cmdtype: fmt.Sprintf("backtask%d", i),
			appctrl: path.Join(ctl.ControlFilePath, "control.yaml")}
		if fs := strings.Fields(task.Period); len(fs) == 5 {
			if id, err := t.cron.AddJob(task.Period, entry); err == nil {
				entryids = append(entryids, id)
			} else {
				fmt.Println("ctl backtask", err)
			}
		} else if task.Period == "end" {
			t.endtasks[ctl.Name] = append(t.endtasks[ctl.Name], entry)
		} else if task.Period == "start" {
			entry.Run()
		}

		//exec copy lib dependence command
		if len(task.Cmd) > 0 && task.Packdepend {
			entry.CopyDependence()
		}
		//exec add and copy command
		if len(task.Add) > 0 || len(task.Copy) > 0 {
			entry.runADDCopy()
		}
	}

	if len(entryids) > 0 {
		t.records[ctl.Name] = entryids
	}
}

func (t *TaskScheduler) Start() {
	t.cron.Start()
}

func (t *TaskScheduler) Stop() {
	t.cron.Stop()

	for _, tasks := range t.endtasks {
		for _, task := range tasks {
			task.Run()
		}
	}
}

func (t *TaskScheduler) AddCronTask(TaskUniqueName, cronPeriod string, job cron.Job) error {
	if fs := strings.Fields(cronPeriod); len(fs) == 5 {
		if id, err := t.cron.AddJob(cronPeriod, job); err == nil {
			t.records[TaskUniqueName] = []cron.EntryID{id}
		} else {
			fmt.Println(fmt.Sprintf("add [%s] task failed:", TaskUniqueName), err)
		}
	}

	return fmt.Errorf("add [%s] task failed: cronPeriod's format does not supported", TaskUniqueName)
}
