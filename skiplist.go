package skiplist

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

const MaxLevel = 64

type SkipList struct {
	Header *SkipNode
	Tail   *SkipNode // 跳跃表尾节点
	Level  int       // 最大的层数
	Len    uint64    // 节点数量
}

type SkipLevel struct {
	Forward *SkipNode
	Span    uint64 // 与 Forward 节点的跨度（距离）
}

type SkipNode struct {
	Level    []SkipLevel
	Backward *SkipNode
	Score    float64
	Value    string
}

func createNode(level int, score float64, value string) *SkipNode {
	return &SkipNode{
		Level:    make([]SkipLevel, level),
		Backward: nil,
		Score:    score,
		Value:    value,
	}
}

// Range 范围，左右闭区间[Min, Max]
type Range struct {
	Min float64
	Max float64
}

func (r *Range) GteMin(score float64) bool {
	return score >= r.Min
}

func (r *Range) LetMax(score float64) bool {
	return score <= r.Max
}

func New() *SkipList {
	return &SkipList{
		Level:  1,
		Len:    0,
		Header: createNode(MaxLevel, 0, ""),
		Tail:   nil,
	}
}

// Insert 插入元素，注意：该方法内没做元素唯一性检测
func (sl *SkipList) Insert(score float64, value string) {
	update := make([]*SkipNode, MaxLevel) // 存储各层的前置节点
	rank := make([]uint64, MaxLevel)      // 存储各层前置节点的排名

	// 找到前一个节点的位置及排名
	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		if i != sl.Level-1 {
			rank[i] = rank[i+1]
		}
		for x.Level[i].Forward != nil && (x.Level[i].Forward.Score < score ||
			(x.Level[i].Forward.Score == score && x.Level[i].Forward.Value < value)) {
			rank[i] += x.Level[i].Span
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	// 随机层数
	level := sl.randomLevel()
	if level > sl.Level {
		for i := sl.Level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.Header
			update[i].Level[i].Span = sl.Len
		}
		sl.Level = level
	}

	// 插入节点并更新forward与backward
	x = createNode(level, score, value)
	for i := 0; i < level; i++ {
		x.Level[i].Forward = update[i].Level[i].Forward // 设置x的i层后置索引
		update[i].Level[i].Forward = x                  // 设置前置节点第i层的后置索引

		x.Level[i].Span = update[i].Level[i].Span - (rank[0] - rank[i])
		update[i].Level[i].Span = (rank[0] - rank[i]) + 1
	}

	for i := level; i < sl.Level; i++ {
		update[i].Level[i].Span++
	}

	// 重排后退指针
	if update[0] != sl.Header {
		x.Backward = update[0]
	}
	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x
	} else {
		sl.Tail = x
	}

	sl.Len++
}

// Delete 删除匹配的元素<score, value>
func (sl *SkipList) Delete(score float64, value string) error {
	update := make([]*SkipNode, MaxLevel)

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && (x.Level[i].Forward.Score < score ||
			(x.Level[i].Forward.Score == score && x.Level[i].Forward.Value < value)) {
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	x = x.Level[0].Forward
	if x != nil && x.Score == score && x.Value == value {
		sl.DeleteNode(x, update)
		return nil
	}

	return ErrNotFound
}

// DeleteNode 删除给定的节点
func (sl *SkipList) DeleteNode(x *SkipNode, update []*SkipNode) {
	for i := 0; i < sl.Level; i++ {
		if update[i].Level[i].Forward == x {
			update[i].Level[i].Span += x.Level[i].Span - 1
			update[i].Level[i].Forward = x.Level[i].Forward
		} else {
			update[i].Level[i].Span -= 1
		}
	}

	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x.Backward
	} else {
		sl.Tail = x.Backward
	}

	for sl.Level > 1 && sl.Header.Level[sl.Level-1].Forward == nil {
		sl.Level--
	}

	sl.Len--
}

