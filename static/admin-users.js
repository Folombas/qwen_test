// Vue.js Admin Users компонент

const AdminUsers = {
    name: 'AdminUsers',
    setup() {
        const { ref, reactive, onMounted } = Vue;
        
        const users = ref([]);
        const isLoading = ref(false);
        const searchQuery = ref('');
        const currentPage = ref(1);
        const totalPages = ref(1);
        const totalUsers = ref(0);
        const selectedUser = ref(null);
        const showEditModal = ref(false);
        const showBanModal = ref(false);
        const banReason = ref('');
        
        const editForm = reactive({
            name: '',
            email: '',
            role: 'user',
            isActive: true,
            isBanned: false
        });
        
        onMounted(() => {
            loadUsers();
        });
        
        async function loadUsers(page = 1) {
            isLoading.value = true;
            const result = await AdminStore.fetchUsers(page, 20, searchQuery.value);
            if (result) {
                users.value = result.users;
                currentPage.value = result.page;
                totalPages.value = result.totalPages;
                totalUsers.value = result.total;
            }
            isLoading.value = false;
        }
        
        function search() {
            currentPage.value = 1;
            loadUsers(1);
        }
        
        function prevPage() {
            if (currentPage.value > 1) {
                loadUsers(currentPage.value - 1);
            }
        }
        
        function nextPage() {
            if (currentPage.value < totalPages.value) {
                loadUsers(currentPage.value + 1);
            }
        }
        
        function openEditUser(user) {
            selectedUser.value = user;
            editForm.name = user.name;
            editForm.email = user.email;
            editForm.role = user.role;
            editForm.isActive = user.is_active;
            editForm.isBanned = user.is_banned;
            showEditModal.value = true;
        }
        
        async function saveUser() {
            if (!selectedUser.value) return;
            
            const result = await AdminStore.updateUser(selectedUser.value.id, editForm);
            if (result.success) {
                alert('✅ Пользователь обновлён');
                showEditModal.value = false;
                loadUsers(currentPage.value);
            } else {
                alert('❌ Ошибка: ' + result.error);
            }
        }
        
        function openBanUser(user) {
            selectedUser.value = user;
            banReason.value = '';
            showBanModal.value = true;
        }
        
        async function confirmBan() {
            if (!selectedUser.value) return;
            
            const result = await AdminStore.banUser(selectedUser.value.id, banReason.value);
            if (result.success) {
                alert('✅ Пользователь забанен');
                showBanModal.value = false;
                loadUsers(currentPage.value);
            } else {
                alert('❌ Ошибка: ' + result.error);
            }
        }
        
        async function unbanUser(user) {
            if (!confirm('Разбанить пользователя ' + user.name + '?')) return;
            
            const result = await AdminStore.unbanUser(user.id);
            if (result.success) {
                alert('✅ Пользователь разбанен');
                loadUsers(currentPage.value);
            } else {
                alert('❌ Ошибка: ' + result.error);
            }
        }
        
        async function deleteUser(user) {
            if (!confirm('Вы уверены что хотите удалить пользователя ' + user.name + '? Это действие необратимо.')) return;
            
            const result = await AdminStore.deleteUser(user.id);
            if (result.success) {
                alert('✅ Пользователь удалён');
                loadUsers(currentPage.value);
            } else {
                alert('❌ Ошибка: ' + result.error);
            }
        }
        
        function closeEditModal() {
            showEditModal.value = false;
            selectedUser.value = null;
        }
        
        function closeBanModal() {
            showBanModal.value = false;
            selectedUser.value = null;
        }
        
        function getRoleBadgeClass(role) {
            const classes = {
                'admin': 'badge-admin',
                'moderator': 'badge-mod',
                'user': 'badge-user'
            };
            return classes[role] || 'badge-user';
        }
        
        return {
            users,
            isLoading,
            searchQuery,
            currentPage,
            totalPages,
            totalUsers,
            selectedUser,
            showEditModal,
            showBanModal,
            editForm,
            banReason,
            loadUsers,
            search,
            prevPage,
            nextPage,
            openEditUser,
            saveUser,
            openBanUser,
            confirmBan,
            unbanUser,
            deleteUser,
            closeEditModal,
            closeBanModal,
            getRoleBadgeClass
        };
    },
    template: `
        <div class="users-page">
            <!-- Header -->
            <div class="page-header">
                <div class="search-box">
                    <input 
                        v-model="searchQuery" 
                        type="text" 
                        placeholder="🔍 Поиск по имени или email..."
                        @keyup.enter="search"
                        class="search-input"
                    />
                    <button @click="search" class="btn-search">Найти</button>
                </div>
                <div class="page-info">
                    Всего: {{ totalUsers }} пользователей
                </div>
            </div>
            
            <!-- Users Table -->
            <div v-if="isLoading" class="loading">
                <div class="spinner"></div>
                <p>Загрузка пользователей...</p>
            </div>
            
            <div v-else class="users-table-container">
                <table class="users-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Пользователь</th>
                            <th>Роль</th>
                            <th>Статус</th>
                            <th>Уровень</th>
                            <th>Действия</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="user in users" :key="user.id">
                            <td>{{ user.id }}</td>
                            <td>
                                <div class="user-info">
                                    <div class="user-avatar">{{ user.name.charAt(0) }}</div>
                                    <div>
                                        <div class="user-name">{{ user.name }}</div>
                                        <div class="user-email">{{ user.email }}</div>
                                    </div>
                                </div>
                            </td>
                            <td>
                                <span class="badge" :class="getRoleBadgeClass(user.role)">
                                    {{ user.role }}
                                </span>
                            </td>
                            <td>
                                <span v-if="user.is_banned" class="status-banned">🚫 Забанен</span>
                                <span v-else-if="!user.is_active" class="status-inactive">⏸️ Неактивен</span>
                                <span v-else class="status-active">✅ Активен</span>
                            </td>
                            <td>Ур. {{ user.stats?.level || 1 }}</td>
                            <td>
                                <div class="action-buttons">
                                    <button class="btn-edit" @click="openEditUser(user)" title="Редактировать">✏️</button>
                                    <button v-if="user.is_banned" class="btn-unban" @click="unbanUser(user)" title="Разбанить">🔓</button>
                                    <button v-else class="btn-ban" @click="openBanUser(user)" title="Забанить">🔒</button>
                                    <button class="btn-delete" @click="deleteUser(user)" title="Удалить">🗑️</button>
                                </div>
                            </td>
                        </tr>
                        <tr v-if="!users.length">
                            <td colspan="6" class="empty">Пользователи не найдены</td>
                        </tr>
                    </tbody>
                </table>
                
                <!-- Pagination -->
                <div class="pagination">
                    <button @click="prevPage" :disabled="currentPage === 1" class="btn-page">← Назад</button>
                    <span class="page-num">{{ currentPage }} / {{ totalPages }}</span>
                    <button @click="nextPage" :disabled="currentPage === totalPages" class="btn-page">Вперёд →</button>
                </div>
            </div>
            
            <!-- Edit Modal -->
            <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
                <div class="modal">
                    <div class="modal-header">
                        <h3>✏️ Редактирование пользователя</h3>
                        <button class="modal-close" @click="closeEditModal">×</button>
                    </div>
                    <div class="modal-body">
                        <div class="form-group">
                            <label class="form-label">Имя</label>
                            <input v-model="editForm.name" type="text" class="form-input" />
                        </div>
                        <div class="form-group">
                            <label class="form-label">Email</label>
                            <input v-model="editForm.email" type="email" class="form-input" />
                        </div>
                        <div class="form-group">
                            <label class="form-label">Роль</label>
                            <select v-model="editForm.role" class="form-input">
                                <option value="user">user</option>
                                <option value="moderator">moderator</option>
                                <option value="admin">admin</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" v-model="editForm.isActive" />
                                Активен
                            </label>
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" v-model="editForm.isBanned" />
                                Забанен
                            </label>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button @click="closeEditModal" class="btn-secondary">Отмена</button>
                        <button @click="saveUser" class="btn-primary">Сохранить</button>
                    </div>
                </div>
            </div>
            
            <!-- Ban Modal -->
            <div v-if="showBanModal" class="modal-overlay" @click.self="closeBanModal">
                <div class="modal">
                    <div class="modal-header">
                        <h3>🔒 Бан пользователя</h3>
                        <button class="modal-close" @click="closeBanModal">×</button>
                    </div>
                    <div class="modal-body">
                        <p>Пользователь: <strong>{{ selectedUser?.name }}</strong></p>
                        <div class="form-group">
                            <label class="form-label">Причина бана</label>
                            <textarea v-model="banReason" class="form-input" rows="4" placeholder="Укажите причину бана..."></textarea>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button @click="closeBanModal" class="btn-secondary">Отмена</button>
                        <button @click="confirmBan" class="btn-danger">Забанить</button>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.AdminUsers = AdminUsers;
