// 全局状态
const state = {
    token: localStorage.getItem('vve_token') || null,
    user: JSON.parse(localStorage.getItem('vve_user')) || null,
    currentTab: 'chat', // chat, contacts, me
    activeConversationId: null,
    activeContactId: null,
    friends: [],
    friendRequests: [],
    conversations: [], // 暂时 mock
    
    // 临时状态
    currentContactInfo: null, // 当前查看的用户信息（可能是好友或陌生人）
    isViewingStranger: false
};

// 初始化全局缓存 (容量 20)
// 键: uid (账号), 值: id (用户ID)
const userCache = new LRUCache(20);

// API 基础路径
const API_BASE = 'http://localhost:8080/api';

// DOM 元素引用
const els = {
    loginContainer: document.getElementById('login-container'),
    registerContainer: document.getElementById('register-container'),
    mainContainer: document.getElementById('main-container'),
    
    // 登录/注册 inputs
    loginAccount: document.getElementById('login-account'),
    loginPass: document.getElementById('login-password'),
    btnLogin: document.getElementById('btn-login'),
    linkRegister: document.getElementById('link-register'),
    
    regName: document.getElementById('reg-name'),
    regPhone: document.getElementById('reg-phone'),
    regPass: document.getElementById('reg-password'),
    btnRegister: document.getElementById('btn-register'),
    linkLogin: document.getElementById('link-login'),

    // 主界面
    myAvatar: document.getElementById('my-avatar'),
    navItems: document.querySelectorAll('.nav-item[data-tab]'),
    listContent: document.getElementById('list-content'),
    
    // 搜索
    searchInput: document.getElementById('search-input'),
    searchDropdown: document.getElementById('search-dropdown'),

    // 视图
    chatView: document.getElementById('chat-view'),
    contactView: document.getElementById('contact-view'),
    meView: document.getElementById('me-view'),
    emptyView: document.getElementById('empty-view'),
    newFriendsView: document.getElementById('new-friends-view'),
    
    // 聊天相关
    chatTitle: document.getElementById('chat-title'),
    chatMessages: document.getElementById('chat-messages'),
    msgInput: document.getElementById('msg-input'),
    btnSend: document.getElementById('btn-send'),

    // 新的朋友列表
    newFriendsList: document.getElementById('new-friends-list'),

    // 联系人详情
    contactName: document.getElementById('contact-name'),
    contactInfoContainer: document.getElementById('contact-info-container'),
    btnContactMsg: document.getElementById('btn-contact-msg'),
    btnContactAdd: document.getElementById('btn-contact-add'),
    
    // 联系人操作菜单
    btnContactOptions: document.getElementById('btn-contact-options'),
    contactOptionsMenu: document.getElementById('contact-options-menu'),
    btnEditRemark: document.getElementById('btn-edit-remark'),
    btnDeleteFriend: document.getElementById('btn-delete-friend'),

    // 我 详情
    myName: document.getElementById('my-name'),
    mySignature: document.getElementById('my-signature'),
    myUid: document.getElementById('my-uid'),
    btnOpenSettings: document.getElementById('btn-open-settings'),

    // 模态框
    modalOverlay: document.getElementById('modal-overlay'),
    modalAddFriend: document.getElementById('modal-add-friend'),
    modalSettings: document.getElementById('modal-settings'),
    
    // 加好友
    addFriendMsg: document.getElementById('add-friend-msg'),
    btnCancelAdd: document.getElementById('btn-cancel-add'),
    btnConfirmAdd: document.getElementById('btn-confirm-add'),

    // 设置
    settingUid: document.getElementById('setting-uid'),
    settingName: document.getElementById('setting-name'),
    settingOldPass: document.getElementById('setting-old-pass'),
    settingNewPass: document.getElementById('setting-new-pass'),
    settingConfirmPass: document.getElementById('setting-confirm-pass'),
    btnCancelSetting: document.getElementById('btn-cancel-setting'),
    btnSaveSetting: document.getElementById('btn-save-setting'),

    // 修改备注
    modalRemark: document.getElementById('modal-remark'),
    remarkInput: document.getElementById('remark-input'),
    btnCancelRemark: document.getElementById('btn-cancel-remark'),
    btnSaveRemark: document.getElementById('btn-save-remark'),
};

