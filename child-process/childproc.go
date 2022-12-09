package childprocess

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

type ProcessID int

type Event struct {
	raised bool
}

func NewEvent() *Event {
	return &Event{
		raised: false,
	}
}
func (eve *Event)Wait() {
	for {
		if eve.raised  {
			return;
		} else {
			time.Sleep(10 * time.Millisecond);
		}
	}
}

func (eve *Event)IsInvoked() bool {
	return bool(eve.raised);
}
func (eve *Event)Raise() {
	eve.raised = true;
}




type ChildProcess struct {
	cmd *exec.Cmd

	shutdown *Event
	done *Event
}
type ChildProcesses struct {
	ready bool
	count int
	mutex sync.Mutex
	procs map[ProcessID]*ChildProcess

	out chan string

	ExitEvent chan ProcessID
}


func NewChildProcessSystem() *ChildProcesses {
	ret := ChildProcesses {
		ExitEvent: make(chan ProcessID),

		out: make(chan string,1000),

		procs: make(map[ProcessID]*ChildProcess),
		mutex: sync.Mutex{},
		count: 0,
		ready: true,
	}

	stdinSelf := os.Stdin;
	go func () {
		for {
			time.Sleep(1 * time.Second);
			_,err := stdinSelf.Write([]byte("\n"))
			if err != nil {
				return
			}
		}
	}()

	go func ()  {
		for{
			str := <-ret.out;
			go fmt.Printf("%s",str);
		}
	}()
	return &ret;
}



func (procs *ChildProcesses)handleProcess(id ProcessID){
	proc := procs.procs[id]

	processname := proc.cmd.Args[0]
	stdoutIn, _ := proc.cmd.StdoutPipe()
	stderrIn, _ := proc.cmd.StderrPipe()
	stdinOut, _ := proc.cmd.StdinPipe()
	go func () {
		for {
			time.Sleep(1 * time.Second);
			_,err := stdinOut.Write([]byte("\n"));
			if err != nil {
				return
			}
		}
	}()

	log := make([]byte, 0)
	for _, i := range proc.cmd.Args {
		log = append(log, append([]byte(i), []byte(" ")...)...)
	}
	fmt.Printf("starting %s : %s\n",processname , string(log))
	err := proc.cmd.Start()
	if err != nil{
		fmt.Printf("error init process %s\n",err.Error())
		return;
	}

	
	go procs.copyAndCapture(processname, stdoutIn)
	go procs.copyAndCapture(processname, stderrIn)

	go func() {
		proc.shutdown.Wait()
		if !proc.done.IsInvoked() {
			proc.done.Raise();
			proc.cmd.Process.Kill();
		}
	}();
	go func() {
		proc.cmd.Wait()
		if !proc.done.IsInvoked() {
			proc.done.Raise();
		}
	}()
	

	proc.done.Wait();
}


func (procs *ChildProcesses)NewChildProcess(cmd *exec.Cmd) ProcessID {
	if !procs.ready {
		return ProcessID(-1)
	}

	if cmd == nil {
		return -1;
	}


	procs.mutex.Lock();
	defer func ()  {
		procs.mutex.Unlock();
		procs.count++;
	} ()

	id := ProcessID(procs.count)
	procs.procs[id] = &ChildProcess{
		cmd: cmd,
		shutdown: NewEvent(),
		done: NewEvent(),
	}

	go func ()  {
		fmt.Printf("process %s, process id %d booting up\n",cmd.Args[0],int(id));
		procs.handleProcess(id)
		procs.ExitEvent <- id	
	} ()

	return ProcessID(procs.count)
}

func (procs *ChildProcesses)CloseAll() {
	procs.mutex.Lock();
	defer procs.mutex.Unlock();

	procs.ready = false;
	for _,proc := range procs.procs {
		proc.shutdown.Raise()
	}
}

func (procs *ChildProcesses)CloseID(ID ProcessID) {
	procs.mutex.Lock();
	defer procs.mutex.Unlock();


	proc := procs.procs[ID]
	if proc == nil {
		return;
	}

	fmt.Printf("force terminate process name %s, process id %d \n",proc.cmd.Args[0],int(ID));
	proc.shutdown.Raise()
}

func (procs *ChildProcesses)WaitID(ID ProcessID) {
	for {
		id := <-procs.ExitEvent
		if id == ID {
			fmt.Printf("process name %s with id %d exited \n",procs.procs[ID].cmd.Args[0],int(ID));
			return;	
		} else {
			procs.ExitEvent<-id;
			time.Sleep(10 * time.Millisecond)
		}
	}
}























func findLineEnd(dat []byte) (out [][]byte) {
	prev := 0
	for pos, i := range dat {
		if i == []byte("\n")[0] {
			out = append(out, dat[prev:pos])
			prev = pos + 1
		}
	}

	out = append(out, dat[prev:])
	return
}

func (procs *ChildProcesses)copyAndCapture(process string, r io.Reader) {
	prefix := []byte(fmt.Sprintf("Child process (%s): ", process))
	after := []byte("\n")

	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return
		}

		if n > 0 {
			d := buf[:n]
			lines := findLineEnd(d)
			for _, line := range lines {
				out := append(prefix, line...)
				out = append(out, after...)

				procs.out <- string(out);
			}
		}
	}
}
