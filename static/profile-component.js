// Vue.js компонент Profile

const ProfileComponent = {
    name: 'Profile',
    setup() {
        const { ref, reactive, onMounted } = Vue;
        
        const user = ref(null);
        const isLoading = ref(true);
        const isEditing = ref(false);
        const showChangePassword = ref(false);
        
        const form = reactive({
            name: '',
            email: '',
            bio: '',
            website: '',
            location: ''
        });
        
        const passwordForm = reactive({
            oldPassword: '',
            newPassword: '',
            confirmPassword: ''
        });
        
        const errors = reactive({
            general: '',
            password: ''
        });
        
        const successMessage = ref('');
        
        // Загрузка профиля
        onMounted(async () => {
            await loadProfile();
        });
        
        async function loadProfile() {
            isLoading.value = true;
            try {
                const currentUser = await AuthStore.fetchCurrentUser();
                user.value = currentUser;
                if (currentUser) {
                    form.name = currentUser.name;
                    form.email = currentUser.email;
                }
            } catch (error) {
                errors.general = 'Failed to load profile';
            } finally {
                isLoading.value = false;
            }
        }
        
        function startEditing() {
            isEditing.value = true;
        }
        
        function cancelEditing() {
            isEditing.value = false;
            if (user.value) {
                form.name = user.value.name;
                form.email = user.value.email;
            }
        }
        
        async function saveProfile() {
            errors.general = '';
            // TODO: API для обновления профиля
            alert('Функция обновления профиля в разработке!');
            isEditing.value = false;
        }
        
        async function handleChangePassword() {
            errors.password = '';
            successMessage.value = '';
            
            if (passwordForm.newPassword !== passwordForm.confirmPassword) {
                errors.password = 'Пароли не совпадают';
                return;
            }
            
            if (passwordForm.newPassword.length < 6) {
                errors.password = 'Минимум 6 символов';
                return;
            }
            
            const result = await AuthStore.changePassword(
                passwordForm.oldPassword,
                passwordForm.newPassword
            );
            
            if (result.success) {
                successMessage.value = '✅ Пароль успешно изменён!';
                passwordForm.oldPassword = '';
                passwordForm.newPassword = '';
                passwordForm.confirmPassword = '';
                showChangePassword.value = false;
            } else {
                errors.password = result.error;
            }
        }
        
        async function handleLogout() {
            if (confirm('Вы уверены что хотите выйти?')) {
                await AuthStore.logout();
                window.location.hash = '';
            }
        }
        
        function toggleChangePassword() {
            showChangePassword.value = !showChangePassword.value;
            errors.password = '';
            successMessage.value = '';
        }
        
        return {
            user,
            isLoading,
            isEditing,
            showChangePassword,
            form,
            passwordForm,
            errors,
            successMessage,
            loadProfile,
            startEditing,
            cancelEditing,
            saveProfile,
            handleChangePassword,
            handleLogout,
            toggleChangePassword
        };
    },
    template: `
        <div class="profile-page">
            <div class="profile-container">
                <div v-if="isLoading" class="loading-profile">
                    <div class="spinner"></div>
                    <p>Загрузка профиля...</p>
                </div>
                
                <div v-else class="profile-content">
                    <!-- Profile Header -->
                    <div class="profile-header">
                        <div class="profile-avatar">
                            <span class="avatar-letter">{{ user?.name?.charAt(0) || 'U' }}</span>
                        </div>
                        <div class="profile-info">
                            <h1 class="profile-name">{{ user?.name || 'User' }}</h1>
                            <p class="profile-email">{{ user?.email || '' }}</p>
                            <div class="profile-badges">
                                <span class="badge" :class="'badge-' + (user?.role || 'user')">
                                    {{ user?.role || 'user' }}
                                </span>
                                <span v-if="user?.email_verified" class="badge badge-verified">
                                    ✓ verified
                                </span>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Stats -->
                    <div class="profile-stats">
                        <div class="stat-card">
                            <div class="stat-value">🏆</div>
                            <div class="stat-label">Уровень</div>
                        </div>
                        <div class="stat-card">
                            <div class="stat-value">⚡</div>
                            <div class="stat-label">EXP</div>
                        </div>
                        <div class="stat-card">
                            <div class="stat-value">📚</div>
                            <div class="stat-label">Знание Go</div>
                        </div>
                        <div class="stat-card">
                            <div class="stat-value">🔥</div>
                            <div class="stat-label">Серия</div>
                        </div>
                    </div>
                    
                    <!-- Actions -->
                    <div class="profile-actions">
                        <button class="action-btn" @click="startEditing">✏️ Редактировать</button>
                        <button class="action-btn" @click="toggleChangePassword">🔑 Сменить пароль</button>
                        <button class="action-btn danger" @click="handleLogout">🚪 Выйти</button>
                    </div>
                    
                    <!-- Change Password Form -->
                    <div v-if="showChangePassword" class="password-form">
                        <h3>🔑 Смена пароля</h3>
                        
                        <div v-if="errors.password" class="error-message">{{ errors.password }}</div>
                        <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
                        
                        <div class="form-group">
                            <label class="form-label">Текущий пароль</label>
                            <input v-model="passwordForm.oldPassword" type="password" class="form-input" placeholder="••••••••" />
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Новый пароль</label>
                            <input v-model="passwordForm.newPassword" type="password" class="form-input" placeholder="••••••••" />
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Подтвердите пароль</label>
                            <input v-model="passwordForm.confirmPassword" type="password" class="form-input" placeholder="••••••••" />
                        </div>
                        
                        <div class="form-actions">
                            <button class="btn-primary" @click="handleChangePassword">Сохранить</button>
                            <button class="btn-secondary" @click="toggleChangePassword">Отмена</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.ProfileComponent = ProfileComponent;
