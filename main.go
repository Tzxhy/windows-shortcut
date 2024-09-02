package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	// "github.com/go-vgo/robotgo"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

var middlePressed = false
var lastClearTimer *time.Timer

var lastMoveTimer *time.Timer
var startX int16

var startY int16

var lastClearTimerChan = make(chan int)

var rwMutex sync.RWMutex

func main() {
	hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
		fmt.Println("numbers: ", runtime.NumGoroutine())
		if e.Button == 3 {
			if lastClearTimer != nil {
				if lastClearTimer.Stop() {
					lastClearTimerChan <- 1
				}
			}
			middlePressed = true
			go func() {
				// 定时器关闭
				lastClearTimer = time.NewTimer(time.Second * 1)
				select {
				case <-lastClearTimer.C:
					middlePressed = false
				case <-lastClearTimerChan:
				}
			}()
		}
	})

	var lastXChan = make(chan int16)
	var lastYChan = make(chan int16)

	hook.Register(hook.MouseMove, []string{}, func(e hook.Event) {
		if middlePressed {
			if lastMoveTimer == nil {
				startX = e.X
				startY = e.Y
				lastMoveTimer = time.NewTimer(time.Millisecond * 400)
				go func() {
					lastX := int16(0)
					lastY := int16(0)
					var loop = true
					for loop {
						select {
						case <-lastMoveTimer.C:
							rwMutex.Lock()
							defer rwMutex.Unlock()

							finalX := lastX
							finalY := lastY
							if finalX == 0 || finalY == 0 {
								loop = false
								lastMoveTimer = nil
								middlePressed = false
								return
							}
							arr := []string{"lctrl", "cmd"}

							if finalX-startX >= 150 {
								robotgo.KeyTap("right", arr)
								fmt.Println("right")
							} else if startX-finalX >= 150 {
								robotgo.KeyTap("left", arr)
								fmt.Println("left")
							} else if startY-finalY >= 150 {
								robotgo.KeyTap("tab", []string{"cmd"})
							} else if finalY-startY >= 150 {
								robotgo.KeyTap("tab", []string{"lalt", "ctrl"})
							}
							lastMoveTimer = nil
							middlePressed = false

							loop = false
							return
						case lastXIn := <-lastXChan:
							lastX = lastXIn
						case lastYIn := <-lastYChan:
							lastY = lastYIn

						}
					}
				}()
			} else {
				rwMutex.RLock()
				if lastMoveTimer != nil {
					lastXChan <- e.X
					lastYChan <- e.Y
				}
				rwMutex.RUnlock()
			}
		}
	})

	s := hook.Start()
	<-hook.Process(s)
}
