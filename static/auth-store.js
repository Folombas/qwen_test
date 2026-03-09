// Vue.js Auth Store для управления аутентификацией

const AuthStore = {
    state: {
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
        error: null
    },

    // Инициализация
    init() {
        // Проверяем сохранённые токены
        const savedTokens = localStorage.getItem('auth_tokens');
        const savedUser = localStorage.getItem('auth_user');
        
        if (savedTokens) {
            const tokens = JSON.parse(savedTokens);
            this.state.accessToken = tokens.access_token;
            this.state.refreshToken = tokens.refresh_token;
        }
        
        if (savedUser) {
            this.state.user = JSON.parse(savedUser);
            this.state.isAuthenticated = true;
        }
        
        // Настраиваем interceptor для всех fetch запросов
        this.setupInterceptor();
    },

    // Setup interceptor для добавления токена к запросам
    setupInterceptor() {
        const originalFetch = window.fetch;
        const self = this;
        
        window.fetch = async function(url, options = {}) {
            // Добавляем токен к API запросам
            if (url.includes('/api/')) {
                options.headers = {
                    ...options.headers,
                    'Content-Type': 'application/json',
                };
                
                // Добавляем Authorization header если есть токен
                if (self.state.accessToken) {
                    options.headers['Authorization'] = 'Bearer ' + self.state.accessToken;
                }
            }
            
            try {
                const response = await originalFetch.call(this, url, options);
                
                // Проверяем на 401 (токен истёк)
                if (response.status === 401 && url.includes('/api/')) {
                    // Пробуем refreshнуть токен
                    const refreshed = await self.refreshToken();
                    if (refreshed) {
                        // Повторяем оригинальный запрос
                        options.headers['Authorization'] = 'Bearer ' + self.state.accessToken;
                        return originalFetch.call(this, url, options);
                    } else {
                        // Refresh не удался, logout
                        self.logout();
                        window.location.href = '/#/login';
                    }
                }
                
                return response;
            } catch (error) {
                console.error('Fetch error:', error);
                throw error;
            }
        };
    },

    // Регистрация
    async register(email, password, name) {
        this.state.isLoading = true;
        this.state.error = null;
        
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                body: JSON.stringify({ email, password, name })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Registration failed');
            }
            
            // Сохраняем токены и пользователя
            this.setTokens(data.tokens);
            this.setUser(data.user);
            
            return { success: true, user: data.user };
        } catch (error) {
            this.state.error = error.message;
            return { success: false, error: error.message };
        } finally {
            this.state.isLoading = false;
        }
    },

    // Вход
    async login(email, password) {
        this.state.isLoading = true;
        this.state.error = null;
        
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                body: JSON.stringify({ email, password })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Login failed');
            }
            
            // Сохраняем токены и пользователя
            this.setTokens(data.tokens);
            this.setUser(data.user);
            
            return { success: true, user: data.user };
        } catch (error) {
            this.state.error = error.message;
            return { success: false, error: error.message };
        } finally {
            this.state.isLoading = false;
        }
    },

    // Выход
    async logout() {
        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                body: JSON.stringify({ refresh_token: this.state.refreshToken })
            });
        } catch (error) {
            console.error('Logout error:', error);
        }
        
        // Очищаем состояние
        this.clearAuth();
    },

    // Refresh токена
    async refreshToken() {
        if (!this.state.refreshToken) {
            return false;
        }
        
        try {
            const response = await fetch('/api/auth/refresh', {
                method: 'POST',
                body: JSON.stringify({ refresh_token: this.state.refreshToken })
            });
            
            const data = await response.json();
            
            if (response.ok && data.tokens) {
                this.setTokens(data.tokens);
                return true;
            }
        } catch (error) {
            console.error('Token refresh error:', error);
        }
        
        return false;
    },

    // Получить текущего пользователя
    async fetchCurrentUser() {
        try {
            const response = await fetch('/api/auth/me');
            
            if (response.ok) {
                const user = await response.json();
                this.setUser(user);
                return user;
            }
        } catch (error) {
            console.error('Fetch user error:', error);
        }
        
        return null;
    },

    // Смена пароля
    async changePassword(oldPassword, newPassword) {
        try {
            const response = await fetch('/api/auth/change-password', {
                method: 'POST',
                body: JSON.stringify({
                    old_password: oldPassword,
                    new_password: newPassword
                })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Password change failed');
            }
            
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Восстановление пароля
    async forgotPassword(email) {
        try {
            const response = await fetch('/api/auth/forgot-password', {
                method: 'POST',
                body: JSON.stringify({ email })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Failed to send reset email');
            }
            
            return { success: true, resetToken: data.reset_token };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Сброс пароля
    async resetPassword(token, newPassword) {
        try {
            const response = await fetch('/api/auth/reset-password', {
                method: 'POST',
                body: JSON.stringify({
                    token: token,
                    new_password: newPassword
                })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Password reset failed');
            }
            
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Подтверждение email
    async verifyEmail(token) {
        try {
            const response = await fetch('/api/auth/verify-email', {
                method: 'POST',
                body: JSON.stringify({ token })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Email verification failed');
            }
            
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Установить токены
    setTokens(tokens) {
        this.state.accessToken = tokens.access_token;
        this.state.refreshToken = tokens.refresh_token;
        localStorage.setItem('auth_tokens', JSON.stringify(tokens));
    },

    // Установить пользователя
    setUser(user) {
        this.state.user = user;
        this.state.isAuthenticated = true;
        localStorage.setItem('auth_user', JSON.stringify(user));
    },

    // Очистить аутентификацию
    clearAuth() {
        this.state.user = null;
        this.state.accessToken = null;
        this.state.refreshToken = null;
        this.state.isAuthenticated = false;
        this.state.error = null;
        localStorage.removeItem('auth_tokens');
        localStorage.removeItem('auth_user');
    },

    // Геттеры
    getUser() {
        return this.state.user;
    },

    isAuthenticated() {
        return this.state.isAuthenticated;
    },

    isLoading() {
        return this.state.isLoading;
    },

    getError() {
        return this.state.error;
    },

    clearError() {
        this.state.error = null;
    }
};

// Экспортируем глобально
window.AuthStore = AuthStore;
