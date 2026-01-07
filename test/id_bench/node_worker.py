import os
import time

from id_generator import SnowflakeGenerator

def worker_loop(node_id, duration, target_qps, mode, result_queue, out_dir, report_interval=1):
    """
    Worker process that generates IDs at a fixed QPS.
    """
    generator = SnowflakeGenerator(node_id, mode=mode)
    
    start_time = time.time()
    end_time = start_time + duration
    
    generated_ids = []
    
    # Rate limiting state
    interval_per_req = 1.0 / target_qps if target_qps > 0 else 0
    count = 0
    
    # Reporting state
    last_report_time = start_time
    ids_since_last_report = 0
    
    # Metrics
    latencies = [] # Sampled
    
    try:
        while time.time() < end_time:
            now = time.time()
            
            # Fixed QPS Logic:
            # Expected time for next request = start_time + (count * interval)
            # If now < expected, sleep.
            
            expected_time = start_time + (count * interval_per_req)
            if expected_time > now:
                sleep_time = expected_time - now
                # Only sleep if it's worth it (e.g. > 1ms), otherwise spin or just proceed (busy wait is more precise)
                if sleep_time > 0.001:
                    time.sleep(sleep_time)
                else:
                    # Busy wait for high precision
                    while time.time() < expected_time:
                        pass
            
            # Generate ID
            t0 = time.perf_counter()
            try:
                new_id = generator.next_id()
                generated_ids.append(new_id)
                ids_since_last_report += 1
                count += 1
            except Exception as e:
                # Clock moved backwards or other error
                print(f"Worker {node_id} error: {e}")
                time.sleep(0.01)
                continue
            
            t1 = time.perf_counter()
            if count % 100 == 0: # Sample latency
                latencies.append(t1 - t0)
            
            # Report
            if now - last_report_time >= report_interval:
                actual_qps = ids_since_last_report / (now - last_report_time)
                # Send stats to main process (lightweight)
                result_queue.put({
                    "type": "stats",
                    "node_id": node_id,
                    "timestamp": now,
                    "actual_qps": actual_qps,
                    "target_qps": target_qps
                })
                last_report_time = now
                ids_since_last_report = 0
                
    except KeyboardInterrupt:
        pass
        
    # Send all IDs for verification (in chunks if needed, but for this test we send all)
    # Warning: Large lists might block the queue. In real prod, write to file.
    # Here we write to a temp file and send the filename.
    
    os.makedirs(out_dir, exist_ok=True)
    filename = os.path.join(out_dir, f"ids_{node_id}.txt")
    with open(filename, "w", encoding="utf-8") as f:
        for id_val in generated_ids:
            f.write(f"{id_val}\n")
            
    result_queue.put({
        "type": "done",
        "node_id": node_id,
        "count": len(generated_ids),
        "filename": filename,
        "avg_latency": sum(latencies)/len(latencies) if latencies else 0
    })
