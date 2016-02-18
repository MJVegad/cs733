package main

import (
	"fmt"
)

type AppendEntriesRespEv struct {	
		from uint64
		term uint64
		success bool	
}

func (sm *StateMachine) AppendEntriesRespEventHandler ( event interface{} ) (actions []interface{}) {
	cmd := event.(AppendEntriesRespEv)
	fmt.Printf("%v\n", cmd)
	switch sm.currentState {
		case "leader":
			if cmd.success == false {
				if sm.currentTerm < cmd.term {
					sm.currentTerm = cmd.term
					sm.currentState = "follower"
					actions = append(actions, StateStore{state: sm.currentState, term: sm.currentTerm, votedFor:sm.votedFor})
				} else {
					sm.nextIndex[cmd.from] = sm.nextIndex[cmd.from]-uint64(1)
					actions = append(actions, Send{peerId: cmd.from, ev: AppendEntriesReqEv{term: sm.currentTerm, leaderId: sm.serverId, prevLogIndex: sm.nextIndex[cmd.from]-uint64(1), prevLogTerm: sm.log[sm.nextIndex[cmd.from]-uint64(1)].term, entries: sm.log[sm.nextIndex[cmd.from]:], commitIndex: sm.commitIndex}})
				}
			} else {
				sm.nextIndex[cmd.from] = uint64(len(sm.log))
				// Update match index
			}
		case "follower":
			if cmd.term > sm.currentTerm {
				sm.currentTerm = cmd.term
				actions = append(actions, StateStore{state: sm.currentState, term: sm.currentTerm, votedFor:sm.votedFor})
			}	
		case "candidate":
		default: println("Invalid state")		
	}	
	return actions
}

