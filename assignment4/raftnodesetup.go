package main

import (
	"encoding/gob"
	"fmt"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/log"
	"math"
	"time"
	"os"
	"bufio"
	"strconv"
	//"reflect"
)

type CommitInfo struct {
	Data  []byte
	Index int64
	Err   error
}

type Config struct {
	cluster          []NetConfig
	Id               int64
	LogDir           string
	ElectionTimeout  int64
	HeartbeatTimeout int64
}

type NetConfig struct {
	Id   int64
	Host string
	Port int64
}

func assert(val bool) {
	if !val {
		panic("Assertion Failed")
	}
}

type NodePers struct {
	CurrentTerm  int64
	VotedFor     int64
	CurrentState string
}

type RaftNode struct {
	sm           StateMachine
	sm_messaging cluster.Server
	logfile      string
	statefile    *os.File
	resettimer   int64
	eventch      chan interface{}
	//timeoutch    chan TimeoutEv
	commitch chan CommitInfo
	endch    chan bool
	timer    *time.Timer
}

var lg *log.Log

func (rn *RaftNode) processEvents() {
	for {
		var ev interface{}
		select {
		case ev = <-rn.eventch:
		case envelop := <-rn.sm_messaging.Inbox():
			switch envelop.Msg.(type) {
			case AppendEv:
				rn.eventch <- envelop.Msg.(AppendEv)
			case AppendEntriesReqEv:
				//fmt.Printf("follower->%v, received log entries->%v\n",rn.sm.serverId, envelop.Msg.(AppendEntriesReqEv).Entries)
				rn.eventch <- envelop.Msg.(AppendEntriesReqEv)
			case AppendEntriesRespEv:
				rn.eventch <- envelop.Msg.(AppendEntriesRespEv)
			case VoteReqEv:
				rn.eventch <- envelop.Msg.(VoteReqEv)
			case VoteRespEv:
				rn.eventch <- envelop.Msg.(VoteRespEv)
			}
			//fmt.Printf("%v received message %v\n", rn.sm.serverId, envelop.Msg)
			continue
		case <-rn.timer.C:
			{
				ev = TimeoutEv{}
			}
		case <-rn.endch:
			return

		}
		actions := rn.sm.ProcessEvent(ev)
		rn.doActions(actions)
	}
}

