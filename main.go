package main

import (
	"sync"
	"time"

	// "github.com/go-vgo/robotgo"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

var middlePressed = false
var lastClearTimer *time.Timer

// var lastMoveTimer *time.Timer
var startX int16

var startY int16

var lastClearTimerChan = make(chan int)

var rwMutex sync.RWMutex

func main() {

	var lastXChan = make(chan int16)
	var lastYChan = make(chan int16)

	hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
		// fmt.Println("numbers: ", runtime.NumGoroutine())
		if e.Button == 3 {
			if lastClearTimer != nil {
				if lastClearTimer.Stop() {
					lastClearTimerChan <- 1
				}
			}
			startX = e.X
			startY = e.Y
			middlePressed = true
			go func() {
				// 定时器关闭
				rwMutex.Lock()
				lastClearTimer = time.NewTimer(time.Millisecond * 400)
				rwMutex.Unlock()
				lastX := int16(0)
				lastY := int16(0)
				loop := true
				var exit = func() {
					middlePressed = false

					go func() {
						time.Sleep(time.Millisecond * 10)
						loop = false
						lastClearTimer = nil
					}()

				}
				var check = func() {
					if !middlePressed {
						return
					}
					finalX := lastX
					finalY := lastY
					arr := []string{"lctrl", "cmd"}
					needExit := false
					if finalX-startX >= 150 {
						needExit = true
						robotgo.KeyTap("right", arr)
					} else if startX-finalX >= 150 {
						needExit = true
						robotgo.KeyTap("left", arr)
					} else if startY-finalY >= 150 {
						needExit = true
						robotgo.KeyTap("tab", []string{"cmd"})
					} else if finalY-startY >= 150 {
						needExit = true
						robotgo.KeyTap("tab", []string{"lalt", "ctrl"})
					}
					if needExit {
						exit()
					}
				}

				for loop {
					select {
					case <-lastClearTimer.C:
						exit()
					case <-lastClearTimerChan:
						exit()
					case lastXIn := <-lastXChan:
						lastX = lastXIn
					case lastYIn := <-lastYChan:
						lastY = lastYIn
						check()
					}
				}
			}()
		}
	})

	hook.Register(hook.MouseMove, []string{}, func(e hook.Event) {
		if middlePressed {
			rwMutex.RLock()
			if lastClearTimer != nil {
				lastXChan <- e.X
				lastYChan <- e.Y
			}
			rwMutex.RUnlock()
		}
	})

	s := hook.Start()
	<-hook.Process(s)
}
