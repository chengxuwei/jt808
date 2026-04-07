package pkg

import (
	"sync"
	"sync/atomic"
)

// terminalSeqNo 单终端下行序号管理；注释原因：每个终端独立维护，互不干扰；注释人：Cursor
type terminalSeqNo struct {
	counter atomic.Uint32
}

// next 返回下一个序号（1~65535 循环，跳过 0）；注释原因：JT808 协议流水号范围 0x0001~0xFFFF；注释人：Cursor
func (t *terminalSeqNo) next() uint16 {
	for {
		old := t.counter.Load()
		// 1~65535 循环：old % 0xFFFF ∈ [0,65534]，+1 ∈ [1,65535]；注释人：Cursor
		next := (old%0xFFFF + 1)
		if t.counter.CompareAndSwap(old, next) {
			return uint16(next)
		}
	}
}

// seqNoStore 全局每终端序号存储；注释原因：sync.Map 保证并发多连接安全；注释人：Cursor
var seqNoStore sync.Map // key: terminalNo(string) → *terminalSeqNo

// NextSeqNo 获取指定终端的下一个下行消息序号（线程安全）；
// 注释原因：平台侧下行每个终端独立递增，与终端上行序号互相独立；注释人：Cursor
func NextSeqNo(terminalNo string) uint16 {
	v, _ := seqNoStore.LoadOrStore(terminalNo, &terminalSeqNo{})
	return v.(*terminalSeqNo).next()
}
