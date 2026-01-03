package utils

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

const (
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	epoch int64 = 1288834974657

	// NodeBits holds the number of bits to use for Node
	nodeBits uint8 = 10

	// StepBits holds the number of bits to use for Step
	stepBits uint8 = 12
)

// Snowflake struct holds the basic information needed for a snowflake generator
type Snowflake struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64 // 上一个 ID 的时间
	node  int64 // 节点号
	step  int64 // 序列号

	nodeMax   int64 // 节点 ID 最大值
	stepMask  int64 // 序列号最大值，作掩码使用
	timeShift uint8 // 时间偏移长度
	nodeShift uint8 // 节点偏移长度
}

var sf *Snowflake

// InitSnowflake returns a new snowflake node that can be used to generate snowflake ID
func InitSnowflake(node int64) error {
	if nodeBits+stepBits > 22 {
		return errors.New("节点号和序列号超出22位")
	}

	s := Snowflake{}
	s.node = node
	s.nodeMax = (1 << nodeBits) - 1
	s.stepMask = (1 << stepBits) - 1
	s.nodeShift = stepBits
	s.timeShift = nodeBits + stepBits

	if s.node < 0 || s.node > s.nodeMax {
		return errors.New("Node number must be between 0 and " + strconv.FormatInt(s.nodeMax, 10))
	}

	var curTime = time.Now()
	// 通过 time.Now() 用time包的特性确保同一进程使用单调时钟
	s.epoch = curTime.Add(time.Unix(epoch/1000, (epoch%1000)*1000000).Sub(curTime))

	sf = &s
	return nil
}

// NewUniqueID return a unique snowflake ID
func NewUniqueID() uint64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Since(sf.epoch).Milliseconds()

	if now == sf.time {
		sf.step = (sf.step + 1) & sf.stepMask

		if sf.step == 0 {
			for now <= sf.time {
				now = time.Since(sf.epoch).Milliseconds()
			}
		}
	} else if now < sf.time {
		for now <= sf.time {
			now = time.Since(sf.epoch).Milliseconds()
		}
		sf.step = 0
	} else {
		sf.step = 0
	}

	sf.time = now

	r := uint64((now)<<sf.timeShift | (sf.node << sf.nodeShift) | (sf.step))

	return r
}
