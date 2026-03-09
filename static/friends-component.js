// Vue.js Friends компонент

const FriendsComponent = {
    name: 'Friends',
    setup() {
        const { ref, reactive, onMounted } = Vue;

        const friends = ref([]);
        const requests = ref([]);
        const isLoading = ref(false);
        const showRequests = ref(false);
        const searchQuery = ref('');
        const showAddModal = ref(false);
        const addFriendEmail = ref('');

        onMounted(() => {
            loadFriends();
            loadRequests();
        });

        async function loadFriends() {
            isLoading.value = true;
            friends.value = await SocialStore.loadFriends();
            isLoading.value = false;
        }

        async function loadRequests() {
            requests.value = await SocialStore.getFriendRequests();
        }

        async function acceptRequest(requestId) {
            const result = await SocialStore.acceptFriendRequest(requestId);
            if (result.success) {
                await loadFriends();
                await loadRequests();
            } else {
                alert('❌ ' + result.error);
            }
        }

        async function rejectRequest(requestId) {
            await SocialStore.rejectFriendRequest(requestId);
            await loadRequests();
        }

        async function removeFriend(friendId) {
            if (!confirm('Удалить этого друга?')) return;
            await SocialStore.removeFriend(friendId);
            await loadFriends();
        }

        async function addFriend() {
            if (!addFriendEmail.value) return;

            // В реальном приложении здесь был бы поиск по email
            alert('Функция поиска по email в разработке!');
            showAddModal.value = false;
        }

        function getFilteredFriends() {
            if (!searchQuery.value) return friends.value;
            const query = searchQuery.value.toLowerCase();
            return friends.value.filter(f =>
                f.name.toLowerCase().includes(query)
            );
        }

        return {
            friends,
            requests,
            isLoading,
            showRequests,
            searchQuery,
            showAddModal,
            addFriendEmail,
            loadFriends,
            loadRequests,
            acceptRequest,
            rejectRequest,
            removeFriend,
            addFriend,
            getFilteredFriends
        };
    },
    template: `
        <div class="friends-page">
            <div class="friends-header">
                <h2>👥 Друзья</h2>
                <div class="friends-actions">
                    <button class="btn-primary" @click="showAddModal = true">
                        ➕ Добавить
                    </button>
                    <button class="btn-secondary" @click="showRequests = !showRequests">
                        📬 Запросы ({{ requests.length }})
                    </button>
                </div>
            </div>

            <!-- Запросы в друзья -->
            <transition name="slide">
                <div v-if="showRequests" class="requests-panel">
                    <h3>Входящие запросы</h3>
                    <div v-if="!requests.length" class="empty">Нет запросов</div>
                    <div v-for="req in requests" :key="req.id" class="request-item">
                        <div class="request-info">
                            <strong>{{ req.sender_name }}</strong>
                            <span class="request-date">{{ new Date(req.created_at).toLocaleDateString() }}</span>
                        </div>
                        <div class="request-actions">
                            <button class="btn-accept" @click="acceptRequest(req.id)">✅</button>
                            <button class="btn-reject" @click="rejectRequest(req.id)">❌</button>
                        </div>
                    </div>
                </div>
            </transition>

            <!-- Поиск -->
            <div class="search-box">
                <input v-model="searchQuery" type="text" placeholder="🔍 Поиск друзей..." class="search-input" />
            </div>

            <!-- Список друзей -->
            <div v-if="isLoading" class="loading">
                <div class="spinner"></div>
                <p>Загрузка друзей...</p>
            </div>

            <div v-else class="friends-list">
                <div v-if="!getFilteredFriends().length" class="empty">
                    Нет друзей. Добавь первым!
                </div>

                <div v-for="friend in getFilteredFriends()" :key="friend.id" class="friend-card">
                    <div class="friend-header">
                        <div class="friend-avatar">{{ friend.name.charAt(0) }}</div>
                        <div class="friend-info">
                            <div class="friend-name">{{ friend.name }}</div>
                            <div class="friend-stats">
                                <span>Ур.{{ friend.level }}</span>
                                <span>⭐ {{ friend.rating }}</span>
                            </div>
                        </div>
                        <div class="friend-status" :class="{ online: friend.is_online }">
                            {{ friend.is_online ? '🟢 Онлайн' : '⚫ Офлайн' }}
                        </div>
                    </div>
                    <div class="friend-actions">
                        <button class="btn-message" title="Написать">💬</button>
                        <button class="btn-challenge" title="Вызвать">⚔️</button>
                        <button class="btn-profile" title="Профиль">👤</button>
                        <button class="btn-remove" @click="removeFriend(friend.id)" title="Удалить">🗑️</button>
                    </div>
                </div>
            </div>

            <!-- Modal: Добавить друга -->
            <transition name="fade">
                <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
                    <div class="modal">
                        <div class="modal-header">
                            <h3>➕ Добавить друга</h3>
                            <button class="modal-close" @click="showAddModal = false">×</button>
                        </div>
                        <div class="modal-body">
                            <div class="form-group">
                                <label class="form-label">Email или имя</label>
                                <input v-model="addFriendEmail" type="text" class="form-input" placeholder="friend@example.com" />
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button class="btn-secondary" @click="showAddModal = false">Отмена</button>
                            <button class="btn-primary" @click="addFriend">Отправить запрос</button>
                        </div>
                    </div>
                </div>
            </transition>
        </div>
    `
};

window.FriendsComponent = FriendsComponent;
