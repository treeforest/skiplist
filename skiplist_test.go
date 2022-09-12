package skiplist

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSkipList(t *testing.T) {
	sl := New()

	printInfo := func() {
		fmt.Println("SkipList: ")
		for i := 1; i <= sl.Level; i++ {
			fmt.Printf("%2d ", i)
		}
		fmt.Print("\n")
		x := sl.Header.Level[0].Forward
		for x != nil {
			for i := 0; i < sl.Level; i++ {
				if i < len(x.Level) {
					fmt.Printf("%2s ", "*")
				} else {
					fmt.Printf("%2s ", " ")
				}
			}
			fmt.Printf(" [score:%3.1f value:%s] ", x.Score, x.Value)
			fmt.Print("\n")
			x = x.Level[0].Forward
		}
	}

	sl.Insert(3, "hello")
	sl.Insert(3, "world")
	sl.Insert(10, "golang")
	sl.Insert(9, "c++")
	sl.Insert(9, "rust")
	sl.Insert(9, "java")
	printInfo()
	/*
		例如：SkipList:
		 1  2  3  4
		 *  *        [score:3.0 value:hello]
		 *  *  *  *  [score:3.0 value:world]
		 *  *        [score:9.0 value:c++]
		 *           [score:9.0 value:java]
		 *  *        [score:9.0 value:rust]
		 *           [score:10.0 value:golang]
	*/

	require.Equal(t, sl.IsInRange(Range{Min: 3, Max: 10}), true)
	require.Equal(t, sl.IsInRange(Range{Min: 0, Max: 10}), true)
	require.Equal(t, sl.IsInRange(Range{Min: 0, Max: 20}), true)
	require.Equal(t, sl.IsInRange(Range{Min: 11, Max: 20}), false)
	require.Equal(t, sl.IsInRange(Range{Min: 0, Max: 2}), false)

	x, err := sl.FirstInRange(Range{Min: 4, Max: 10})
	require.NoError(t, err)
	require.Equal(t, "c++", x.Value)

	x, err = sl.LastInRange(Range{Min: 4, Max: 20})
	require.NoError(t, err)
	require.Equal(t, "golang", x.Value)

	rank, err := sl.GetRank(9, "java")
	require.NoError(t, err)
	require.Equal(t, uint64(4), rank)

	err = sl.Delete(9, "c++")
	require.NoError(t, err)

	rank, err = sl.GetRank(9, "java")
	require.NoError(t, err)
	require.Equal(t, uint64(3), rank)

	printInfo()

	value, err := sl.GetValueByRank(3)
	require.NoError(t, err)
	require.Equal(t, "java", value)

	removed := sl.DeleteRangeByRank(3, 4)
	require.Equal(t, uint64(2), removed)

	printInfo()

	removed = sl.DeleteRangeByScore(Range{Min: 3, Max: 8})
	require.Equal(t, uint64(2), removed)

	printInfo()
}

func BenchmarkSkipList_Insert(b *testing.B) {
	sl := New()
	rand.Seed(time.Now().UnixNano())
	p := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		rand.Read(p)
		score := float64(rand.Intn(10000))
		value := hex.EncodeToString(p)
		sl.Insert(score, value)
	}
}
