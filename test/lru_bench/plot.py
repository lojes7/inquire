import csv
import matplotlib.pyplot as plt
import os
import sys


def _parse_optional_float(value):
    if value is None:
        return None
    s = str(value).strip().lower()
    if s in ("", "none", "null"):
        return None
    return float(value)


def _default_out_dir():
    here = os.path.dirname(os.path.abspath(__file__))
    repo_root = os.path.abspath(os.path.join(here, "..", ".."))
    return os.path.join(repo_root, "test", "out")

def plot_lru(out_dir):
    csv_file = os.path.join(out_dir, "lru_bench_results.csv")
    if not os.path.exists(csv_file):
        print(f"No LRU bench results found at: {os.path.abspath(csv_file)}")
        print("Hint: run `python test/lru_bench/run_lru_bench.py` first, or pass the correct out_dir as an argument.")
        return

    # series_points: name -> list[(capacity, hit_rate, hit_rate_std, throughput, throughput_std)]
    series_points = {}

    with open(csv_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            workload = row.get('workload', '').strip()
            capacity = int(row['capacity'])
            hit_rate = float(row['hit_rate'])
            hit_rate_std = _parse_optional_float(row.get('hit_rate_std'))
            throughput = _parse_optional_float(row.get('throughput_ops_per_sec'))
            throughput_std = _parse_optional_float(row.get('throughput_ops_per_sec_std'))

            if workload == 'uniform':
                series_name = 'uniform'
            elif workload == 'zipf':
                s = _parse_optional_float(row.get('zipf_s'))
                series_name = f'zipf_s={s}' if s is not None else 'zipf'
            else:
                continue

            series_points.setdefault(series_name, []).append((capacity, hit_rate, hit_rate_std, throughput, throughput_std))

    # Sort each series by capacity
    for name in list(series_points.keys()):
        series_points[name].sort(key=lambda t: t[0])

    # ---- Hit rate plot ----
    plt.figure(figsize=(10, 6))
    for name, points in series_points.items():
        xs = [p[0] for p in points]
        ys = [p[1] for p in points]
        yerr = [p[2] for p in points]
        use_err = all(v is not None for v in yerr) and any(v > 0 for v in yerr)
        label = 'Uniform' if name == 'uniform' else name.replace('zipf_s=', 'Zipf (s=') + (')' if name.startswith('zipf_s=') else '')
        if use_err:
            plt.errorbar(xs, ys, yerr=yerr, label=label, marker='o', capsize=3)
        else:
            plt.plot(xs, ys, label=label, marker='o')

    plt.xlabel('Cache Capacity')
    plt.ylabel('Hit Rate')
    plt.title('LRU Cache Hit Rate vs Capacity')
    plt.legend()
    plt.grid(True)
    plt.xscale('log')

    out_png = os.path.join(out_dir, 'lru_bench_hitrate.png')
    plt.savefig(out_png)
    print(f"Plot saved to {out_png}")

    # ---- Throughput plot ----
    plt.figure(figsize=(10, 6))
    plotted_any = False
    for name, points in series_points.items():
        xs = [p[0] for p in points]
        ys = [p[3] for p in points]
        yerr = [p[4] for p in points]
        if any(v is None for v in ys):
            continue
        use_err = all(v is not None for v in yerr) and any(v > 0 for v in yerr)
        label = 'Uniform' if name == 'uniform' else name.replace('zipf_s=', 'Zipf (s=') + (')' if name.startswith('zipf_s=') else '')
        if use_err:
            plt.errorbar(xs, ys, yerr=yerr, label=label, marker='o', capsize=3)
        else:
            plt.plot(xs, ys, label=label, marker='o')
        plotted_any = True

    if plotted_any:
        plt.xlabel('Cache Capacity')
        plt.ylabel('Throughput (ops/sec)')
        plt.title('LRU Cache Throughput vs Capacity')
        plt.legend()
        plt.grid(True)
        plt.xscale('log')
        out_png = os.path.join(out_dir, 'lru_bench_throughput.png')
        plt.savefig(out_png)
        print(f"Plot saved to {out_png}")

if __name__ == "__main__":
    out_dir = _default_out_dir()
    if len(sys.argv) > 1:
        out_dir = sys.argv[1]
    plot_lru(out_dir)
