package calm

import (
	"log"
)

var Ppool *WorkerPool

type Job interface {
	Do()
}

type Worker struct {
	JobQueue chan Job
	Quit     chan bool
}

func init() {
	num := 100 * 100
	Ppool = NewWorkerPool(num)
	Ppool.Run()
}

func NewWorker() Worker {
	return Worker{
		JobQueue: make(chan Job),
		Quit:     make(chan bool),
	}
}

/*
整个过程中 每个Worker(工人)都会被运行在一个协程中，
在整个WorkerPool(领导)中就会有num个可空闲的Worker(工人)，
当来一条数据的时候，领导就会小组中取一个空闲的Worker(工人)去执行该Job，
当工作池中没有可用的worker(工人)时，就会阻塞等待一个空闲的worker(工人)。
每读到一个通道参数 运行一个 worker
*/

func (w Worker) Run(workqueue chan chan Job) {
	//这是一个独立的协程 循环读取通道内的数据，
	//保证 每读到一个通道参数就 去做这件事，没读到就阻塞
	go func() {
		for {
			workqueue <- w.JobQueue //注册空闲工作通道重新放到到工作线程池()
			select {
			case job := <-w.JobQueue: //读到参数
				job.Do()
			case <-w.Quit: //终止当前任务
				log.Println("Quit worker")
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}

// workerpool 领导
type WorkerPool struct {
	workerlen   int      //线程池中  worker(工人) 的数量
	JobQueue    chan Job //线程池的  job 通道
	WorkerQueue chan chan Job
}

func NewWorkerPool(workerlen int) *WorkerPool {
	return &WorkerPool{
		workerlen:   workerlen,                      //开始建立 workerlen 个worker(工人)协程
		JobQueue:    make(chan Job),                 //工作队列 通道
		WorkerQueue: make(chan chan Job, workerlen), //最大通道参数设为 最大协程数 workerlen 工人的数量最大值
	}
}

// 运行线程池
func (wp *WorkerPool) Run() {
	//初始化时会按照传入的num，启动num个后台协程，然后循环读取Job通道里面的数据，
	//读到一个数据时，再获取一个可用的Worker，并将Job对象传递到该Worker的chan通道
	log.Println("初始化worker:", wp.workerlen)
	for i := 0; i < wp.workerlen; i++ {
		//新建 workerlen 20万 个 worker(工人) 协程(并发执行)，每个协程可处理一个请求
		worker := NewWorker() //运行一个协程 将线程池 通道的参数  传递到 worker协程的通道中 进而处理这个请求
		worker.Run(wp.WorkerQueue)
	}

	// 循环获取可用的worker,往worker中写job
	go func() { //这是一个单独的协程 只负责保证 不断获取可用的worker
		for {
			select {
			case job := <-wp.JobQueue: //读取任务
				//尝试获取一个可用的worker作业通道。
				//这将阻塞，直到一个worker空闲

				//获取工作线程池()
				worker1 := <-wp.WorkerQueue
				worker1 <- job //将任务 分配给该工人

			}
		}
	}()
}
