import subprocess
import csv
import json
import os

def main():
    out_dir = "d:/mmy/vvechat/test/out"
    os.makedirs(out_dir, exist_ok=True)
    
    capacities = [100, 500, 1000, 5000, 10000]
    workloads = ['uniform', 'zipf']
    keyspace = 10000
    requests = 100000
    zipf_s_values = [0.8, 1.0, 1.2, 1.5]  # 不同的 Zipf 参数
    
    results = []
    
    print("=" * 60)
    print("Running LRU Benchmarks...")
    print("=" * 60)
    
    for w in workloads:
        s_list = zipf_s_values if w == 'zipf' else [1.1]
        
        for s in s_list:
            for c in capacities:
                print(f"Running {w} (s={s}) with capacity {c}...")
                cmd = [
                    'node', 
                    'd:/mmy/vvechat/test/lru_bench/benchmark_lru.js',
                    str(c),
                    w,
                    str(keyspace),
                    str(requests),
                    str(s),
                    '42'   # seed
                ]
                
                try:
                    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
                    data = json.loads(result.stdout)
                    results.append(data)
                    
                    # 打印关键指标
                    print(f"  -> Hit Rate: {data['hit_rate']:.2%}, Throughput: {data.get('throughput_ops_per_sec', 0):.0f} ops/s")
                    
                except subprocess.CalledProcessError as e:
                    print(f"Error running benchmark: {e.stderr}")
                except json.JSONDecodeError:
                    print(f"Invalid JSON output: {result.stdout}")

    # Write CSV
    csv_file = os.path.join(out_dir, "lru_bench_results.csv")
    if results:
        with open(csv_file, 'w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=results[0].keys())
            writer.writeheader()
            writer.writerows(results)
        
    # Write JSON for detailed analysis
    json_file = os.path.join(out_dir, "lru_bench_results.json")
    with open(json_file, 'w') as f:
        json.dump(results, f, indent=2)

    # Print Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    
    # Group by workload
    for w in workloads:
        print(f"\n{w.upper()} Workload:")
        w_results = [r for r in results if r['workload'] == w]
        for r in w_results:
            s_info = f" (s={r.get('zipf_s', '')})" if w == 'zipf' else ""
            print(f"  Capacity {r['capacity']:>5}{s_info}: Hit Rate = {r['hit_rate']:.2%}")
        
    print(f"\nResults written to {csv_file}")

if __name__ == "__main__":
    main()