func New(config Config, jsonFile string) (rnode RaftNode) {
	rnode.sm.serverId = config.Id

	rnode.sm.peerIds = make([]int64, len(config.cluster)-1, len(config.cluster)-1)
	k := 0
	for _, peer := range config.cluster {

		if peer.Id != config.Id {
			rnode.sm.peerIds[k] = peer.Id
			k++
		}
	}

	rnode.sm.majority = int64(math.Ceil(float64(len(config.cluster)) / 2.0))

	rnode.sm.commitIndex = -1

	rnode.logfile = config.LogDir + "/" + "logfile"
	lg, _ = log.Open(rnode.logfile)
	//fmt.Printf("type of log:\n",reflect.TypeOf(lg))
	lg.RegisterSampleEntry(logEntry{})
	//assert(err == nil)
	//defer lg.Close()
	if lg.GetLastIndex() == -1 {
		rnode.sm.log = []logEntry{}
	} else {
		i := lg.GetLastIndex()

		for j := int64(0); j <= i; j++ {
			data, _ := lg.Get(j)
			rnode.sm.log = append(rnode.sm.log, data.(logEntry))
		}
	}

	rnode.sm.nextIndex = make([]int64, len(config.cluster)-1, len(config.cluster)-1)
	for n := 0; n < len(config.cluster)-1; n++ {
		rnode.sm.nextIndex[n] = int64(len(rnode.sm.log))
	}

	rnode.sm.matchIndex = make([]int64, len(config.cluster)-1, len(config.cluster)-1)
	for m := 0; m < len(config.cluster)-1; m++ {
		rnode.sm.matchIndex[m] = -1
	}

	//rnode.statefile = config.LogDir + "/" + "statefile"
	
	//currstate, err := log.Open(rnode.statefile)
	
	//if _,err5:=os.Stat(config.LogDir + "/" + "statefile");os.IsNotExist(err5) {
		/*	rnode.statefile,_ = os.Create(config.LogDir + "/" + "statefile")
			w:=bufio.NewWriter(rnode.statefile)
			_,err:=fmt.Fprintf(w,"%s %s %s\n", "follower","0","0")
			if err!=nil {
				fmt.Printf("statefile write error:%v\n",err)
			}
			fmt.Printf("statefile created\n")
			w.Flush()
			rnode.sm.currentTerm = int64(0)
			rnode.sm.currentState = "follower"*/
		
	//} else {
			var err9 error
			rnode.statefile,err9= os.OpenFile(config.LogDir + "_" + "statefile", os.O_RDWR, 0666)
			if err9 != nil {
		fmt.Println(err9)
	}
			r := bufio.NewReader(rnode.statefile)
			var currentState, currentTerm, votedFor string
			_, err := fmt.Fscanf(r, "%s %s %s\n", &currentState, &currentTerm, &votedFor)
			if err!=nil {
				fmt.Printf("statefile read error:%v\n",err)
			}
			s1,_:=strconv.Atoi(currentTerm)
			s2,_:=strconv.Atoi(votedFor)
			rnode.sm.currentTerm = int64(s1)
			rnode.sm.currentState = currentState
			rnode.sm.votedFor = int64(s2)
		
	//}	
	//currstate.RegisterSampleEntry(NodePers{})
	//assert(err == nil)
	//defer currstate.Close()
	/*if currstate.GetLastIndex() == -1 {
		rnode.sm.currentTerm = int64(0)
		rnode.sm.currentState = "follower"
	} else {
		i := currstate.GetLastIndex()
		h, _ := currstate.Get(i)
		rnode.sm.currentTerm = h.(NodePers).CurrentTerm
		rnode.sm.currentState = h.(NodePers).CurrentState
		rnode.sm.votedFor = h.(NodePers).VotedFor
	}*/

	rnode.sm.totalvotes = int64(0)
	rnode.sm.novotes = int64(0)

	rnode.eventch = make(chan interface{}, 5000)
	rnode.commitch = make(chan CommitInfo, 1000)
	rnode.endch = make(chan bool)
	//rnode.timeoutch = make(chan TimeoutEv)
	//rnode.resettimer = 0

	rnode.sm.ElectionTimeout = config.ElectionTimeout
	rnode.sm.HeartbeatTimeout = config.HeartbeatTimeout

	gob.Register(AppendEv{})
	gob.Register(AppendEntriesReqEv{})
	gob.Register(AppendEntriesRespEv{})
	gob.Register(TimeoutEv{})
	gob.Register(VoteReqEv{})
	gob.Register(VoteRespEv{})

	rnode.timer = time.NewTimer(time.Duration(config.ElectionTimeout) * time.Millisecond)
	var err3 error
	rnode.sm_messaging, err3 = cluster.New(int(config.Id), jsonFile)

	if err3 != nil {
		fmt.Printf("Error in sm_messaging.")
	}
	return
}

func (rnode *RaftNode) StateStoreHandler(obj StateStore) {
	/*lg, err := log.Open(rnode.statefile)
	lg.RegisterSampleEntry(NodePers{})
	assert(err == nil)
	defer lg.Close()
	i := lg.GetLastIndex()
	lg.TruncateToEnd(i)
	lg.Append(NodePers{CurrentTerm: obj.term, VotedFor: obj.votedFor, CurrentState: obj.state})*/
	w:=bufio.NewWriter(rnode.statefile)
			_,err:=fmt.Fprintf(w,"%s %s %s\n", obj.state,obj.state,strconv.Itoa(int(obj.term)),strconv.Itoa(int(obj.votedFor)))
			if err!=nil {
				fmt.Printf("statefile write error:%v\n",err)
			}
			w.Flush()
			//rnode.sm.currentTerm = int64(0)
			//rnode.sm.currentState = "follower"
}

