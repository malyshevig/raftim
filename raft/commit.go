package raft

import "raft/pq"

type CommitInfo struct {
	pq      pq.PriorityQueue
	pqItems map[int]*pq.Item
}

func InitCommittedInfo() *CommitInfo {
	commitInfo := CommitInfo{pq: pq.PriorityQueue{}, pqItems: make(map[int]*pq.Item)}

	return &commitInfo
}

func (cm *CommitInfo) addFollower(fv *FollowerInfo) {
	item := pq.MakeItem(fv, fv.nextIndex-1)
	cm.pq.Push(item)
	cm.pqItems[fv.id] = item
	cm.pq.Update(item, fv, fv.nextIndex-1)
}

func (cm *CommitInfo) updateFollowerIndex(fv *FollowerInfo) {
	item := cm.pqItems[fv.id]
	item.SetPriority(fv.nextIndex)
	cm.pq.Update(item, fv, fv.nextIndex-1)
}

func (cm *CommitInfo) GetNewCommitIndex() int {
	n := (cm.pq.Len() + 1) / 2
	fv := cm.pq[n-1].GetValue().(*FollowerInfo)
	newIndex := fv.nextIndex - 1

	return newIndex

}
