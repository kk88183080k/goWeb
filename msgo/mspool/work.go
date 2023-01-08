package mspool

import (
	"log"
	"time"
)

type Work struct {
	pool     *Pool       //协程池引用
	task     chan func() //要执行的任务队列
	lastTime time.Time   //执行任务的最后时间
}

func (w *Work) run() {
	w.pool.increaseByRunning()
	go w.running()
}

func (w *Work) running() {
	defer func() {
		//任务运行完成，worker空闲
		w.pool.decreaseByRunning()
		w.pool.workCache.Put(w)
		if r := recover(); r != nil {
			if w.pool.recoverHandler != nil {
				w.pool.recoverHandler(r)
			} else {
				log.Println("执行任务出错", r)
			}
		}
		w.pool.condition.Signal()
	}()

	// 一直轮询处理任务
	for v := range w.task {
		if v == nil { // 当接收到nil值时会退出
			// 退出时才放到
			return
		}

		v()
		//任务运行完成，worker空闲
		w.pool.putIdleWork(w)
	}
}
