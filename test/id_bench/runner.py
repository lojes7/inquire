import argparse
import multiprocessing
import time
import csv
import os
import json
from node_worker import worker_loop
from verifier import Verifier

def main():
    parser = argparse.ArgumentParser(description='ID Generator Benchmark')
    parser.add_argument('--backend', default='process', choices=['process', 'thread'])
    parser.add_argument('--nodes', type=int, default=8)
    parser.add_argument('--workers-per-node', type=int, default=1)
    parser.add_argument('--duration', type=int, default=30)
    parser.add_argument('--target-qps', type=int, default=200000)
    parser.add_argument('--mode', default='mutex', choices=['mutex', 'batched', 'cachedtime'])
    parser.add_argument('--report-interval', type=int, default=1)
    
    args = parser.parse_args()
    
    print(f"Starting benchmark: {args}")
    
    # Calculate QPS per worker
    total_workers = args.nodes * args.workers_per_node
    qps_per_worker = args.target_qps / total_workers
    
    manager = multiprocessing.Manager()
    result_queue = manager.Queue()
    
    processes = []
    for i in range(args.nodes):
        # For simplicity, 1 worker per node process in this runner, 
        # or we could spawn subprocesses. 
        # The prompt asks for "nodes" and "workers per node".
        # We will spawn (nodes * workers_per_node) processes, assigning them logical node_ids.
        # Logical node_id must be unique for Snowflake.
        
        for w in range(args.workers_per_node):
            logical_node_id = i # If workers share node_id, they must coordinate. 
                                # Snowflake usually requires unique worker_id (machine_id + process_id).
                                # We will assign unique logical_node_id to each process to avoid collision 
                                # unless we implement shared memory lock between workers of same node.
                                # For this test, let's assume logical_node_id = i * workers + w
            
            # However, Snowflake has limited node_id bits (10 bits = 1024).
            # If nodes*workers > 1024, we have a problem.
            # We'll use unique IDs.
            
            unique_node_id = i * args.workers_per_node + w
            
            p = multiprocessing.Process(
                target=worker_loop,
                args=(unique_node_id, args.duration, qps_per_worker, args.mode, result_queue, args.report_interval)
            )
            processes.append(p)
            p.start()
            
    # Collect results
    stats_data = []
    finished_workers = 0
    total_generated = 0
    id_files = []
    
    start_time = time.time()
    
    while finished_workers < total_workers:
        if not result_queue.empty():
            msg = result_queue.get()
            if msg['type'] == 'stats':
                stats_data.append(msg)
                # print(f"Node {msg['node_id']} QPS: {msg['actual_qps']:.2f}")
            elif msg['type'] == 'done':
                finished_workers += 1
                total_generated += msg['count']
                id_files.append(msg['filename'])
                print(f"Worker {msg['node_id']} finished. Generated {msg['count']} IDs.")
        else:
            time.sleep(0.1)
            # Check for timeouts or dead processes
            if time.time() - start_time > args.duration + 10:
                print("Timeout waiting for workers.")
                break
                
    for p in processes:
        p.join()
        
    print("Benchmark finished. Aggregating results...")
    
    # 1. Write Time Series Stats
    out_dir = "d:/mmy/vvechat/test/out"
    os.makedirs(out_dir, exist_ok=True)
    
    ts_file = os.path.join(out_dir, "id_bench_timeseries.csv")
    with open(ts_file, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(["timestamp", "node_id", "actual_qps", "target_qps"])
        for s in stats_data:
            writer.writerow([s['timestamp'], s['node_id'], s['actual_qps'], s['target_qps']])
            
    # 2. Verify IDs (Sample check or full check)
    # Reading all files might be heavy. We'll do a streaming check for duplicates if possible,
    # or just load them if memory permits. For 30s * 200k QPS = 6M IDs. 
    # 6M * 8 bytes = 48MB. Python ints are larger (28 bytes). ~168MB. It fits in memory.
    
    all_ids = []
    for fname in id_files:
        if os.path.exists(fname):
            with open(fname, 'r') as f:
                for line in f:
                    all_ids.append(int(line.strip()))
            # os.remove(fname) # Clean up
            
    print(f"Verifying {len(all_ids)} IDs...")
    verifier = Verifier()
    verification_result = verifier.verify(all_ids)
    
    # 分析性能瓶颈
    bottleneck_analysis = Verifier.analyze_bottlenecks(stats_data, args.duration)
    
    # 3. Write Summary
    summary = {
        "config": vars(args),
        "total_generated": total_generated,
        "verification": verification_result,
        "avg_qps": total_generated / args.duration,
        "bottleneck_analysis": bottleneck_analysis,
        "scalability_notes": {
            "node_scaling": f"测试使用 {args.nodes} 个节点，最大支持 {1023} 个节点 (10位)",
            "concurrency_scaling": f"每节点 {args.workers_per_node} 个工作线程",
            "throughput": f"平均吞吐量 {total_generated / args.duration:.2f} ID/s"
        }
    }
    
    with open(os.path.join(out_dir, "id_bench_results.json"), 'w') as f:
        json.dump(summary, f, indent=2)
        
    print(f"Results written to {out_dir}")
    print(json.dumps(summary, indent=2))

if __name__ == "__main__":
    main()