// --- 初始化 ---
function init() {
    bindEvents();
    if (state.token) {
        showMain();
        loadInitialData();
    } else {
        showLogin();
    }
}

// --- 事件绑定 ---
function bindEvents() {
    // 登录注册切换
    els.linkRegister.addEventListener('click', (e) => {
        e.preventDefault();
        els.loginContainer.classList.add('hidden');
        els.registerContainer.classList.remove('hidden');
    });
    els.linkLogin.addEventListener('click', (e) => {
        e.preventDefault();
        els.registerContainer.classList.add('hidden');
        els.loginContainer.classList.remove('hidden');
    });

    // 登录动作
    els.btnLogin.addEventListener('click', handleLogin);
    
    // 注册动作
    els.btnRegister.addEventListener('click', handleRegister);

    // 导航切换
    els.navItems.forEach(item => {
        item.addEventListener('click', () => {
            const tab = item.dataset.tab;
            switchTab(tab);
        });
    });

    // 发送消息 (Mock)
    els.btnSend.addEventListener('click', handleSendMessage);

    // 搜索相关
    els.searchInput.addEventListener('input', handleSearchInput);
    document.addEventListener('click', (e) => {
        if (!els.searchDropdown.contains(e.target) && e.target !== els.searchInput) {
            els.searchDropdown.classList.add('hidden');
        }
        // 点击其他地方关闭联系人操作菜单
        if (els.contactOptionsMenu && !els.contactOptionsMenu.contains(e.target) && e.target !== els.btnContactOptions) {
            els.contactOptionsMenu.classList.add('hidden');
        }
    });

    // 联系人详情按钮
    els.btnContactMsg.addEventListener('click', () => {
        // 跳转到聊天
        const friend = state.currentContactInfo;
        const conv = {
            id: 'conv_' + friend.id,
            name: friend.name || friend.friend_remark,
            avatar: `https://ui-avatars.com/api/?name=${friend.name}`,
            lastMsg: '',
            time: ''
        };
        // 检查是否已存在
        const exist = state.conversations.find(c => c.id === conv.id);
        if (!exist) {
            state.conversations.unshift(conv);
        }
        switchTab('chat');
        openChat(conv);
    });

    els.btnContactAdd.addEventListener('click', () => {
        openModal('add-friend');
    });
    
    // 联系人操作菜单
    els.btnContactOptions.addEventListener('click', (e) => {
        e.stopPropagation();
        els.contactOptionsMenu.classList.toggle('hidden');
    });
    
    els.btnEditRemark.addEventListener('click', (e) => {
        e.stopPropagation();
        els.contactOptionsMenu.classList.add('hidden');
        // Pre-fill current remark if available
        const currentRemark = state.currentContactInfo ? (state.currentContactInfo.friend_remark || '') : '';
        els.remarkInput.value = currentRemark;
        openModal('remark');
    });
    
    els.btnDeleteFriend.addEventListener('click', handleDeleteFriend);

    // 设置按钮
    els.btnOpenSettings.addEventListener('click', () => {
        els.settingUid.value = state.user.uid;
        els.settingName.value = state.user.name;
        els.settingOldPass.value = '';
        els.settingNewPass.value = '';
        els.settingConfirmPass.value = '';
        openModal('settings');
    });

    // 模态框按钮
    els.btnCancelAdd.addEventListener('click', closeModal);
    els.btnCancelSetting.addEventListener('click', closeModal);
    els.btnCancelRemark.addEventListener('click', closeModal);
    
    els.btnConfirmAdd.addEventListener('click', handleSendFriendRequest);
    els.btnSaveSetting.addEventListener('click', handleSaveSettings);
    els.btnSaveRemark.addEventListener('click', handleSaveRemark);
}

// --- 业务逻辑 ---

// 1. 登录
async function handleLogin() {
    const account = els.loginAccount.value.trim();
    const password = els.loginPass.value.trim();

    if (!account || !password) {
        alert('请输入账号和密码');
        return;
    }

    const isPhone = /^\d{11}$/.test(account);
    const endpoint = isPhone ? '/login/phone_number' : '/login/uid';
    const body = isPhone 
        ? { phone_number: account, password } 
        : { uid: account, password };

    try {
        const res = await apiCall(endpoint, 'POST', body);
        if (res && res.token_class) {
            state.token = res.token_class.token;
            state.user = res.user_info;
            
            localStorage.setItem('vve_token', state.token);
            localStorage.setItem('vve_user', JSON.stringify(state.user));

            showMain();
            loadInitialData();
        } else {
            alert('登录失败，请检查账号密码');
        }
    } catch (err) {
        console.error(err);
        alert('登录请求出错: ' + err.message);
    }
}

