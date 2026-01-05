/**
 * LRU Cache Implementation
 * 
 * 需求：
 * 1. 在内存中存储
 * 2. 使用哈希表 + 双向链表实现
 * 3. 键值映射：用户标识 (uid) -> 用户 ID (id)
 * 4. 复杂度：O(1)
 * 5. 仅存储 ID，不存储用户信息，避免数据不一致
 */

/**
 * 双向链表节点
 */
class Node {
    constructor(key, value) {
        this.key = key;   // 用户标识 (uid)
        this.value = value; // 用户 ID (id)
        this.prev = null;
        this.next = null;
    }
}

/**
 * 双向链表 (手写实现)
 */
class DoublyLinkedList {
    constructor() {
        this.head = new Node(null, null); // 虚拟头节点
        this.tail = new Node(null, null); // 虚拟尾节点
        this.head.next = this.tail;
        this.tail.prev = this.head;
    }

    // 添加节点到头部
    addToHead(node) {
        node.prev = this.head;
        node.next = this.head.next;
        this.head.next.prev = node;
        this.head.next = node;
    }

    // 移除节点
    removeNode(node) {
        node.prev.next = node.next;
        node.next.prev = node.prev;
    }

    // 移动节点到头部
    moveToHead(node) {
        this.removeNode(node);
        this.addToHead(node);
    }

    // 移除尾部节点 (最久未使用)
    removeTail() {
        if (this.tail.prev === this.head) return null;
        const node = this.tail.prev;
        this.removeNode(node);
        return node;
    }
}

/**
 * 手写哈希表（数组桶 + 链地址法）
 * - key 统一转为 string
 * - 手写哈希函数：FNV-1a 32-bit
 */
class HashTableEntry {
    constructor(key, value, next = null) {
        this.key = key;
        this.value = value;
        this.next = next;
    }
}

class HashTable {
    constructor(bucketCount = 53) {
        this.bucketCount = bucketCount;
        this.buckets = new Array(bucketCount).fill(null);
        this.size = 0;
    }

    _toKey(key) {
        return String(key);
    }

    // FNV-1a 32-bit
    _hash(keyStr) {
        let hash = 0x811c9dc5;
        for (let i = 0; i < keyStr.length; i++) {
            hash ^= keyStr.charCodeAt(i);
            hash = (hash * 0x01000193) >>> 0;
        }
        return hash;
    }

    _index(keyStr) {
        return this._hash(keyStr) % this.bucketCount;
    }

    has(key) {
        return this.get(key) !== undefined;
    }

    get(key) {
        const keyStr = this._toKey(key);
        const idx = this._index(keyStr);
        let cur = this.buckets[idx];
        while (cur) {
            if (cur.key === keyStr) return cur.value;
            cur = cur.next;
        }
        return undefined;
    }

    set(key, value) {
        const keyStr = this._toKey(key);
        const idx = this._index(keyStr);
        let cur = this.buckets[idx];
        while (cur) {
            if (cur.key === keyStr) {
                cur.value = value;
                return;
            }
            cur = cur.next;
        }

        const entry = new HashTableEntry(keyStr, value, this.buckets[idx]);
        this.buckets[idx] = entry;
        this.size++;

        if (this.size / this.bucketCount > 0.75) {
            this._rehash(this._nextBucketCount(this.bucketCount));
        }
    }

    delete(key) {
        const keyStr = this._toKey(key);
        const idx = this._index(keyStr);
        let cur = this.buckets[idx];
        let prev = null;
        while (cur) {
            if (cur.key === keyStr) {
                if (prev) prev.next = cur.next;
                else this.buckets[idx] = cur.next;
                this.size--;
                return true;
            }
            prev = cur;
            cur = cur.next;
        }
        return false;
    }

    _rehash(newBucketCount) {
        const oldBuckets = this.buckets;
        this.bucketCount = newBucketCount;
        this.buckets = new Array(newBucketCount).fill(null);
        this.size = 0;

        for (const head of oldBuckets) {
            let cur = head;
            while (cur) {
                this.set(cur.key, cur.value);
                cur = cur.next;
            }
        }
    }

    _nextBucketCount(current) {
        // 简单扩容策略：翻倍后再 +1，保证是奇数，减少某些模式碰撞
        return current * 2 + 1;
    }
}

/**
 * LRU 缓存类
 */
class LRUCache {
    constructor(capacity) {
        this.capacity = Number.isFinite(capacity) ? capacity : 0;
        // 手写哈希表: uid -> Node
        this.map = new HashTable(53);
        this.list = new DoublyLinkedList(); // 双向链表
        this.size = 0;
        // 因容量限制触发的淘汰次数（仅统计 eviction，不统计 remove）
        this._evictionCount = 0;
    }

    // 只读属性：允许 cache.evictionCount 读取，但禁止外部写入
    get evictionCount() {
        return this._evictionCount;
    }

    // 只读方法：对外提供稳定接口，便于以后扩展统计逻辑
    getEvictionCount() {
        return this._evictionCount;
    }

    /**
     * 获取用户 ID
     * @param {string} key - 用户标识 (uid)
     * @returns {string|number|null} - 用户 ID
     */
    get(key) {
        const node = this.map.get(key);
        if (node !== undefined) {
            this.list.moveToHead(node); // 访问后移动到头部，标记为最近使用
            return node.value;
        }
        return null;
    }

    /**
     * 写入缓存
     * @param {string} key - 用户标识 (uid)
     * @param {string|number} value - 用户 ID
     */
    put(key, value) {
        if (this.capacity <= 0) return;

        const existNode = this.map.get(key);
        if (existNode !== undefined) {
            // 如果已存在，更新值并移动到头部
            existNode.value = value;
            this.list.moveToHead(existNode);
        } else {
            // 如果不存在，创建新节点
            const newNode = new Node(key, value);
            this.map.set(key, newNode);
            this.list.addToHead(newNode);
            this.size++;

            // 如果超过容量，移除尾部节点
            if (this.size > this.capacity) {
                const tail = this.list.removeTail();
                if (tail) {
                    this.map.delete(tail.key);
                    this.size--;
                    this._evictionCount++;
                }
            }
        }
    }

    // 移除 getUserId 方法，因为 get 直接返回 ID
    
    /**
     * 移除缓存
     * @param {string} key 
     */
    remove(key) {
        const node = this.map.get(key);
        if (node !== undefined) {
            this.list.removeNode(node);
            this.map.delete(key);
            this.size--;
        }
    }
}

module.exports = { LRUCache };

module.exports = { LRUCache };
