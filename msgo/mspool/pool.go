package mspool

import (
	"errors"
	"github.com/kk88183080k/goWeb/msgo/msconf"
	"sync"
	"sync/atomic"
	"time"
)

// 任务默认过期时间
const defaultExpireTime = 3

var (
	capError        = errors.New("pool cap <0 ")
	expireTimeError = errors.New("pool expireTime <0")
	shutdownError   = errors.New("pool shutdown ")
)

// 关闭协程池的消息数据结构
type signal struct {
}

// recoverHandler work 错误处理函数
type recoverHandler func(v any)

type Pool struct {
	cap            int32          // 协程池容量
	running        int32          // 运行的协程数量
	idleWorkList   []*Work        // 空闲work列表
	expireTime     time.Duration  // 协程空闲多长时间后回收，单位为秒
	release        chan signal    // 关闭协程池的信号
	lock           sync.Mutex     // 保障协程池中资源并发操作安全
	one            sync.Once      // 协程池关闭操作只能执行一次
	workCache      sync.Pool      // 缓存work对象
	condition      *sync.Cond     // 通过信号来通知阻塞的协程
	recoverHandler recoverHandler //work 执行出错的错误处理函数
}

func NewPool(cap int, expireTime time.Duration) (*Pool, error) {
	if cap < 0 {
		return nil, capError
	}

	if expireTime < 0 {
		return nil, expireTimeError
	}

	p := &Pool{cap: int32(cap), running: 0, expireTime: expireTime * time.Second, release: make(chan signal, 1), lock: sync.Mutex{}, one: sync.Once{}}
	p.workCache.New = func() any {
		return &Work{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	p.condition = sync.NewCond(&p.lock)
	go p.timerClearWork()
	return p, nil
}

func NewDefaultPool() (*Pool, error) {
	cap := 10
	confCap, ok := msconf.Conf.Pool["cap"]
	if !ok {
		cap = confCap.(int)
	}

	var expireTime time.Duration
	confExpireTime, ok := msconf.Conf.Pool["expireTime"]
	if ok {
		expireTime = time.Duration(confExpireTime.(int64))
	} else {
		expireTime = defaultExpireTime
	}

	return NewPool(cap, expireTime)
}

func NewPoolCap(cap int) (*Pool, error) {
	return NewPool(cap, defaultExpireTime)
}

// timerClearTask 定时清理过期的空闲work
func (p *Pool) timerClearWork() {
	ticker := time.NewTicker(p.expireTime)
	for range ticker.C {
		if p.IsShutdown() {
			break
		}

		p.lock.Lock()
		workList := p.idleWorkList
		n := len(workList) - 1
		if n > 0 {
			clearN := -1
			for i, w := range workList {
				if time.Now().Sub(w.lastTime) <= p.expireTime {
					break
				}
				// 需要清理
				w.task <- nil
				workList[i] = nil
				clearN = i
			}
			if clearN != -1 { // 表示需要清理
				if clearN >= len(workList)-1 { // 全部清空
					p.idleWorkList = p.idleWorkList[:0]
				} else { // 清空之后
					p.idleWorkList = p.idleWorkList[clearN+1:]
				}
			}
		}

		//log.Println("清理work后：", p.idleWorkList, ",running:", p.running)
		p.lock.Unlock()
	}
}

func (p *Pool) Submit(fn func()) error {
	if p.IsShutdown() {
		return shutdownError
	}

	w := p.getWork()
	w.task <- fn
	return nil
}

func (p *Pool) getWork() *Work {
	// 1 获取协程池中的work
	// 2 如果协程池中有空闲的协程就直接返回
	p.lock.Lock()
	workList := p.idleWorkList
	n := len(workList) - 1
	if n >= 0 { // 代表有空闲
		work := workList[n]
		workList[n] = nil
		p.idleWorkList = workList[:n]
		p.lock.Unlock()
		return work
	}

	// 3 如果没有空间的线程就新建一个
	if p.running < p.cap {
		cacheWork := p.workCache.Get()
		var work *Work
		if cacheWork == nil {
			work = &Work{pool: p, task: make(chan func(), 1)}
		}
		work = cacheWork.(*Work)
		work.run()
		p.lock.Unlock()
		return work
	}

	// 4 如果正在运行的work数量大于池中最大的容量， 阻塞等待，直到work释放
	p.lock.Unlock()
	return p.waitIdleWork()
}

func (p *Pool) waitIdleWork() *Work {
	p.lock.Lock()
	p.condition.Wait()
	//log.Println("收到释放的通知")
	workList := p.idleWorkList
	n := len(workList) - 1
	if n < 0 {
		p.lock.Unlock()
		return p.waitIdleWork()
	}
	work := workList[n]
	workList[n] = nil
	p.idleWorkList = workList[:n]
	p.lock.Unlock()
	return work
}

func (p *Pool) putIdleWork(w *Work) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.idleWorkList = append(p.idleWorkList, w)
	p.lock.Unlock()
	// 通知等待的协程去获取空闲的work
	p.condition.Signal()
}

// IsShutdown 返回 true : 关闭 false: 运行中
func (p *Pool) IsShutdown() bool {
	return len(p.release) > 0
}

func (p *Pool) Shutdown() bool {
	if len(p.release) > 0 {
		return true
	}

	// 只执行一次
	p.one.Do(func() {
		p.lock.Lock()
		workList := p.idleWorkList
		for i, w := range workList {
			w.pool = nil
			w.task = nil
			workList[i] = nil
		}
		p.idleWorkList = nil
		p.release <- signal{}
		p.lock.Unlock()
	})
	return true
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}

	_ = <-p.release
	p.idleWorkList = make([]*Work, 1)

	return true
}

func (p *Pool) increaseByRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decreaseByRunning() {
	/*running := */ atomic.AddInt32(&p.running, -1)
	//log.Println("running", running)
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