// 2. 注册
async function handleRegister() {
    const name = els.regName.value.trim();
    const phone = els.regPhone.value.trim();
    const password = els.regPass.value.trim();

    if (!name || !phone || !password) {
        alert('请填写完整信息');
        return;
    }

    try {
        await apiCall('/register', 'POST', { name, phone_number: phone, password });
        alert('注册成功，请登录');
        els.linkLogin.click();
    } catch (err) {
        alert('注册失败: ' + err.message);
    }
}

// 3. 切换 Tab
function switchTab(tabName) {
    state.currentTab = tabName;
    
    els.navItems.forEach(item => {
        if (item.dataset.tab === tabName) {
            item.classList.add('active');
        } else {
            item.classList.remove('active');
        }
    });

    renderSidebarList();

    if (tabName === 'me') {
        renderMeView();
        showView('me');
    } else if (tabName === 'contacts') {
        showView('empty'); 
    } else {
        if (state.activeConversationId) {
            showView('chat');
        } else {
            showView('empty');
        }
    }
}

// 4. 加载初始数据
async function loadInitialData() {
    els.myAvatar.src = `https://ui-avatars.com/api/?name=${state.user.name}&background=random`;

    try {
        const [friends, requests] = await Promise.all([
            apiCall('/auth/friendships', 'GET'),
            apiCall('/auth/friendship_requests', 'GET')
        ]);
        
        state.friends = friends || [];
        state.friendRequests = requests || [];
        
        // Mock 会话
        if (state.conversations.length === 0) {
            state.conversations = [
                {
                    id: 'sys_1',
                    name: '欢迎信息',
                    avatar: 'https://ui-avatars.com/api/?name=WeChat',
                    lastMsg: '欢迎使用',
                    time: '12:00'
                }
            ];
        }

        renderSidebarList();

    } catch (err) {
        console.error('加载初始数据失败', err);
    }
}

// 5. 渲染左侧列表
function renderSidebarList() {
    els.listContent.innerHTML = '';

    if (state.currentTab === 'chat') {
        state.conversations.forEach(conv => {
            const el = createListItem(conv.id, conv.name, conv.lastMsg, conv.time, conv.avatar);
            el.onclick = () => openChat(conv);
            if (state.activeConversationId === conv.id) el.classList.add('active');
            els.listContent.appendChild(el);
        });
    } else if (state.currentTab === 'contacts') {
        // 1. 新的朋友 (固定入口)
        const newFriendItem = document.createElement('div');
        newFriendItem.className = 'list-item';
        newFriendItem.innerHTML = `
            <div class="item-avatar" style="background:#fa9d3b; display:flex; align-items:center; justify-content:center; color:white;">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>
            </div>
            <div class="item-info">
                <div class="item-top"><span class="item-name">新的朋友</span></div>
            </div>
        `;
        newFriendItem.onclick = () => openNewFriends();
        els.listContent.appendChild(newFriendItem);

        // 2. 联系人标题
        const friendTitle = document.createElement('div');
        friendTitle.className = 'group-title';
        friendTitle.innerText = '联系人';
        els.listContent.appendChild(friendTitle);

        // 3. 好友列表
        state.friends.forEach(friend => {
            const name = friend.friend_remark || friend.name || '好友 ' + friend.friend_id;
            const el = createListItem(friend.friend_id, name, '', '', `https://ui-avatars.com/api/?name=${name}`);
            el.onclick = () => openContact(friend.friend_id, false); // false = not stranger
            if (state.activeContactId === friend.friend_id) el.classList.add('active');
            els.listContent.appendChild(el);
        });
    }
}

function createListItem(id, title, subtitle, time, avatarUrl) {
    const div = document.createElement('div');
    div.className = 'list-item';
    div.dataset.id = id;
    div.innerHTML = `
        <img src="${avatarUrl}" class="item-avatar">
        <div class="item-info">
            <div class="item-top">
                <span class="item-name">${title}</span>
                <span class="item-time">${time}</span>
            </div>
            <div class="item-msg">${subtitle}</div>
        </div>
    `;
    return div;
}

