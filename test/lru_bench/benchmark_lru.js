const { LRUCache } = require('../../front_end/lru_cache.js');
const fs = require('fs');
const path = require('path');

// Parse args
const args = process.argv.slice(2);
const capacity = parseInt(args[0]) || 1000;
const workload = args[1] || 'uniform'; // uniform | zipf
const keyspace = parseInt(args[2]) || 10000;
const requests = parseInt(args[3]) || 20000;
const zipf_s = parseFloat(args[4]) || 1.1;
const seed = parseInt(args[5]) || 42;

// Simple PRNG (Mulberry32)
function mulberry32(a) {
    return function() {
      var t = a += 0x6D2B79F5;
      t = Math.imul(t ^ t >>> 15, t | 1);
      t ^= t + Math.imul(t ^ t >>> 7, t | 61);
      return ((t ^ t >>> 14) >>> 0) / 4294967296;
    }
}

const rand = mulberry32(seed);

/**
 * Zipf 分布生成器
 * 使用预计算 CDF + 二分查找实现
 * 
 * 参数:
 *   s: Zipf 指数 (通常 1.0-2.0, 越大热点越集中)
 *   N: 键空间大小
 */
class ZipfGenerator {
    constructor(s, N, randFunc) {
        this.s = s;
        this.N = N;
        this.rand = randFunc;
        
        // 预计算 CDF
        // p(k) = c / k^s, 其中 c = 1 / H_{N,s}
        // H_{N,s} = sum(1/k^s for k=1 to N)
        
        let harmonic = 0;
        for (let k = 1; k <= N; k++) {
            harmonic += 1.0 / Math.pow(k, s);
        }
        this.c = 1.0 / harmonic;
        
        // 预计算累积分布函数 (CDF)
        this.cdf = new Array(N + 1);
        this.cdf[0] = 0;
        let cumulative = 0;
        for (let k = 1; k <= N; k++) {
            cumulative += this.c / Math.pow(k, s);
            this.cdf[k] = cumulative;
        }
    }
    
    /**
     * 生成一个符合 Zipf 分布的随机数 (0 到 N-1)
     * 使用二分查找
     */
    next() {
        const u = this.rand();
        
        // 二分查找 CDF
        let lo = 1, hi = this.N;
        while (lo < hi) {
            const mid = (lo + hi) >> 1;
            if (this.cdf[mid] < u) {
                lo = mid + 1;
            } else {
                hi = mid;
            }
        }
        return lo - 1; // 返回 0-indexed
    }
}

function run() {
    const cache = new LRUCache(capacity);
    let hits = 0;
    let misses = 0;
    
    // 初始化键生成器
    let keyGen;
    if (workload === 'zipf') {
        // 真正的 Zipf 分布
        const zipfGen = new ZipfGenerator(zipf_s, keyspace, rand);
        keyGen = () => zipfGen.next();
    } else {
        // 均匀随机分布
        keyGen = () => Math.floor(rand() * keyspace);
    }
    
    const start = process.hrtime.bigint();
    
    for (let i = 0; i < requests; i++) {
        const key = keyGen();
        const val = cache.get(key);
        if (val !== -1 && val !== undefined && val !== null) {
            hits++;
        } else {
            misses++;
            cache.put(key, key); // Store key as value
        }
    }
    
    const end = process.hrtime.bigint();
    const duration_ns = Number(end - start);
    const avg_latency_ns = duration_ns / requests;
    
    const result = {
        capacity,
        workload,
        keyspace,
        requests,
        zipf_s: workload === 'zipf' ? zipf_s : null,
        seed,
        hits,
        misses,
        hit_rate: hits / requests,
        eviction_count: 0,
        avg_latency_ns,
        throughput_ops_per_sec: requests / (duration_ns / 1e9)
    };
    
    // Try to read eviction count if available
    if (cache.evictionCount !== undefined) {
        result.eviction_count = cache.evictionCount;
    }

    console.log(JSON.stringify(result));
}

run();