func (rnode *RaftNode) LogStoreHandler(obj LogStore) {
	//fmt.Printf("LogstoreBegin: index->%d, command->%v\n", obj.index, obj.command)
	//lg, err := log.Open(rnode.logfile)
	//lg.RegisterSampleEntry(logEntry{})
	//assert(err == nil)
	//defer lg.Close()
	lg.TruncateToEnd(int64(obj.index))
	lg.Append(obj.command)
	//fmt.Printf("LogstoreEnd: index->%d, command->%v\n", obj.index, obj.command)
}

func (rnode *RaftNode) AlarmHandler(obj Alarm) {
	rnode.timer.Reset(time.Duration(obj.t) * time.Millisecond)
}

func (rnode *RaftNode) CommitHandler(obj Commit) {
	//fmt.Printf("%v In CommitHandler: %v\n", rnode.sm.serverId, obj)
	t1 := CommitInfo{Data: obj.command, Index: obj.index, Err: obj.err}
	rnode.commitch <- t1
	//fmt.Printf("On %v -> Commitchannel: %v\n", rnode.sm.serverId, t1.Err)
}

func (rnode *RaftNode) SendHandler(obj Send) {
	//fmt.Printf("%v In send handler: %v, %v \n", rnode.Id(), reflect.TypeOf(obj.ev), obj)
	switch obj.ev.(type) {
	case TimeoutEv:
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: TimeoutEv{}}
	case AppendEv:
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: AppendEv{Data: obj.ev.(AppendEv).Data}}
	case AppendEntriesReqEv:
		// rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: AppendEntriesReqEv{Term: obj.ev.(AppendEntriesReqEv).Term, LeaderId: obj.ev.(AppendEntriesReqEv).LeaderId, PrevLogIndex: obj.ev.(AppendEntriesReqEv).PrevLogIndex, PrevLogTerm: obj.ev.(AppendEntriesReqEv).PrevLogTerm, Entries: obj.ev.(AppendEntriesReqEv).Entries, CommitIndex: obj.ev.(AppendEntriesReqEv).CommitIndex}}
		//fmt.Printf("Leader->%v, Sent log entries->%v\n",rnode.sm.serverId, obj.ev.(AppendEntriesReqEv).Entries)
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: obj.ev.(AppendEntriesReqEv)}
		//fmt.Printf("%v New send handler %v\n", rnode.sm.serverId, obj.ev.(AppendEntriesReqEv).Entries)
	case AppendEntriesRespEv:
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: AppendEntriesRespEv{From: obj.ev.(AppendEntriesRespEv).From, Term: obj.ev.(AppendEntriesRespEv).Term, Success: obj.ev.(AppendEntriesRespEv).Success, Lastindex: obj.ev.(AppendEntriesRespEv).Lastindex}}
	case VoteReqEv:
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: VoteReqEv{Term: obj.ev.(VoteReqEv).Term, CandidateId: obj.ev.(VoteReqEv).CandidateId, LastLogIndex: obj.ev.(VoteReqEv).LastLogIndex, LastLogTerm: obj.ev.(VoteReqEv).LastLogTerm}}
	case VoteRespEv:
		rnode.sm_messaging.Outbox() <- &cluster.Envelope{Pid: int(obj.peerId), Msg: VoteRespEv{Term: obj.ev.(VoteRespEv).Term, VoteGranted: obj.ev.(VoteRespEv).VoteGranted}}
	default:
		println("unrecognized event")
	}
}

func (rn *RaftNode) doActions(actions []interface{}) {
	for _, action := range actions {
		switch action.(type) {
		case StateStore:
			rn.StateStoreHandler(action.(StateStore))
		case LogStore:
			rn.LogStoreHandler(action.(LogStore))
		case Alarm:
			rn.AlarmHandler(action.(Alarm))
		case Commit:
			rn.CommitHandler(action.(Commit))
			//fmt.Printf("Node->%v, data to commit: %v -> %v\n", rn.sm.serverId, action.(Commit).index, action.(Commit).command)
		case Send:
			rn.SendHandler(action.(Send))
		default:
			println("unrecognized action")
		}
	}

}