// 6. 打开聊天
function openChat(conv) {
    state.activeConversationId = conv.id;
    renderSidebarList();
    
    els.chatTitle.innerText = conv.name;
    els.chatMessages.innerHTML = `
        <div class="message-item received">
            <img src="${conv.avatar}" class="avatar" style="cursor:pointer" onclick="handleAvatarClick('${conv.id}')">
            <div class="bubble">${conv.lastMsg || '开始聊天吧'}</div>
        </div>
    `;
    showView('chat');
}

// 处理头像点击
window.handleAvatarClick = function(convId) {
    console.log('Avatar clicked:', convId);
    // 简单处理：如果是好友会话，尝试跳转到好友详情
    // convId 格式: conv_{friend_id}
    if (convId && convId.startsWith('conv_')) {
        const friendId = convId.split('_')[1];
        // 检查是否是好友 (注意类型转换，friend_id 可能是字符串)
        const isFriend = state.friends.some(f => String(f.friend_id) === String(friendId));
        
        if (isFriend) {
            switchTab('contacts');
            openContact(friendId, false);
        } else {
            console.log('Not a friend or friend not found in list');
        }
    }
};

// 7. 打开联系人 (好友或陌生人)
async function openContact(id, isStranger) {
    state.activeContactId = id;
    state.isViewingStranger = isStranger;
    
    if (isStranger) {
        document.querySelectorAll('.list-item').forEach(el => el.classList.remove('active'));
    } else {
        renderSidebarList();
    }

    try {
        let info;
        if (isStranger) {
            info = await apiCall(`/auth/info/strangers/id/${id}`, 'GET');
        } else {
            info = await apiCall(`/auth/info/friends/id/${id}`, 'GET');
        }
        
        // 更新缓存 (如果返回信息中有 uid)
        if (info && info.uid && info.id) {
            userCache.put(info.uid, info.id);
        }
        
        state.currentContactInfo = info;
        renderContactView(info, isStranger);
        
    } catch (err) {
        console.error(err);
        alert('获取用户信息失败');
    }
}

function renderContactView(info, isStranger) {
    els.contactName.innerText = info.friend_remark || info.name;
    
    // 动态渲染信息行
    els.contactInfoContainer.innerHTML = '';
    
    // 只有当字段存在时才显示
    if (info.uid) {
        els.contactInfoContainer.appendChild(createInfoRow('账号', info.uid));
    }
    if (info.region) {
        els.contactInfoContainer.appendChild(createInfoRow('地区', info.region));
    }
    if (info.signature) {
        // 签名通常显示在名字下面，但这里我们放在信息栏或者忽略
        // 如果要显示在名字下面，需要修改 HTML 结构，这里简单起见放在信息栏
        els.contactInfoContainer.appendChild(createInfoRow('个性签名', info.signature));
    }

    if (isStranger) {
        els.btnContactMsg.classList.add('hidden');
        els.btnContactAdd.classList.remove('hidden');
        els.btnContactOptions.classList.add('hidden'); // 陌生人没有更多操作
    } else {
        els.btnContactMsg.classList.remove('hidden');
        els.btnContactAdd.classList.add('hidden');
        els.btnContactOptions.classList.remove('hidden'); // 好友显示更多操作
    }
    
    showView('contact');
}

function createInfoRow(label, value) {
    const div = document.createElement('div');
    div.className = 'info-row';
    div.innerHTML = `<span class="label">${label}</span><span class="value">${value}</span>`;
    return div;
}

// 8. 打开新的朋友列表
function openNewFriends() {
    state.activeContactId = null;
    renderSidebarList(); // 清除高亮
    
    els.newFriendsList.innerHTML = '';
    
    if (state.friendRequests.length === 0) {
        els.newFriendsList.innerHTML = '<div style="text-align:center; color:#999; margin-top:50px;">暂无好友申请</div>';
    } else {
        state.friendRequests.forEach(req => {
            const item = document.createElement('div');
            item.className = 'new-friend-item';
            
            let actionHtml = '';
            if (req.status === 'accepted') {
                actionHtml = '<span class="new-friend-status">已添加</span>';
            } else {
                actionHtml = `
                    <button class="btn-accept" onclick="handleAcceptRequest('${req.request_id}')">接受</button>
                    <button class="btn-accept" onclick="handleRejectRequest('${req.request_id}')">拒绝</button>
                `;
            }

            item.innerHTML = `
                <img src="https://ui-avatars.com/api/?name=${req.sender_name}" class="new-friend-avatar">
                <div class="new-friend-info">
                    <div class="new-friend-name">${req.sender_name}</div>
                    <div class="new-friend-msg">${req.verification_message}</div>
                </div>
                ${actionHtml}
            `;
            els.newFriendsList.appendChild(item);
        });
    }
    
    showView('new-friends');
}

