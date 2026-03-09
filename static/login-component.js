// Vue.js компонент Login

const LoginComponent = {
    name: 'Login',
    setup() {
        const { ref, reactive } = Vue;
        
        // Состояние формы
        const form = reactive({
            email: '',
            password: '',
            rememberMe: false
        });
        
        const errors = reactive({
            email: '',
            password: '',
            general: ''
        });
        
        const isLoading = ref(false);
        const showPassword = ref(false);
        
        // Валидация
        function validateEmail() {
            if (!form.email) {
                errors.email = 'Email обязателен';
                return false;
            }
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(form.email)) {
                errors.email = 'Некорректный email';
                return false;
            }
            errors.email = '';
            return true;
        }
        
        function validatePassword() {
            if (!form.password) {
                errors.password = 'Пароль обязателен';
                return false;
            }
            if (form.password.length < 6) {
                errors.password = 'Минимум 6 символов';
                return false;
            }
            errors.password = '';
            return true;
        }
        
        // Отправка формы
        async function handleSubmit() {
            errors.general = '';
            
            const emailValid = validateEmail();
            const passwordValid = validatePassword();
            
            if (!emailValid || !passwordValid) {
                return;
            }
            
            isLoading.value = true;
            
            try {
                const result = await AuthStore.login(form.email, form.password);
                
                if (result.success) {
                    // Перенаправляем на главную или куда планировал пользователь
                    const redirect = localStorage.getItem('login_redirect') || '/';
                    localStorage.removeItem('login_redirect');
                    window.location.hash = redirect;
                } else {
                    errors.general = result.error;
                }
            } catch (error) {
                errors.general = 'Ошибка соединения с сервером';
            } finally {
                isLoading.value = false;
            }
        }
        
        // Навигация
        function goToRegister() {
            window.location.hash = 'register';
        }
        
        function goToForgotPassword() {
            window.location.hash = 'forgot-password';
        }
        
        return {
            form,
            errors,
            isLoading,
            showPassword,
            validateEmail,
            validatePassword,
            handleSubmit,
            goToRegister,
            goToForgotPassword
        };
    },
    template: `
        <div class="auth-page">
            <div class="auth-container">
                <div class="auth-card">
                    <div class="auth-header">
                        <h1 class="auth-title">🔐 Вход</h1>
                        <p class="auth-subtitle">Войдите для продолжения</p>
                    </div>
                    
                    <form @submit.prevent="handleSubmit" class="auth-form">
                        <!-- Email -->
                        <div class="form-group" :class="{ 'has-error': errors.email }">
                            <label class="form-label" for="email">
                                📧 Email
                            </label>
                            <input
                                id="email"
                                v-model="form.email"
                                type="email"
                                class="form-input"
                                :class="{ 'input-error': errors.email }"
                                placeholder="your@email.com"
                                @blur="validateEmail"
                                autocomplete="email"
                            />
                            <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
                        </div>
                        
                        <!-- Password -->
                        <div class="form-group" :class="{ 'has-error': errors.password }">
                            <label class="form-label" for="password">
                                🔑 Пароль
                            </label>
                            <div class="password-input-wrapper">
                                <input
                                    id="password"
                                    v-model="form.password"
                                    :type="showPassword ? 'text' : 'password'"
                                    class="form-input"
                                    :class="{ 'input-error': errors.password }"
                                    placeholder="••••••••"
                                    @blur="validatePassword"
                                    autocomplete="current-password"
                                />
                                <button
                                    type="button"
                                    class="password-toggle"
                                    @click="showPassword = !showPassword"
                                >
                                    {{ showPassword ? '🙈' : '👁️' }}
                                </button>
                            </div>
                            <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
                        </div>
                        
                        <!-- Remember & Forgot -->
                        <div class="form-options">
                            <label class="checkbox-label">
                                <input type="checkbox" v-model="form.rememberMe" class="checkbox" />
                                <span>Запомнить меня</span>
                            </label>
                            <button type="button" class="forgot-link" @click="goToForgotPassword">
                                Забыли пароль?
                            </button>
                        </div>
                        
                        <!-- General Error -->
                        <div v-if="errors.general" class="general-error">
                            ❌ {{ errors.general }}
                        </div>
                        
                        <!-- Submit Button -->
                        <button type="submit" class="auth-btn" :disabled="isLoading">
                            <span v-if="!isLoading">🚀 Войти</span>
                            <span v-else class="loading-spinner">⏳ Загрузка...</span>
                        </button>
                    </form>
                    
                    <!-- Divider -->
                    <div class="auth-divider">
                        <span>или</span>
                    </div>
                    
                    <!-- Social Login (placeholder) -->
                    <div class="social-login">
                        <button class="social-btn google" disabled title="Скоро">
                            <span class="social-icon">G</span>
                            Войти через Google
                        </button>
                        <button class="social-btn github" disabled title="Скоро">
                            <span class="social-icon">GH</span>
                            Войти через GitHub
                        </button>
                    </div>
                    
                    <!-- Register Link -->
                    <div class="auth-footer">
                        Нет аккаунта?
                        <button class="link-btn" @click="goToRegister">Зарегистрироваться</button>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.LoginComponent = LoginComponent;
