// Vue.js компонент Register

const RegisterComponent = {
    name: 'Register',
    setup() {
        const { ref, reactive } = Vue;
        
        // Состояние формы
        const form = reactive({
            name: '',
            email: '',
            password: '',
            confirmPassword: '',
            acceptTerms: false
        });
        
        const errors = reactive({
            name: '',
            email: '',
            password: '',
            confirmPassword: '',
            acceptTerms: '',
            general: ''
        });
        
        const isLoading = ref(false);
        const showPassword = ref(false);
        const showConfirmPassword = ref(false);
        const passwordStrength = ref(0);
        
        // Валидация имени
        function validateName() {
            if (!form.name) {
                errors.name = 'Имя обязательно';
                return false;
            }
            if (form.name.length < 2) {
                errors.name = 'Минимум 2 символа';
                return false;
            }
            if (form.name.length > 50) {
                errors.name = 'Максимум 50 символов';
                return false;
            }
            errors.name = '';
            return true;
        }
        
        // Валидация email
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
        
        // Валидация пароля
        function validatePassword() {
            if (!form.password) {
                errors.password = 'Пароль обязателен';
                return false;
            }
            if (form.password.length < 6) {
                errors.password = 'Минимум 6 символов';
                return false;
            }
            if (form.password.length > 100) {
                errors.password = 'Максимум 100 символов';
                return false;
            }
            
            // Проверка сложности пароля
            let strength = 0;
            if (form.password.length >= 8) strength++;
            if (/[a-z]/.test(form.password)) strength++;
            if (/[A-Z]/.test(form.password)) strength++;
            if (/[0-9]/.test(form.password)) strength++;
            if (/[^a-zA-Z0-9]/.test(form.password)) strength++;
            
            passwordStrength.value = strength;
            
            if (strength < 3) {
                errors.password = 'Пароль слишком слабый';
                return false;
            }
            
            errors.password = '';
            return true;
        }
        
        // Валидация подтверждения пароля
        function validateConfirmPassword() {
            if (!form.confirmPassword) {
                errors.confirmPassword = 'Подтвердите пароль';
                return false;
            }
            if (form.confirmPassword !== form.password) {
                errors.confirmPassword = 'Пароли не совпадают';
                return false;
            }
            errors.confirmPassword = '';
            return true;
        }
        
        // Валидация условий
        function validateAcceptTerms() {
            if (!form.acceptTerms) {
                errors.acceptTerms = 'Необходимо принять условия';
                return false;
            }
            errors.acceptTerms = '';
            return true;
        }
        
        // Отправка формы
        async function handleSubmit() {
            errors.general = '';
            
            const nameValid = validateName();
            const emailValid = validateEmail();
            const passwordValid = validatePassword();
            const confirmValid = validateConfirmPassword();
            const termsValid = validateAcceptTerms();
            
            if (!nameValid || !emailValid || !passwordValid || !confirmValid || !termsValid) {
                return;
            }
            
            isLoading.value = true;
            
            try {
                const result = await AuthStore.register(form.email, form.password, form.name);
                
                if (result.success) {
                    // Успешная регистрация
                    alert('✅ Регистрация успешна! Добро пожаловать, ' + form.name + '!');
                    window.location.hash = '';
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
        function goToLogin() {
            window.location.hash = 'login';
        }
        
        // Вычисление сложности пароля
        function getPasswordStrengthColor() {
            const colors = ['#ef4444', '#f59e0b', '#f59e0b', '#10b981', '#10b981'];
            return colors[passwordStrength.value] || '#ef4444';
        }
        
        function getPasswordStrengthText() {
            const texts = ['Очень слабый', 'Слабый', 'Средний', 'Хороший', 'Отличный'];
            return texts[passwordStrength.value] || '';
        }
        
        return {
            form,
            errors,
            isLoading,
            showPassword,
            showConfirmPassword,
            passwordStrength,
            validateName,
            validateEmail,
            validatePassword,
            validateConfirmPassword,
            validateAcceptTerms,
            handleSubmit,
            goToLogin,
            getPasswordStrengthColor,
            getPasswordStrengthText
        };
    },
    template: `
        <div class="auth-page">
            <div class="auth-container">
                <div class="auth-card">
                    <div class="auth-header">
                        <h1 class="auth-title">🎉 Регистрация</h1>
                        <p class="auth-subtitle">Создайте аккаунт для начала</p>
                    </div>
                    
                    <form @submit.prevent="handleSubmit" class="auth-form">
                        <!-- Name -->
                        <div class="form-group" :class="{ 'has-error': errors.name }">
                            <label class="form-label" for="name">
                                👤 Имя
                            </label>
                            <input
                                id="name"
                                v-model="form.name"
                                type="text"
                                class="form-input"
                                :class="{ 'input-error': errors.name }"
                                placeholder="Ваше имя"
                                @blur="validateName"
                                autocomplete="name"
                            />
                            <span v-if="errors.name" class="error-message">{{ errors.name }}</span>
                        </div>
                        
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
                                    @input="validatePassword"
                                    autocomplete="new-password"
                                />
                                <button
                                    type="button"
                                    class="password-toggle"
                                    @click="showPassword = !showPassword"
                                >
                                    {{ showPassword ? '🙈' : '👁️' }}
                                </button>
                            </div>
                            
                            <!-- Password Strength -->
                            <div v-if="form.password" class="password-strength">
                                <div class="strength-bar">
                                    <div class="strength-fill" :style="{ width: (passwordStrength / 5) * 100 + '%', background: getPasswordStrengthColor() }"></div>
                                </div>
                                <span class="strength-text" :style="{ color: getPasswordStrengthColor() }">
                                    {{ getPasswordStrengthText() }}
                                </span>
                            </div>
                            
                            <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
                        </div>
                        
                        <!-- Confirm Password -->
                        <div class="form-group" :class="{ 'has-error': errors.confirmPassword }">
                            <label class="form-label" for="confirmPassword">
                                🔒 Подтвердите пароль
                            </label>
                            <div class="password-input-wrapper">
                                <input
                                    id="confirmPassword"
                                    v-model="form.confirmPassword"
                                    :type="showConfirmPassword ? 'text' : 'password'"
                                    class="form-input"
                                    :class="{ 'input-error': errors.confirmPassword }"
                                    placeholder="••••••••"
                                    @blur="validateConfirmPassword"
                                    autocomplete="new-password"
                                />
                                <button
                                    type="button"
                                    class="password-toggle"
                                    @click="showConfirmPassword = !showConfirmPassword"
                                >
                                    {{ showConfirmPassword ? '🙈' : '👁️' }}
                                </button>
                            </div>
                            <span v-if="errors.confirmPassword" class="error-message">{{ errors.confirmPassword }}</span>
                        </div>
                        
                        <!-- Accept Terms -->
                        <div class="form-group" :class="{ 'has-error': errors.acceptTerms }">
                            <label class="checkbox-label">
                                <input type="checkbox" v-model="form.acceptTerms" class="checkbox" />
                                <span>
                                    Я принимаю
                                    <a href="#" class="link">Условия использования</a>
                                    и
                                    <a href="#" class="link">Политику конфиденциальности</a>
                                </span>
                            </label>
                            <span v-if="errors.acceptTerms" class="error-message">{{ errors.acceptTerms }}</span>
                        </div>
                        
                        <!-- General Error -->
                        <div v-if="errors.general" class="general-error">
                            ❌ {{ errors.general }}
                        </div>
                        
                        <!-- Submit Button -->
                        <button type="submit" class="auth-btn" :disabled="isLoading">
                            <span v-if="!isLoading">🚀 Создать аккаунт</span>
                            <span v-else class="loading-spinner">⏳ Создание...</span>
                        </button>
                    </form>
                    
                    <!-- Login Link -->
                    <div class="auth-footer">
                        Уже есть аккаунт?
                        <button class="link-btn" @click="goToLogin">Войти</button>
                    </div>
                </div>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.RegisterComponent = RegisterComponent;
