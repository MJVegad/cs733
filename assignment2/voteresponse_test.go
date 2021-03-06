package main

import (
	"testing"
)

func TestCandidateVoteResponse1 (t *testing.T) {
	//when candidate becomes leader
	sm := StateMachine {serverId: uint64(1), peerIds: []uint64{uint64(2),uint64(3),uint64(4),uint64(5)}, 
		majority: uint64(3), commitIndex: uint64(1), nextIndex: []uint64{uint64(2),uint64(2),uint64(2),uint64(2)}, 
		matchIndex: []uint64{uint64(1),uint64(1),uint64(1),uint64(1)}, 
		log: []logEntry{logEntry{term: 1, command: []byte("add")},logEntry{term: 2, command: []byte("disp")}}, 
		currentTerm: 2, votedFor: 1, currentState: "candidate", totalvotes: 2}

	result := sm.ProcessEvent(VoteRespEv{term: 2, voteGranted: true})

	exsm := StateMachine {serverId: uint64(1), peerIds: []uint64{uint64(2),uint64(3),uint64(4),uint64(5)}, 
		majority: uint64(3), commitIndex: uint64(1), nextIndex: []uint64{uint64(2),uint64(2),uint64(2),uint64(2)}, 
		matchIndex: []uint64{uint64(0),uint64(0),uint64(0),uint64(0)}, 
		log: []logEntry{logEntry{term: 1, command: []byte("add")},logEntry{term: 2, command: []byte("disp")}}, 
		currentTerm: 2, votedFor: 1, currentState: "leader", totalvotes: 3}

	exactions := []interface{}{StateStore{state: "leader", term: 2, votedFor:1}, Alarm{t: 100},
		Send{peerId: 2, ev: AppendEntriesReqEv{term: 2, leaderId: 1, prevLogIndex: 0, 
		prevLogTerm: 1, commitIndex: 1}}, 
		Send{peerId: 3, ev: AppendEntriesReqEv{term: 2, leaderId: 1, prevLogIndex: 0, prevLogTerm: 1, commitIndex: 1}}, 
		Send{peerId: 4, ev: AppendEntriesReqEv{term: 2, leaderId: 1, prevLogIndex: 0, prevLogTerm: 1, commitIndex: 1}}, 
		Send{peerId: 5, ev: AppendEntriesReqEv{term: 2, leaderId: 1, prevLogIndex: 0, prevLogTerm: 1, commitIndex: 1}}}

	//expect (t, result, excpectedPeerIds)
	ExpectStateMachine(t, &sm, &exsm)	
	ExpectActions (t, result, exactions)
	
}


func TestCandidateVoteResponse2 (t *testing.T) {
	//when candidate changes state to follower due to novotes
	sm := StateMachine {serverId: uint64(1), peerIds: []uint64{uint64(2),uint64(3),uint64(4),uint64(5)}, 
		majority: uint64(3), commitIndex: uint64(1), nextIndex: []uint64{uint64(2),uint64(2),uint64(2),uint64(2)}, 
		matchIndex: []uint64{uint64(1),uint64(1),uint64(1),uint64(1)}, 
		log: []logEntry{logEntry{term: 1, command: []byte("add")},logEntry{term: 2, command: []byte("disp")}}, 
		currentTerm: 2, votedFor: 1, currentState: "candidate", totalvotes: 2, novotes: 2}

	result := sm.ProcessEvent(VoteRespEv{term: 2, voteGranted: false})

	exsm := StateMachine {serverId: uint64(1), peerIds: []uint64{uint64(2),uint64(3),uint64(4),uint64(5)}, 
		majority: uint64(3), commitIndex: uint64(1), nextIndex: []uint64{uint64(2),uint64(2),uint64(2),uint64(2)}, 
		matchIndex: []uint64{uint64(1),uint64(1),uint64(1),uint64(1)}, 
		log: []logEntry{logEntry{term: 1, command: []byte("add")},logEntry{term: 2, command: []byte("disp")}}, 
		currentTerm: 2, votedFor: 1, currentState: "follower", totalvotes: 2, novotes: 3}

	exactions := []interface{}{Alarm{t: 100}, StateStore{state: "follower", term: 2, votedFor:1}}
		
	ExpectStateMachine(t, &sm, &exsm)	
	ExpectActions (t, result, exactions)
	
} 