// GetRank 返回目标元素在有序集中的 rank
func (sl *SkipList) GetRank(score float64, value string) (uint64, error) {
	rank := uint64(0)
	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && (x.Level[i].Forward.Score < score ||
			(x.Level[i].Forward.Score == score && x.Level[i].Forward.Value <= value)) {
			rank += x.Level[i].Span
			x = x.Level[i].Forward
		}
		if x.Value == value {
			return rank, nil
		}
	}
	return 0, ErrNotFound
}

// GetValueByRank 根据给定的 rank 查找元素
func (sl *SkipList) GetValueByRank(rank uint64) (string, error) {
	x := sl.Header
	traversed := uint64(0)
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && (traversed+x.Level[i].Span) <= rank {
			traversed += x.Level[i].Span
			x = x.Level[i].Forward
		}
		if traversed == rank {
			return x.Value, nil
		}
	}
	return "", ErrNotFound
}

// IsInRange 检查在给定范围内是否存在元素
func (sl *SkipList) IsInRange(r Range) bool {
	if r.Min > r.Max {
		return false
	}

	x := sl.Tail
	if x == nil || !r.GteMin(x.Score) {
		// x == nil || x.score < min
		return false
	}

	x = sl.Header.Level[0].Forward
	if x == nil || !r.LetMax(x.Score) {
		// x == nil || x.score > max
		return false
	}

	return true
}

// FirstInRange 找到跳跃表中第一个符合给定范围的元素
func (sl *SkipList) FirstInRange(r Range) (*SkipNode, error) {
	if !sl.IsInRange(r) {
		return nil, ErrNotFound
	}

	// 找到第一个 Score 值小于给定范围最小值的节点
	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && !r.GteMin(x.Level[i].Forward.Score) {
			x = x.Level[i].Forward
		}
	}

	x = x.Level[0].Forward
	if x == nil || !r.LetMax(x.Score) {
		return nil, ErrNotFound
	}

	return x, nil
}

// LastInRange 找到跳跃表中最后一个符合给定范围的元素
func (sl *SkipList) LastInRange(r Range) (*SkipNode, error) {
	if !sl.IsInRange(r) {
		return nil, ErrNotFound
	}

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && r.LetMax(x.Level[i].Forward.Score) {
			x = x.Level[i].Forward
		}
	}

	if !r.GteMin(x.Score) {
		return nil, ErrNotFound
	}

	return x, nil
}

// DeleteRangeByScore 删除给定范围内的 score 的元素
func (sl *SkipList) DeleteRangeByScore(r Range) uint64 {
	update := make([]*SkipNode, MaxLevel)
	removed := uint64(0)

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && !r.GteMin(x.Level[i].Forward.Score) {
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	// 待删除的第一个节点
	x = x.Level[0].Forward

	for x != nil && r.LetMax(x.Score) {
		// 后继指针
		next := x.Level[0].Forward
		// 删除
		sl.DeleteNode(x, update)
		removed++
		x = next
	}

	return removed
}

// DeleteRangeByRank 删除给定排序范围内的所有元素
func (sl *SkipList) DeleteRangeByRank(start, end uint64) uint64 {
	update := make([]*SkipNode, MaxLevel)
	traversed, removed := uint64(0), uint64(0)

	x := sl.Header
	for i := sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && (traversed+x.Level[i].Span) < start {
			traversed += x.Level[i].Span
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	traversed++
	x = x.Level[0].Forward

	for x != nil && traversed <= end {
		next := x.Level[0].Forward
		sl.DeleteNode(x, update)
		removed++
		traversed++
		x = next
	}

	return removed
}

var rd *rand.Rand

func init() {
	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
}
func (sl *SkipList) randomLevel() int {
	level := 1
	for rd.Intn(100) < 25 { // 默认25%的几率
		level++
	}
	if level > MaxLevel {
		return MaxLevel
	}
	return level
}
