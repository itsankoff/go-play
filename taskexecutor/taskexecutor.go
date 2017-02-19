package taskexecutor

import (
	"fmt"
	"sync"
	"time"
)

type Task interface {
	Execute(int) (int, error)
}

type SequentialExecutor struct {
	tasks []Task
}

func (s SequentialExecutor) Execute(arg int) (int, error) {
	if len(s.tasks) == 0 {
		return 0, fmt.Errorf("No tasks to execute")
	}

	result, err := s.tasks[0].Execute(arg)
	if err != nil {
		return 0, err
	}

	for _, task := range s.tasks[1:] {
		result, err = task.Execute(result)
		if err != nil {
			return 0, err
		}
	}

	return result, err
}

func Pipeline(tasks ...Task) Task {
	return SequentialExecutor{
		tasks: tasks,
	}
}

type ConcurrentExecutor struct {
	tasks []Task
}

type taskResult struct {
	result int
	err    error
}

func (c ConcurrentExecutor) Execute(arg int) (int, error) {
	if len(c.tasks) == 0 {
		return 0, fmt.Errorf("No tasks to execute")
	}

	results := make(chan taskResult)
	done := make(chan struct{})
	defer close(results)
	defer close(done)
	for _, task := range c.tasks {
		go func(task Task) {
			var tr taskResult
			tr.result, tr.err = task.Execute(arg)

			select {
			case <-done:
				return
			default:
				results <- tr
			}
		}(task)
	}

	res := <-results
	return res.result, res.err
}

func Fastest(tasks ...Task) Task {
	return ConcurrentExecutor{
		tasks: tasks,
	}
}

type TimedExecutor struct {
	task    Task
	timeout time.Duration
}

func (t TimedExecutor) Execute(arg int) (int, error) {
	result := make(chan taskResult)
	go func() {
		var tr taskResult
		tr.result, tr.err = t.task.Execute(arg)
		result <- tr
	}()

	select {
	case tr := <-result:
		return tr.result, tr.err
	case <-time.After(t.timeout):
		return 0, fmt.Errorf("Timeout reached")
	}
}

func Timed(task Task, timeout time.Duration) Task {
	return TimedExecutor{
		task:    task,
		timeout: timeout,
	}
}

type ReducerExecutor struct {
	tasks  []Task
	reduce func([]int) int
}

func (r ReducerExecutor) Execute(arg int) (int, error) {
	results := make(chan taskResult)
	done := make(chan struct{})
	defer close(results)
	defer close(done)

	for _, task := range r.tasks {
		go func(task Task) {
			var tr taskResult
			tr.result, tr.err = task.Execute(arg)

			select {
			case <-done:
				return
			default:
				results <- tr
			}
		}(task)
	}

	var resultVals []int
	for {
		tr := <-results
		if tr.err != nil {
			return tr.result, tr.err
		} else {
			resultVals = append(resultVals, tr.result)
			if len(resultVals) == len(r.tasks) {
				break
			}
		}
	}

	return r.reduce(resultVals), nil
}

func ConcurrentMapReduce(reduce func(results []int) int, tasks ...Task) Task {
	return ReducerExecutor{
		tasks:  tasks,
		reduce: reduce,
	}
}

type GreatestExecutor struct {
	tasks      <-chan Task
	errorLimit int
}

func (g GreatestExecutor) Execute(arg int) (int, error) {
	results := make(chan taskResult)
	defer close(results)

	var errCount int
	var resultVals []int
	wg := sync.WaitGroup{}

	go func() {
		for {
			tr, ok := <-results
			if !ok {
				fmt.Println("stops result listener")
				return
			}

			fmt.Println("has result", tr, ok)
			if tr.err != nil {
				errCount++
			}

			resultVals = append(resultVals, tr.result)
		}
	}()

	for task := range g.tasks {
		wg.Add(1)
		go func(task Task) {
			var tr taskResult
			tr.result, tr.err = task.Execute(arg)
			results <- tr
			wg.Done()
		}(task)
	}

	wg.Wait()
	fmt.Println("All tasks are done", errCount, resultVals)
	if errCount > g.errorLimit {
		return 0, fmt.Errorf("Error limit exceeded")
	}

	if len(resultVals) == 0 {
		return 0, fmt.Errorf("No tasks to execute")
	}

	var greatest int = resultVals[0]
	for _, res := range resultVals[1:] {
		if res > greatest {
			greatest = res
		}
	}

	return greatest, nil
}

func GreatestSearcher(errorLimit int, tasks <-chan Task) Task {
	return GreatestExecutor{
		tasks:      tasks,
		errorLimit: errorLimit,
	}
}
