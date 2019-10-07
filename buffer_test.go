package xload

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type n int

func (i n) Pick(collection interface{}) interface{} {
	for _, v := range collection.([]int) {
		if int(i) == v {
			return v
		}
	}
	return nil
}

func bufferizeNs(ctx context.Context, interfaces ...interface{}) (interface{}, error) {
	m := make(map[int]struct{})
	var out []int
	for _, f := range interfaces {
		var i int
		switch t := f.(type) {
		case n:
			i = int(t)
		case int:
			i = t
		}
		if _, ok := m[i]; ok {
			continue
		}
		m[i] = struct{}{}
		out = append(out, i)
	}
	atomic.AddInt32(ctx.Value("cycles").(*int32), 1)
	return out, nil
}

func TestBuffer(t *testing.T) {

	type Job struct {
		Input  interface{}
		Output int
		Err    error
	}

	type Given struct {
		Inputs [][]Job
	}

	type Want struct {
		Outputs []int
		Cycles  int
	}

	type Case struct {
		Sentence string
		Given
		Want
	}

	var cases = []Case{
		{
			Sentence: "When time is the limit",
			Given: Given{
				Inputs: [][]Job{
					[]Job{Job{Input: n(1)}, Job{Input: n(1)}, Job{Input: n(2)}},
					[]Job{Job{Input: n(3)}, Job{Input: n(3)}},
					[]Job{Job{Input: n(1)}, Job{Input: n(5)}},
				},
			},
			Want: Want{
				Outputs: []int{1, 1, 2, 3, 3, 1, 5},
				Cycles:  2,
			},
		},
		{
			Sentence: "When size is the limit",
			Given: Given{
				Inputs: [][]Job{
					[]Job{Job{Input: n(1)}, Job{Input: n(1)}, Job{Input: n(2)}},
					[]Job{Job{Input: n(3)}, Job{Input: n(3)}, Job{Input: n(1)}, Job{Input: n(5)}},
					[]Job{Job{Input: n(7)}, Job{Input: n(1)}, Job{Input: n(9)}, Job{Input: n(7)}, Job{Input: n(1)}, Job{Input: n(9)}},
				},
			},
			Want: Want{
				Outputs: []int{1, 1, 2, 3, 3, 1, 5, 7, 1, 9, 7, 1, 9},
				Cycles:  3,
			},
		},
		{
			Sentence: "When the timer expires",
			Given: Given{
				Inputs: [][]Job{
					[]Job{Job{Input: n(1)}, Job{Input: n(1)}, Job{Input: n(2)}},
				},
			},
			Want: Want{
				Outputs: []int{1, 1, 2},
				Cycles:  1,
			},
		},
		{
			Sentence: "When doing requests after timer expires",
			Given: Given{
				Inputs: [][]Job{
					[]Job{},
					[]Job{},
					[]Job{},
					[]Job{},
					[]Job{},
					[]Job{Job{Input: n(1)}, Job{Input: n(1)}, Job{Input: n(2)}},
				},
			},
			Want: Want{
				Outputs: []int{},
				Cycles:  0,
			},
		},
	}

	const (
		InputLag    = time.Millisecond * 20
		BufferFreq  = time.Millisecond * 30
		CancelAfter = time.Millisecond * 80
		BufferSize  = 6
	)

	formatJobs := func(jobs []Job) string {
		vals := make([]string, len(jobs))
		for i, v := range jobs {
			vals[i] = strconv.Itoa(int(v.Input.(n)))
		}
		return strings.Join(vals, ", ")
	}

	for _, c := range cases {
		t.Run(c.Sentence, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), CancelAfter)
			ctx = context.WithValue(ctx, "cycles", new(int32)) // nolint
			defer cancel()

			buffer := NewBuffer(ctx, bufferizeNs, BufferSize, BufferFreq)
			var wg sync.WaitGroup
			for i := range c.Given.Inputs {
				for j := range c.Given.Inputs[i] {
					wg.Add(1)
					go func(i, j int) {
						defer wg.Done()
						time.Sleep(time.Duration(i) * InputLag)
						v, err := buffer.Do(c.Given.Inputs[i][j].Input)
						out, _ := v.(int)
						c.Given.Inputs[i][j].Output, c.Given.Inputs[i][j].Err = out, err
					}(i, j)
				}
			}
			wg.Wait()

			t.Log("as a buffer, given")
			t.Log("	a size of", BufferSize)
			t.Log("	a freq of", BufferFreq)
			t.Logf("	%d batches of inputs\n", len(c.Inputs))
			for i, jobs := range c.Inputs {
				t.Logf("		%s (%v)\n", formatJobs(jobs), InputLag*time.Duration(i))
			}
			t.Log("	a timeout of", CancelAfter)

			cycles := int(atomic.LoadInt32(ctx.Value("cycles").(*int32)))
			if cycles != c.Want.Cycles {
				t.Errorf("should run %d cycles, %d done", c.Want.Cycles, cycles)
			}

			k := -1
			for x, jobs := range c.Inputs {
				for y, j := range jobs {
					k++
					if len(c.Want.Outputs) > k {
						got := j.Output
						want := c.Want.Outputs[k]
						if want != got {
							t.Errorf("Inputs.Job[%d][%d].Output should be %d, got %d\n", x, y, want, got)
						}
						continue
					}

					if j.Err == nil {
						t.Errorf("Inputs.Job[%d][%d].Err should not be nil\n", x, y)
					}
				}
			}
		})
	}
}