// 9. 接受好友请求
window.handleAcceptRequest = async function(reqId) {
    try {
        // 后端接口修改：POST /:request_id 即为接受，无需 body
        await apiCall(`/auth/friendship_requests/${reqId}`, 'POST');
        
        // 更新本地状态
        const req = state.friendRequests.find(r => r.request_id == reqId);
        if (req) req.status = 'accepted';
        
        // 重新加载好友列表
        const friends = await apiCall('/auth/friendships', 'GET');
        state.friends = friends || [];
        
        // 刷新界面
        openNewFriends();
        
    } catch (err) {
        alert('操作失败: ' + err.message);
    }
};

// 9.1 拒绝/删除好友请求
window.handleRejectRequest = async function(reqId) {
    try {
        await apiCall(`/auth/friendship_requests/${reqId}`, 'DELETE');
        // 从本地状态移除该申请
        state.friendRequests = state.friendRequests.filter(r => String(r.request_id) !== String(reqId));
        openNewFriends();
    } catch (err) {
        alert('操作失败: ' + err.message);
    }
};

// 10. 删除好友
async function handleDeleteFriend() {
    els.contactOptionsMenu.classList.add('hidden');
    if (!confirm('确定要删除该好友吗？')) return;
    
    // 修复 Bug: 使用 state.activeContactId 作为 friend_id
    // state.activeContactId 是从侧边栏点击时传入的 friend.friend_id，这是最可靠的来源
    const friendId = state.activeContactId;
    
    if (!friendId) {
        alert('无法获取好友ID，请刷新重试');
        return;
    }
    
    try {
        await apiCall(`/auth/friendships/${friendId}`, 'DELETE');
        alert('删除成功');
        
        // 从缓存中移除 (如果存在)
        // 注意：我们现在缓存的是 uid -> id，无法直接通过 id 移除
        // 除非我们遍历缓存或者维护反向映射。
        // 鉴于"不用保证与数据库强一致"的要求，这里可以忽略，或者仅当我们知道 uid 时移除
        if (state.currentContactInfo && state.currentContactInfo.uid) {
            userCache.remove(state.currentContactInfo.uid);
        }

        // 更新本地列表
        state.friends = state.friends.filter(f => f.friend_id != friendId);
        state.activeContactId = null;
        
        renderSidebarList();
        showView('empty');
        
    } catch (err) {
        alert('删除失败: ' + err.message);
    }
}

// 11. 渲染 "我"
function renderMeView() {
    if (!state.user) return;
    els.myName.innerText = state.user.name;
    els.myUid.innerText = state.user.uid;
    // 移除手机号显示
    // 移除签名显示 (如果后端没返回)
    if (state.user.signature) {
        els.mySignature.innerText = state.user.signature;
    } else {
        els.mySignature.innerText = '';
    }
}

// 11. 发送消息 (Mock)
function handleSendMessage() {
    const content = els.msgInput.value.trim();
    if (!content) return;

    const msgDiv = document.createElement('div');
    msgDiv.className = 'message-item sent';
    msgDiv.innerHTML = `
        <img src="https://ui-avatars.com/api/?name=${state.user.name}&background=random" class="avatar">
        <div class="bubble">${escapeHtml(content)}</div>
    `;
    els.chatMessages.appendChild(msgDiv);
    els.msgInput.value = '';
    els.chatMessages.scrollTop = els.chatMessages.scrollHeight;
}

