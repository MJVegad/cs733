package main

import (
	//"fmt"
)

type AppendEntriesReqEv struct {
		term uint64
		leaderId uint64
		prevLogIndex uint64
		prevLogTerm uint64
		entries []logEntry
		commitIndex uint64		
}

func (sm *StateMachine) AppendEntriesReqEventHandler ( event interface{} ) (actions []interface{}){
	cmd := event.(AppendEntriesReqEv)
	//fmt.Printf("%v\n", cmd)
	switch sm.currentState {
		case "leader":
			if sm.currentTerm <= cmd.term {
				if sm.currentTerm < cmd.term {
					sm.votedFor = 0
				}
				sm.currentTerm = cmd.term
				sm.currentState = "follower"
			//	actions = append(actions, StateStore{state: sm.currentState, term: sm.currentTerm, votedFor:sm.votedFor}) // make it function
			//	actions = append(actions, Send{peerId: sm.serverId, ev: AppendEntriesReqEv{term: cmd.term, leaderId: cmd.leaderId, prevLogIndex: cmd.prevLogIndex, prevLogTerm: cmd.prevLogTerm, entries: cmd.entries, commitIndex: cmd.commitIndex}}) //make it function
				actions = sm.AppendEntriesReqEventHandler(event) 
			} else {
				actions = append(actions, Send{peerId: cmd.leaderId, ev: AppendEntriesRespEv{from: sm.serverId, term: sm.currentTerm, success: false}})
			}
		case "follower":
			
			if sm.currentTerm <= cmd.term {
				if sm.currentTerm < cmd.term {
					sm.votedFor = 0
				}
				sm.currentTerm = cmd.term
				actions = append(actions, Alarm{t: 100})
				actions = append(actions, StateStore{state: sm.currentState, term: sm.currentTerm, votedFor:sm.votedFor})
				if ( (cmd.prevLogTerm == 0) || ( cmd.prevLogIndex < uint64(len(sm.log)) && (sm.log[cmd.prevLogIndex].term == cmd.prevLogTerm)) ) {
					k:=0
					
					templog := make([]logEntry, cmd.prevLogIndex+uint64(1)+uint64(len(cmd.entries)), cmd.prevLogIndex+uint64(1)+uint64(len(cmd.entries)))
					for i:=uint64(0);i<cmd.prevLogIndex+uint64(1);i++ {
						templog[i]=sm.log[i]
					}
					
					for j:=cmd.prevLogIndex+uint64(1);j<uint64(cmd.prevLogIndex+uint64(1)+uint64(len(cmd.entries)));j++ {
						templog[j] = cmd.entries[k]
						actions = append(actions, LogStore{index: j, command: cmd.entries[k]})
						k++
					}
					sm.log = templog
					//actions = append(actions, LogStore{index: uint64(len(sm.log)-1), command: sm.log[uint64(len(sm.log)-1)]})
					actions = append(actions, Send{peerId: cmd.leaderId, ev: AppendEntriesRespEv{from: sm.serverId, term: cmd.term, success: true}})
					if cmd.commitIndex > sm.commitIndex {
						if uint64(len(sm.log)-1) < sm.commitIndex {
							sm.commitIndex = uint64(len(sm.log)-1)
						}						
					}
				} else {
					actions = append(actions, Send{peerId: cmd.leaderId, ev: AppendEntriesRespEv{from: sm.serverId, term: cmd.term, success: false}})
				}
				
	   		} else {
	   			actions = append(actions, Send{peerId: cmd.leaderId, ev: AppendEntriesRespEv{from: sm.serverId, term: sm.currentTerm, success: false}})
	   		}
	   	case "candidate":
	   			if sm.currentTerm <= cmd.term {
	   				if sm.currentTerm < cmd.term {
					sm.votedFor = 0
				    }
					sm.currentTerm = cmd.term
					sm.currentState = "follower"
					actions = sm.AppendEntriesReqEventHandler(event)
				} else {
					actions = append(actions, Send{peerId: cmd.leaderId, ev: AppendEntriesRespEv{from: sm.serverId, term: sm.currentTerm, success: false}})
				}	
		default: println("Invalid state")		
}
	return actions
}