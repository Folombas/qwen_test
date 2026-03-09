// Vue.js Activity Feed компонент

const ActivityFeedComponent = {
    name: 'ActivityFeed',
    setup() {
        const { ref, onMounted } = Vue;

        const activities = ref([]);
        const isLoading = ref(false);

        onMounted(async () => {
            await loadActivity();
        });

        async function loadActivity() {
            isLoading.value = true;
            activities.value = await SocialStore.getActivityFeed();
            isLoading.value = false;
        }

        function getActionIcon(action) {
            const icons = {
                'level_up': '🎉',
                'achievement': '🏆',
                'quiz_complete': '✅',
                'challenge_win': '⚔️',
                'friend_add': '👥',
                'default': '📝'
            };
            return icons[action] || icons['default'];
        }

        function getActionColor(action) {
            const colors = {
                'level_up': 'level-up',
                'achievement': 'achievement',
                'quiz_complete': 'quiz',
                'challenge_win': 'challenge',
                'default': 'default'
            };
            return colors[action] || colors['default'];
        }

        function formatTime(dateStr) {
            const date = new Date(dateStr);
            const now = new Date();
            const diff = now - date;

            const minutes = Math.floor(diff / 60000);
            const hours = Math.floor(diff / 3600000);
            const days = Math.floor(diff / 86400000);

            if (minutes < 1) return 'Только что';
            if (minutes < 60) return `${minutes} мин. назад`;
            if (hours < 24) return `${hours} ч. назад`;
            if (days < 7) return `${days} дн. назад`;

            return date.toLocaleDateString('ru-RU');
        }

        return {
            activities,
            isLoading,
            loadActivity,
            getActionIcon,
            getActionColor,
            formatTime
        };
    },
    template: `
        <div class="activity-page">
            <div class="activity-header">
                <h2>📜 Лента активности</h2>
                <button class="btn-refresh" @click="loadActivity()">🔄</button>
            </div>

            <div v-if="isLoading" class="loading">
                <div class="spinner"></div>
                <p>Загрузка...</p>
            </div>

            <div v-else class="activity-list">
                <div v-if="!activities.length" class="empty">
                    <div class="empty-icon">📭</div>
                    <p>Пока нет активности</p>
                    <p class="empty-hint">Добавь друзей чтобы видеть их достижения!</p>
                </div>

                <div v-for="act in activities" :key="act.id"
                     class="activity-item"
                     :class="getActionColor(act.action)">
                    <div class="activity-icon">
                        {{ getActionIcon(act.action) }}
                    </div>
                    <div class="activity-content">
                        <div class="activity-user">{{ act.user_name }}</div>
                        <div class="activity-description">{{ act.description }}</div>
                        <div class="activity-meta">
                            <span class="activity-time">{{ formatTime(act.created_at) }}</span>
                            <span v-if="act.score > 0" class="activity-score">+{{ act.score }} ⭐</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Challenges компонент (Дуэли)
const ChallengesComponent = {
    name: 'Challenges',
    setup() {
        const { ref, onMounted } = Vue;

        const challenges = ref([]);
        const friends = ref([]);
        const showSendModal = ref(false);
        const selectedFriend = ref(null);

        onMounted(async () => {
            await loadChallenges();
            await loadFriends();
        });

        async function loadChallenges() {
            challenges.value = await SocialStore.getChallenges('pending');
        }

        async function loadFriends() {
            friends.value = await SocialStore.loadFriends();
        }

        async function sendChallenge() {
            if (!selectedFriend.value) return;

            const result = await SocialStore.sendChallenge(
                selectedFriend.value.id,
                selectedFriend.value.name
            );

            if (result.success) {
                alert('✅ Вызов отправлен!');
                showSendModal.value = false;
                await loadChallenges();
            } else {
                alert('❌ ' + result.error);
            }
        }

        function acceptChallenge(challengeId) {
            // В реальной реализации нужен API endpoint
            alert('Функция принятия вызова в разработке!');
        }

        function rejectChallenge(challengeId) {
            alert('Функция отклонения вызова в разработке!');
        }

        return {
            challenges,
            friends,
            showSendModal,
            selectedFriend,
            loadChallenges,
            loadFriends,
            sendChallenge,
            acceptChallenge,
            rejectChallenge
        };
    },
    template: `
        <div class="challenges-page">
            <div class="challenges-header">
                <h2>⚔️ Дуэли</h2>
                <button class="btn-primary" @click="showSendModal = true">
                    📤 Отправить вызов
                </button>
            </div>

            <div class="challenges-list">
                <div v-if="!challenges.length" class="empty">
                    <div class="empty-icon">⚔️</div>
                    <p>Нет активных вызовов</p>
                    <button class="btn-primary" @click="showSendModal = true">
                        Создать вызов
                    </button>
                </div>

                <div v-for="ch in challenges" :key="ch.id" class="challenge-card">
                    <div class="challenge-header">
                        <div class="challenge-vs">
                            <span class="challenge-player">{{ ch.sender_name }}</span>
                            <span class="vs-text">VS</span>
                            <span class="challenge-player">{{ ch.receiver_name }}</span>
                        </div>
                        <span class="challenge-status">{{ ch.status }}</span>
                    </div>
                    <div class="challenge-info">
                        <span>📅 {{ new Date(ch.created_at).toLocaleDateString() }}</span>
                    </div>
                    <div class="challenge-actions">
                        <button class="btn-accept" @click="acceptChallenge(ch.id)">✅ Принять</button>
                        <button class="btn-reject" @click="rejectChallenge(ch.id)">❌ Отклонить</button>
                    </div>
                </div>
            </div>

            <!-- Modal: Отправить вызов -->
            <transition name="fade">
                <div v-if="showSendModal" class="modal-overlay" @click.self="showSendModal = false">
                    <div class="modal">
                        <div class="modal-header">
                            <h3>⚔️ Отправить вызов</h3>
                            <button class="modal-close" @click="showSendModal = false">×</button>
                        </div>
                        <div class="modal-body">
                            <div class="form-group">
                                <label class="form-label">Выберите друга</label>
                                <select v-model="selectedFriend" class="form-input">
                                    <option v-for="friend in friends" :key="friend.id" :value="friend">
                                        {{ friend.name }} (Ур.{{ friend.level }})
                                    </option>
                                </select>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button class="btn-secondary" @click="showSendModal = false">Отмена</button>
                            <button class="btn-primary" @click="sendChallenge">Отправить вызов</button>
                        </div>
                    </div>
                </div>
            </transition>
        </div>
    `
};

window.ActivityFeedComponent = ActivityFeedComponent;
window.ChallengesComponent = ChallengesComponent;