// 12. 搜索逻辑
function handleSearchInput(e) {
    const val = e.target.value.trim();
    if (!val) {
        els.searchDropdown.classList.add('hidden');
        return;
    }

    els.searchDropdown.innerHTML = '';
    els.searchDropdown.classList.remove('hidden');

    // 1. 搜索好友 (本地过滤)
    const matchedFriends = state.friends.filter(f => {
        const name = f.friend_remark || '';
        return name.includes(val);
    });

    if (matchedFriends.length > 0) {
        const header = document.createElement('div');
        header.style.padding = '5px 10px';
        header.style.color = '#999';
        header.style.fontSize = '12px';
        header.innerText = '好友';
        els.searchDropdown.appendChild(header);

        matchedFriends.forEach(f => {
            const item = document.createElement('div');
            item.className = 'search-item';
            item.innerHTML = `
                <div class="search-item-icon">友</div>
                <span>${f.friend_remark}</span>
            `;
            item.onclick = () => {
                openContact(f.friend_id, false);
                els.searchDropdown.classList.add('hidden');
                els.searchInput.value = '';
            };
            els.searchDropdown.appendChild(item);
        });
    }

    // 2. 搜索网络 (陌生人)
    const searchNetItem = document.createElement('div');
    searchNetItem.className = 'search-item';
    searchNetItem.innerHTML = `
        <div class="search-item-icon" style="background:#576b95;">搜</div>
        <span>搜索账号：${escapeHtml(val)}</span>
    `;
    searchNetItem.onclick = () => searchStranger(val);
    els.searchDropdown.appendChild(searchNetItem);
}

async function searchStranger(uid) {
    els.searchDropdown.classList.add('hidden');
    els.searchInput.value = '';
    
    try {
        let info;
        // 1. 查缓存 (LRU: uid -> id)
        const cachedId = userCache.get(uid);
        
        if (cachedId) {
            console.log('LRU Cache Hit for UID:', uid, '-> ID:', cachedId);
            // 缓存命中：直接使用 ID 获取最新信息 (避免使用可能过期的缓存信息)
            // 满足需求：通过账号以 O(1) 获得 id，然后获取实时数据
            info = await apiCall(`/auth/info/strangers/id/${cachedId}`, 'GET');
        } else {
            console.log('LRU Cache Miss for UID:', uid);
            // 缓存未命中：使用 UID 搜索
            info = await apiCall(`/auth/info/strangers/uid/${uid}`, 'GET');
        }

        // 2. 更新缓存 (uid -> id)
        if (info && info.id) {
            // 注意：如果 info 中没有 uid，我们使用搜索时的 uid
            // 后端 StrangerInfoResp 可能不返回 uid，但我们知道它就是当前的 uid
            userCache.put(uid, info.id);
        }

        // 成功找到
        state.currentContactInfo = info;
        renderContactView(info, true); // true = stranger
    } catch (err) {
        console.error(err);
        alert('未找到该用户');
    }
}

// 13. 发送好友请求
async function handleSendFriendRequest() {
    const msg = els.addFriendMsg.value.trim();
    const targetId = state.currentContactInfo.id; 

    try {
        await apiCall('/auth/friendship_requests', 'POST', {
            receiver_id: String(targetId), 
            sender_name: state.user.name,
            verification_message: msg
        });
        alert('好友申请已发送');
        closeModal();
    } catch (err) {
        alert('发送失败: ' + err.message);
    }
}

// 14. 修改设置 (账号、昵称、密码)
async function handleSaveSettings() {
    const newUid = els.settingUid.value.trim();
    const newName = els.settingName.value.trim();
    const oldPass = els.settingOldPass.value.trim();
    const confirmPass = els.settingConfirmPass.value.trim();
    const newPass = els.settingNewPass.value.trim();
    
    let hasChange = false;
    let msg = [];

    // 1. 修改 UID
    if (newUid && newUid !== state.user.uid) {
        try {
            await apiCall('/auth/me/uid', 'POST', { new_uid: newUid });
            state.user.uid = newUid;
            msg.push('UID修改成功');
            hasChange = true;
        } catch (err) {
            msg.push('UID修改失败: ' + err.message);
        }
    }

    // 2. 修改昵称
    if (newName && newName !== state.user.name) {
        try {
            await apiCall('/auth/me/name', 'POST', { new_name: newName });
            state.user.name = newName;
            msg.push('昵称修改成功');
            hasChange = true;
        } catch (err) {
            msg.push('昵称修改失败: ' + err.message);
        }
    }

    // 3. 修改密码
    if (oldPass || newPass || confirmPass) {
        if (!oldPass || !newPass || !confirmPass) {
            alert('修改密码请填写：旧密码、新密码和确认密码');
            return;
        }
        if (newPass !== confirmPass) {
            alert('两次输入的新密码不一致');
            return;
        }
        if (newPass.length < 6 || newPass.length > 72) {
             alert('新密码长度必须在6-72位之间');
             return;
        }
        try {
            await apiCall('/auth/me/password', 'POST', { prev_password: oldPass, new_password: newPass });
            msg.push('密码修改成功');
        } catch (err) {
            msg.push('密码修改失败: ' + err.message);
        }
    }

    if (msg.length > 0) {
        alert(msg.join('\n'));
    }

    if (hasChange) {
        localStorage.setItem('vve_user', JSON.stringify(state.user));
        renderMeView();
    }
    
    if (msg.length === 0 && !hasChange) {
        // Did nothing
    } else {
        closeModal();
    }
}

