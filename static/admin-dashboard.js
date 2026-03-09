// Vue.js Admin Dashboard компонент

const AdminDashboard = {
    name: 'AdminDashboard',
    setup() {
        const { ref, reactive, onMounted } = Vue;
        
        const stats = ref(null);
        const isLoading = ref(true);
        const activity = ref([]);
        
        onMounted(async () => {
            await loadDashboard();
            await loadActivity();
        });
        
        async function loadDashboard() {
            isLoading.value = true;
            stats.value = await AdminStore.fetchDashboardStats();
            isLoading.value = false;
        }
        
        async function loadActivity() {
            activity.value = await AdminStore.fetchActivity(10);
        }
        
        function formatNumber(num) {
            if (num === null || num === undefined) return '0';
            if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
            if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
            return num.toString();
        }
        
        function formatDate(dateStr) {
            if (!dateStr) return '-';
            const date = new Date(dateStr);
            return date.toLocaleDateString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
        }
        
        return {
            stats,
            isLoading,
            activity,
            loadDashboard,
            formatNumber,
            formatDate
        };
    },
    template: `
        <div class="dashboard">
            <div v-if="isLoading" class="loading">
                <div class="spinner"></div>
                <p>Загрузка статистики...</p>
            </div>
            
            <div v-else class="dashboard-content">
                <!-- Stats Cards -->
                <div class="stats-grid">
                    <div class="stat-card primary">
                        <div class="stat-icon">👥</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.total_users) }}</div>
                            <div class="stat-label">Всего пользователей</div>
                        </div>
                    </div>
                    
                    <div class="stat-card success">
                        <div class="stat-icon">✅</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.active_users) }}</div>
                            <div class="stat-label">Активные</div>
                        </div>
                    </div>
                    
                    <div class="stat-card danger">
                        <div class="stat-icon">🚫</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.banned_users) }}</div>
                            <div class="stat-label">Забаненные</div>
                        </div>
                    </div>
                    
                    <div class="stat-card warning">
                        <div class="stat-icon">📝</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.total_questions) }}</div>
                            <div class="stat-label">Вопросов</div>
                        </div>
                    </div>
                    
                    <div class="stat-card info">
                        <div class="stat-icon">📅</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.new_users_today) }}</div>
                            <div class="stat-label">Новых сегодня</div>
                        </div>
                    </div>
                    
                    <div class="stat-card info">
                        <div class="stat-icon">📆</div>
                        <div class="stat-info">
                            <div class="stat-value">{{ formatNumber(stats?.new_users_week) }}</div>
                            <div class="stat-label">Новых за неделю</div>
                        </div>
                    </div>
                </div>
                
                <!-- Two Columns -->
                <div class="dashboard-grid">
                    <!-- Top Players -->
                    <div class="dashboard-card">
                        <div class="card-header">
                            <h3>🏆 Топ игроков</h3>
                        </div>
                        <div class="card-body">
                            <table class="data-table">
                                <thead>
                                    <tr>
                                        <th>#</th>
                                        <th>Игрок</th>
                                        <th>Уровень</th>
                                        <th>Рейтинг</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(player, idx) in (stats?.top_players || []).slice(0, 5)" :key="player.id">
                                        <td>{{ idx + 1 }}</td>
                                        <td>
                                            <div class="player-name">{{ player.name }}</div>
                                            <div class="player-email">{{ player.email }}</div>
                                        </td>
                                        <td>Ур. {{ player.level }}</td>
                                        <td class="rating">{{ player.rating }}</td>
                                    </tr>
                                    <tr v-if="!stats?.top_players?.length">
                                        <td colspan="4" class="empty">Нет данных</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                    
                    <!-- Recent Activity -->
                    <div class="dashboard-card">
                        <div class="card-header">
                            <h3>📜 Последняя активность</h3>
                        </div>
                        <div class="card-body">
                            <div class="activity-list">
                                <div v-for="log in activity" :key="log.id" class="activity-item">
                                    <div class="activity-icon">
                                        {{ log.action === 'ban' ? '🔒' : log.action === 'unban' ? '🔓' : '📝' }}
                                    </div>
                                    <div class="activity-info">
                                        <div class="activity-user">{{ log.user_name }}</div>
                                        <div class="activity-action">{{ log.action }}: {{ log.details }}</div>
                                    </div>
                                    <div class="activity-time">{{ formatDate(log.created_at) }}</div>
                                </div>
                                <div v-if="!activity.length" class="empty">Нет активности</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.AdminDashboard = AdminDashboard;
