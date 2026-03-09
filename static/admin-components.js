// Vue.js Admin Questions компонент (заглушка)

const AdminQuestions = {
    name: 'AdminQuestions',
    setup() {
        return {
            message: '📝 Управление вопросами в разработке...'
        };
    },
    template: `
        <div class="questions-page">
            <div class="placeholder">
                <div class="placeholder-icon">📝</div>
                <h2>Управление вопросами</h2>
                <p>{{ message }}</p>
                <div class="features-list">
                    <h3>Планируемый функционал:</h3>
                    <ul>
                        <li>➕ Добавление новых вопросов</li>
                        <li>✏️ Редактирование существующих</li>
                        <li>🗑️ Удаление вопросов</li>
                        <li>📊 Статистика по вопросам</li>
                        <li>📥 Импорт/экспорт вопросов</li>
                    </ul>
                </div>
            </div>
        </div>
    `
};

// Activity компонент
const AdminActivity = {
    name: 'AdminActivity',
    setup() {
        const { ref, onMounted } = Vue;
        const activity = ref([]);
        
        onMounted(async () => {
            activity.value = await AdminStore.fetchActivity(100);
        });
        
        function formatDate(dateStr) {
            if (!dateStr) return '-';
            const date = new Date(dateStr);
            return date.toLocaleDateString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        }
        
        return {
            activity,
            formatDate
        };
    },
    template: `
        <div class="activity-page">
            <h2>📜 Журнал активности</h2>
            
            <div v-if="!activity.length" class="loading">
                <p>Загрузка активности...</p>
            </div>
            
            <div v-else class="activity-table-container">
                <table class="activity-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Пользователь</th>
                            <th>Действие</th>
                            <th>Детали</th>
                            <th>Дата</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="log in activity" :key="log.id">
                            <td>{{ log.id }}</td>
                            <td>{{ log.user_name }}</td>
                            <td>
                                <span class="action-badge">{{ log.action }}</span>
                            </td>
                            <td>{{ log.details }}</td>
                            <td>{{ formatDate(log.created_at) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `
};

// Settings компонент
const AdminSettings = {
    name: 'AdminSettings',
    setup() {
        return {
            message: '⚙️ Настройки в разработке...'
        };
    },
    template: `
        <div class="settings-page">
            <div class="placeholder">
                <div class="placeholder-icon">⚙️</div>
                <h2>Настройки</h2>
                <p>{{ message }}</p>
                <div class="features-list">
                    <h3>Планируемый функционал:</h3>
                    <ul>
                        <li>🔧 Общие настройки приложения</li>
                        <li>📧 Настройки email уведомлений</li>
                        <li>🎮 Настройки игры</li>
                        <li>👥 Управление ролями</li>
                        <li>📊 Настройки статистики</li>
                    </ul>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.AdminQuestions = AdminQuestions;
window.AdminActivity = AdminActivity;
window.AdminSettings = AdminSettings;
