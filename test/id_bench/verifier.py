class Verifier:
    def __init__(self):
        # Bit layout for decoding
        self.node_id_bits = 10
        self.sequence_bits = 12
        self.timestamp_bits = 41
        
        self.node_id_shift = self.sequence_bits
        self.timestamp_shift = self.sequence_bits + self.node_id_bits
        
        self.node_mask = (1 << self.node_id_bits) - 1
        self.sequence_mask = (1 << self.sequence_bits) - 1
        self.timestamp_mask = (1 << self.timestamp_bits) - 1

    def decode_id(self, id_val):
        """解码 Snowflake ID 为各组成部分"""
        sequence = id_val & self.sequence_mask
        node_id = (id_val >> self.node_id_shift) & self.node_mask
        timestamp = (id_val >> self.timestamp_shift) & self.timestamp_mask
        return {"timestamp": timestamp, "node_id": node_id, "sequence": sequence}

    def verify(self, ids_list):
        """
        验证 ID 列表：
        1. 全局唯一性（无重复）
        2. 同节点内有序性
        3. 统计各节点生成数量
        """
        if not ids_list:
            return {"valid": True, "count": 0, "errors": []}

        errors = []
        
        # 1. 检查唯一性
        unique_ids = set(ids_list)
        duplicate_count = len(ids_list) - len(unique_ids)
        if duplicate_count > 0:
            errors.append(f"发现 {duplicate_count} 个重复ID。总数: {len(ids_list)}, 唯一: {len(unique_ids)}")

        # 2. 按节点分组并检查有序性
        node_ids_map = {}  # node_id -> [(timestamp, sequence, original_id), ...]
        
        for id_val in ids_list:
            decoded = self.decode_id(id_val)
            node_id = decoded['node_id']
            if node_id not in node_ids_map:
                node_ids_map[node_id] = []
            node_ids_map[node_id].append((decoded['timestamp'], decoded['sequence'], id_val))

        # 检查每个节点内的有序性
        node_stats = {}
        monotonicity_errors = 0
        
        for node_id, entries in node_ids_map.items():
            # 按生成顺序（假设输入列表保持生成顺序）检查
            # 由于多进程写入文件的顺序可能与生成顺序一致，我们直接检查
            prev_ts = -1
            prev_seq = -1
            order_violations = 0
            
            for ts, seq, _ in entries:
                if ts < prev_ts:
                    order_violations += 1
                elif ts == prev_ts and seq <= prev_seq:
                    order_violations += 1
                prev_ts = ts
                prev_seq = seq
            
            node_stats[node_id] = {
                "count": len(entries),
                "order_violations": order_violations,
                "min_timestamp": min(e[0] for e in entries) if entries else 0,
                "max_timestamp": max(e[0] for e in entries) if entries else 0
            }
            
            monotonicity_errors += order_violations

        if monotonicity_errors > 0:
            errors.append(f"发现 {monotonicity_errors} 个有序性违规")

        return {
            "valid": len(errors) == 0,
            "count": len(ids_list),
            "unique_count": len(unique_ids),
            "duplicate_count": duplicate_count,
            "node_count": len(node_ids_map),
            "node_stats": node_stats,
            "errors": errors
        }

    @staticmethod
    def analyze_bottlenecks(stats_data, duration):
        """
        分析性能瓶颈来源
        
        可能的瓶颈:
        1. 锁竞争 - QPS 随并发增加而下降
        2. 系统时钟调用开销 - 高频率下延迟增加
        3. 检查模块汇总成本 - 验证时间占比过高
        """
        analysis = {
            "lock_contention": False,
            "clock_overhead": False,
            "aggregation_cost": False,
            "details": []
        }
        
        if not stats_data:
            return analysis
        
        # 计算平均 QPS 和标准差
        qps_values = [s.get('actual_qps', 0) for s in stats_data]
        target_qps_values = [s.get('target_qps', 0) for s in stats_data]
        
        if qps_values:
            avg_qps = sum(qps_values) / len(qps_values)
            avg_target = sum(target_qps_values) / len(target_qps_values) if target_qps_values else 0
            
            # 如果实际 QPS 远低于目标，可能存在锁竞争
            if avg_target > 0 and avg_qps < avg_target * 0.7:
                analysis["lock_contention"] = True
                analysis["details"].append(
                    f"锁竞争警告: 实际 QPS ({avg_qps:.2f}) 远低于目标 ({avg_target:.2f})"
                )
            
            # 检查 QPS 波动 - 高标准差可能表示锁竞争
            if len(qps_values) > 1:
                variance = sum((x - avg_qps) ** 2 for x in qps_values) / len(qps_values)
                std_dev = variance ** 0.5
                cv = std_dev / avg_qps if avg_qps > 0 else 0  # 变异系数
                
                if cv > 0.3:  # 变异系数超过 30%
                    analysis["lock_contention"] = True
                    analysis["details"].append(
                        f"QPS 波动大 (CV={cv:.2%})，可能存在锁竞争"
                    )
        
        return analysis
