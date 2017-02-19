package taskexecutor

import "fmt"
import "testing"
import "time"

type adder struct {
	augend int
}

func (a adder) Execute(addend int) (int, error) {
	result := a.augend + addend
	if result > 127 {
		return 0, fmt.Errorf("Result %d exceeds the adder threshold", a)
	}
	return result, nil
}

func TestPipeline(t *testing.T) {
	if res, err := Pipeline(adder{50}, adder{60}).Execute(10); err != nil {
		t.Logf("The pipeline returned an error\n", err)
	} else {
		t.Logf("The pipeline returned %d\n", res)
	}
}

type lazyAdder struct {
	adder
	delay time.Duration
}

func (la lazyAdder) Execute(addend int) (int, error) {
	time.Sleep(la.delay * time.Millisecond)
	return la.adder.Execute(addend)
}

func TestFastest(t *testing.T) {
	f := Fastest(
		lazyAdder{adder{20}, 500},
		lazyAdder{adder{50}, 300},
		adder{41},
	)

	res, err := f.Execute(1)
	if res != 42 {
		t.Fatal(res, err)
	} else {
		t.Log(res)
	}
}

func TestTimed_Failed(t *testing.T) {
	_, e1 := Timed(lazyAdder{adder{20}, 50}, 2*time.Millisecond).Execute(2)
	if e1 == nil {
		t.Fatal("Need error")
	}
}

func TestTimed_Succes(t *testing.T) {
	r2, e2 := Timed(lazyAdder{adder{20}, 50}, 300*time.Millisecond).Execute(2)
	if r2 != 22 {
		t.Fatal(r2, e2)
	}
}

func TestConcurentMapReduce(t *testing.T) {
	reduce := func(results []int) int {
		smallest := 128
		for _, v := range results {
			if v < smallest {
				smallest = v
			}
		}
		return smallest
	}

	mr := ConcurrentMapReduce(reduce, adder{30}, adder{50}, adder{20})
	if res, err := mr.Execute(5); err != nil {
		t.Fatal("We got an error!\n", err)
	} else {
		t.Logf("The ConcurrentMapReduce returned %d\n", res)
	}
}

func TestGreatestSearcher(t *testing.T) {
	tasks := make(chan Task)
	gs := GreatestSearcher(2, tasks) // 2 errors are allowed

	go func() {
		tasks <- adder{4}
		tasks <- lazyAdder{adder{22}, 20}
		tasks <- adder{125} // introduce first error 125 + 10 > 127
		tasks <- adder{32}  // success

		// this should timeout and introduce second
		tasks <- Timed(lazyAdder{adder{100}, 2000}, 20*time.Millisecond)

		// If we uncomment the comment bellow we should have an error
		// from gs because we exceeded the error limit
		// 		tasks <- adder{127}
		close(tasks)
	}()
	result, err := gs.Execute(10)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("GreatestSearcher return values should be %d == 42", result)
	}
}
