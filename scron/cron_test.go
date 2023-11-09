package scron

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const OneSecond = 1*time.Second + 50*time.Millisecond

type syncWriter struct {
	wr bytes.Buffer
	m  sync.Mutex
}

func TestMultiThreadedStartAndStop(t *testing.T) {
	cron := New()
	go cron.Run()
	time.Sleep(2 * time.Millisecond)
	cron.Stop()
}

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

func stop(cron *Cron) chan bool {
	ch := make(chan bool)
	go func() {
		cron.Stop()
		ch <- true
	}()
	return ch
}

// newWithSeconds returns a Cron with the seconds field enabled.
func newWithSeconds() *Cron {
	return New(WithParser(secondParser), WithChain())
}

func CronRun() {
	fmt.Println("this is run", time.Now().Format("2006.01.02 15:04:05"))
	time.Sleep(1 * time.Second)
}
func TestError(t *testing.T) {
	err := getErr()
	fmt.Println("lock Taskkey failed" == err.Error())
	fmt.Println(errors.Is(errors.New("lock_Taskkey_failed"), err))
}
func getErr() error {
	return errors.New("lock Taskkey failed")
}
func TestCron_Single_Entry_Operations(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		cron := New()
		_, err := cron.AddSingleton("*/5 * * * * *", CronRun, "cron-run")
		cron.Start()
		fmt.Println("this is test1", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
	t.Run("test3", func(t *testing.T) {
		cron := New()
		_, err := cron.AddSingleton("*/5 * * * * *", CronRun, "cron-run")
		cron.Start()
		fmt.Println("this is test1", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}
func TestCron_Single_Entry_Operations1(t *testing.T) {
	t.Run("test_six_spec_with_qu", func(t *testing.T) {
		cron := New()
		_, err := cron.AddSingleton("*/5 * * * * ?", CronRun, "cron-run")
		cron.Start()
		fmt.Println("this is test_six_spec_with_qu", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
	t.Run("test_six_spec_with_5", func(t *testing.T) {
		cron := New()
		_, err := cron.AddSingleton("*/5 * * * * ?", CronRun, "cron-run")
		cron.Start()
		fmt.Println("this is test_six_spec_with_qu", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}
func TestCron_Single_Entry_Operations2(t *testing.T) {
	t.Run("test_six_spec_with", func(t *testing.T) {
		cron := New()
		_, err := cron.AddSingleton("*/10 * * * * *", CronRun, "cron-run")
		cron.Start()
		fmt.Println("this is test_six_spec_with", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}

func TestCron_Single_Entry_Facade(t *testing.T) {
	t.Run("test_six_spec_facade", func(t *testing.T) {
		_, err := defaultCron.AddSingleton("*/10 * * * * *", CronRun, "cron-run")
		_, err = defaultCron.AddSingleton("*/10 * * * * *", CronRun, "cron-run1")
		defaultCron.Start()
		fmt.Println("this is test_six_spec_with", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}
func TestCron_Single_Entry_Facade1(t *testing.T) {
	t.Run("test_six_spec_facade1", func(t *testing.T) {
		_, err := defaultCron.AddSingleton("*/10 * * * * *", CronRun, "cron-run")
		_, err = defaultCron.AddSingleton("*/10 * * * * *", CronRun, "cron-run")
		defaultCron.Start()
		fmt.Println("this is test_six_spec_with", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}

func TestCron_FiveSpec(t *testing.T) {
	t.Run("test_six_spec_facade1", func(t *testing.T) {
		_, err := defaultCron.AddSingleton("*/1 * * * * *", CronRun, "cron-run")
		defaultCron.Start()
		fmt.Println("this is test_six_spec_with", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}

func TestCron_FiveSpecSecond(t *testing.T) {
	t.Run("test_six_spec_facade1", func(t *testing.T) {
		_, err := defaultCron.AddSingleton("*/1 * * * * *", CronRun, "cron-run")
		defaultCron.Start()
		fmt.Println("this is test_six_spec_with", time.Now().Format("2006.01.02 15:04:05"), err)
		time.Sleep(1 * time.Hour)
	})
}