// 15. 修改好友备注
async function handleSaveRemark() {
    // 使用 activeContactId 作为 friend_id：它来自好友列表点击时传入的 friend.friend_id，是最可靠的来源。
    // currentContactInfo 往往是“用户信息对象”，不一定包含 friend_id 字段，导致这里拿不到 ID 而直接 return。
    const friendId = state.activeContactId;
    const newRemark = els.remarkInput.value.trim();
    
    if (!friendId) {
        alert('无法获取好友ID：请先从联系人列表打开一个好友，再修改备注。');
        return;
    }

    try {
        await apiCall(`/auth/friendships/remark/${friendId}`, 'POST', { remark: newRemark });
        
        // 更新本地数据
        const friend = state.friends.find(f => String(f.friend_id) === String(friendId));
        if (friend) {
            friend.friend_remark = newRemark;
        }
        
        // 更新当前显示信息
        if (state.currentContactInfo && !state.isViewingStranger) {
            state.currentContactInfo.friend_remark = newRemark;
            renderContactView(state.currentContactInfo, false);
        }
        
        // 更新侧边栏
        renderSidebarList();
        
        alert('备注修改成功');
        closeModal();
    } catch (err) {
        alert('修改失败: ' + err.message);
    }
}

// --- 工具函数 ---

function showLogin() {
    els.loginContainer.classList.remove('hidden');
    els.mainContainer.classList.add('hidden');
}

function showMain() {
    els.loginContainer.classList.add('hidden');
    els.registerContainer.classList.add('hidden');
    els.mainContainer.classList.remove('hidden');
}

function showView(viewName) {
    els.chatView.classList.add('hidden');
    els.contactView.classList.add('hidden');
    els.meView.classList.add('hidden');
    els.emptyView.classList.add('hidden');
    els.newFriendsView.classList.add('hidden');

    if (viewName === 'chat') els.chatView.classList.remove('hidden');
    else if (viewName === 'contact') els.contactView.classList.remove('hidden');
    else if (viewName === 'me') els.meView.classList.remove('hidden');
    else if (viewName === 'new-friends') els.newFriendsView.classList.remove('hidden');
    else els.emptyView.classList.remove('hidden');
}

function openModal(type) {
    els.modalOverlay.classList.remove('hidden');
    els.modalAddFriend.classList.add('hidden');
    els.modalSettings.classList.add('hidden');
    els.modalRemark.classList.add('hidden');

    if (type === 'add-friend') els.modalAddFriend.classList.remove('hidden');
    if (type === 'settings') els.modalSettings.classList.remove('hidden');
    if (type === 'remark') els.modalRemark.classList.remove('hidden');
}

function closeModal() {
    els.modalOverlay.classList.add('hidden');
}

async function apiCall(path, method, body) {
    const headers = {
        'Content-Type': 'application/json'
    };
    if (state.token) {
        headers['Authorization'] = 'Bearer ' + state.token;
    }

    const opts = {
        method,
        headers,
    };
    if (body) {
        opts.body = JSON.stringify(body);
    }

    const res = await fetch(API_BASE + path, opts);
    
    if (res.status === 401) {
        logout();
        throw new Error('Unauthorized');
    }

    const json = await res.json();
    if (!res.ok) {
        throw new Error(json.message || json.error || 'API Error');
    }
    return json.data;
}

function logout() {
    state.token = null;
    state.user = null;
    localStorage.removeItem('vve_token');
    localStorage.removeItem('vve_user');
    location.reload();
}

function escapeHtml(text) {
    if (!text) return '';
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}

// 启动
init()
