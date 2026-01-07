import time
import threading

class SnowflakeGenerator:
    def __init__(self, node_id, mode='mutex'):
        self.node_id = node_id
        self.mode = mode
        
        # Bit layout:
        # 1 bit sign | 41 bits timestamp | 10 bits node_id | 12 bits sequence
        self.epoch = 1704067200000  # 2024-01-01 00:00:00 UTC
        self.node_id_bits = 10
        self.sequence_bits = 12
        
        self.max_node_id = -1 ^ (-1 << self.node_id_bits)
        self.max_sequence = -1 ^ (-1 << self.sequence_bits)
        
        self.node_id_shift = self.sequence_bits
        self.timestamp_shift = self.sequence_bits + self.node_id_bits
        
        if self.node_id > self.max_node_id:
            raise ValueError(f"Node ID must be between 0 and {self.max_node_id}")
            
        self.last_timestamp = -1
        self.sequence = 0
        
        self.lock = threading.Lock()
        
        # For batched mode
        self.batch_size = 100
        self.current_batch_end = 0
        
        # For cached time mode
        self.cached_time = -1
        self.cached_time_update_interval = 0.001 # 1ms
        self.last_cache_update = 0

    def _current_timestamp(self):
        return int(time.time() * 1000)

    def _wait_for_next_millis(self, last_timestamp):
        timestamp = self._current_timestamp()
        while timestamp <= last_timestamp:
            timestamp = self._current_timestamp()
        return timestamp

    def next_id(self):
        if self.mode == 'mutex':
            return self._next_id_mutex()
        elif self.mode == 'batched':
            return self._next_id_batched()
        elif self.mode == 'cachedtime':
            return self._next_id_cached_time()
        else:
            return self._next_id_mutex()

    def _next_id_mutex(self):
        with self.lock:
            timestamp = self._current_timestamp()

            if timestamp < self.last_timestamp:
                # Clock moved backwards. Reject or wait.
                # For this test, we'll wait/reject.
                # Simple strategy: wait until it catches up if diff is small, else raise
                diff = self.last_timestamp - timestamp
                if diff < 5:
                    timestamp = self._wait_for_next_millis(self.last_timestamp)
                else:
                    raise Exception(f"Clock moved backwards. Refusing to generate id for {diff} milliseconds")

            if self.last_timestamp == timestamp:
                self.sequence = (self.sequence + 1) & self.max_sequence
                if self.sequence == 0:
                    timestamp = self._wait_for_next_millis(self.last_timestamp)
            else:
                self.sequence = 0

            self.last_timestamp = timestamp

            return ((timestamp - self.epoch) << self.timestamp_shift) | \
                   (self.node_id << self.node_id_shift) | \
                   self.sequence

    def _next_id_batched(self):
        # Batching：一次锁内分配一段 sequence 区间 [start, end)
        # self.sequence 作为“下一个要发放的 sequence”，self.current_batch_end 为区间右开边界。
        with self.lock:
            timestamp = self._current_timestamp()

            # 时钟回拨：直接等待追平
            if timestamp < self.last_timestamp:
                timestamp = self._wait_for_next_millis(self.last_timestamp)

            # 还能在同一毫秒内从当前 batch 继续发放
            if timestamp == self.last_timestamp and self.sequence < self.current_batch_end:
                seq = self.sequence
                self.sequence += 1
                return self._compose_id(timestamp, seq)

            # 需要分配新 batch
            if timestamp != self.last_timestamp:
                start_seq = 0
            else:
                start_seq = self.current_batch_end

            # sequence 空间耗尽：等到下一毫秒再从 0 开始
            if start_seq > self.max_sequence:
                timestamp = self._wait_for_next_millis(self.last_timestamp)
                start_seq = 0

            end_seq = min(start_seq + self.batch_size, self.max_sequence + 1)  # 右开
            self.last_timestamp = timestamp
            self.current_batch_end = end_seq
            self.sequence = start_seq + 1  # 下一个
            return self._compose_id(timestamp, start_seq)

    def _next_id_cached_time(self):
        # Reduces system calls to time.time()
        # Not strictly monotonic across processes if they don't share memory, 
        # but within one process it reduces overhead.
        # Note: This is risky for strict ordering but good for throughput.
        with self.lock:
            now = time.time()
            if now - self.last_cache_update > self.cached_time_update_interval:
                self.cached_time = int(now * 1000)
                self.last_cache_update = now
            
            timestamp = self.cached_time
            
            # Logic similar to mutex but using cached timestamp
            # This might generate IDs with same timestamp for longer than 1ms real time
            # But we must ensure sequence doesn't overflow for that "virtual" ms
            
            if timestamp < self.last_timestamp:
                 # This happens if cached time hasn't updated but we called next_id
                 # We must treat it as same timestamp
                 timestamp = self.last_timestamp

            if self.last_timestamp == timestamp:
                self.sequence = (self.sequence + 1) & self.max_sequence
                if self.sequence == 0:
                    # Force update time
                    timestamp = self._wait_for_next_millis(self.last_timestamp)
                    self.cached_time = timestamp
                    self.last_cache_update = time.time()
            else:
                self.sequence = 0

            self.last_timestamp = timestamp
            return self._compose_id(timestamp, self.sequence)

    def _compose_id(self, timestamp, sequence):
        return ((timestamp - self.epoch) << self.timestamp_shift) | \
               (self.node_id << self.node_id_shift) | \
               sequence
