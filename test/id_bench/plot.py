import csv
import os
import sys
from pathlib import Path

try:
    import matplotlib.pyplot as plt
except ModuleNotFoundError:
    plt = None

def plot_results(out_dir):
    if plt is None:
        print("缺少依赖：matplotlib。无法生成图表。")
        print("解决：在当前 Python 环境里安装 matplotlib，例如：")
        print("  python -m pip install matplotlib")
        return

    ts_file = os.path.join(out_dir, "id_bench_timeseries.csv")
    if not os.path.exists(ts_file):
        print(f"No data file found at {ts_file}")
        return

    timestamps = []
    qps_values = []
    targets = []

    # Aggregate QPS per second across all nodes
    # We need to bucket by timestamp (integer second)
    buckets = {}

    with open(ts_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            ts = float(row['timestamp'])
            sec = int(ts)
            qps = float(row['actual_qps'])
            target = float(row['target_qps'])
            
            if sec not in buckets:
                buckets[sec] = {'actual': 0, 'target': 0, 'count': 0}
            buckets[sec]['actual'] += qps
            buckets[sec]['target'] += target # Sum of targets per node = total target
            buckets[sec]['count'] += 1

    sorted_secs = sorted(buckets.keys())
    if not sorted_secs:
        print("No data to plot.")
        return

    # Filter out first and last second as they might be partial
    if len(sorted_secs) > 2:
        sorted_secs = sorted_secs[1:-1]

    x = []
    y_actual = []
    y_target = []

    for sec in sorted_secs:
        x.append(sec - sorted_secs[0])
        y_actual.append(buckets[sec]['actual'])
        # Target is already summed up if we assume each report contains the node's target share
        # But wait, if we have N nodes, and each reports M times, we sum them up.
        # Actually, we should sum the QPS of all nodes for that second.
        # The target in the CSV is "per node". So sum of targets is correct.
        y_target.append(buckets[sec]['target'])

    plt.figure(figsize=(10, 6))
    plt.plot(x, y_actual, label='Actual QPS')
    plt.plot(x, y_target, label='Target QPS', linestyle='--')
    plt.xlabel('Time (s)')
    plt.ylabel('QPS')
    plt.title('ID Generator Stability')
    plt.legend()
    plt.grid(True)
    
    out_png = os.path.join(out_dir, "id_bench_stability.png")
    plt.savefig(out_png)
    print(f"Plot saved to {out_png}")

if __name__ == "__main__":
    out_dir = str((Path(__file__).resolve().parents[1] / 'out'))
    if len(sys.argv) > 1:
        out_dir = sys.argv[1]
    plot_results(out_dir)
