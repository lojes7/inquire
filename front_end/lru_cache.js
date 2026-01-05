/**
 * LRU Cache Implementation
 * 
 * 需求：
 * 1. 在内存中存储
 * 2. 使用哈希表 + 双向链表实现
 * 3. 键值映射：微信号 (uid) -> 用户 ID (id)
 * 4. 复杂度：O(1)
 * 5. 仅存储 ID，不存储用户信息，避免数据不一致
 */

/**
 * 双向链表节点
 */
class Node {
    constructor(key, value) {
        this.key = key;   // 微信号 (uid)
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
 * LRU 缓存类
 */
class LRUCache {
    constructor(capacity) {
        this.capacity = Number.isFinite(capacity) ? capacity : 0;
        // 哈希表: uid -> Node
        // 使用 Map 避免对象原型链键冲突，更符合“哈希表”语义
        this.map = new Map();
        this.list = new DoublyLinkedList(); // 双向链表
        this.size = 0;
    }

    /**
     * 获取用户 ID
     * @param {string} key - 微信号 (uid)
     * @returns {string|number|null} - 用户 ID
     */
    get(key) {
        if (this.map.has(key)) {
            const node = this.map.get(key);
            this.list.moveToHead(node); // 访问后移动到头部，标记为最近使用
            return node.value;
        }
        return null;
    }

    /**
     * 写入缓存
     * @param {string} key - 微信号 (uid)
     * @param {string|number} value - 用户 ID
     */
    put(key, value) {
        if (this.capacity <= 0) return;

        if (this.map.has(key)) {
            // 如果已存在，更新值并移动到头部
            const node = this.map.get(key);
            node.value = value;
            this.list.moveToHead(node);
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
        if (this.map.has(key)) {
            const node = this.map.get(key);
            this.list.removeNode(node);
            this.map.delete(key);
            this.size--;
        }
    }
}
