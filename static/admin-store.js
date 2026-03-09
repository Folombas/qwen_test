// Vue.js Admin Store для админ-панели

const AdminStore = {
    state: {
        stats: null,
        users: [],
        activity: [],
        isLoading: false,
        error: null
    },

    // Инициализация
    init() {
        console.log('👨‍💼 AdminStore initialized');
    },

    // Получить статистику дашборда
    async fetchDashboardStats() {
        this.state.isLoading = true;
        try {
            const response = await fetch('/api/admin/dashboard');
            
            if (!response.ok) {
                throw new Error('Failed to fetch dashboard stats');
            }
            
            const data = await response.json();
            this.state.stats = data;
            return data;
        } catch (error) {
            this.state.error = error.message;
            console.error('Dashboard fetch error:', error);
            return null;
        } finally {
            this.state.isLoading = false;
        }
    },

    // Получить список пользователей
    async fetchUsers(page = 1, limit = 20, search = '') {
        this.state.isLoading = true;
        const offset = (page - 1) * limit;
        
        try {
            const url = `/api/admin/users?limit=${limit}&offset=${offset}&search=${encodeURIComponent(search)}`;
            const response = await fetch(url);
            
            if (!response.ok) {
                throw new Error('Failed to fetch users');
            }
            
            const data = await response.json();
            this.state.users = data.users;
            return {
                users: data.users,
                total: data.total,
                page: page,
                totalPages: Math.ceil(data.total / limit)
            };
        } catch (error) {
            this.state.error = error.message;
            console.error('Users fetch error:', error);
            return null;
        } finally {
            this.state.isLoading = false;
        }
    },

    // Получить пользователя по ID
    async fetchUser(userId) {
        try {
            const response = await fetch(`/api/admin/user?id=${userId}`);
            
            if (!response.ok) {
                throw new Error('Failed to fetch user');
            }
            
            return await response.json();
        } catch (error) {
            console.error('User fetch error:', error);
            return null;
        }
    },

    // Обновить пользователя
    async updateUser(userId, userData) {
        try {
            const response = await fetch(`/api/admin/user/update?id=${userId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(userData)
            });
            
            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to update user');
            }
            
            return { success: true };
        } catch (error) {
            console.error('User update error:', error);
            return { success: false, error: error.message };
        }
    },

    // Забанить пользователя
    async banUser(userId, reason) {
        try {
            const response = await fetch(`/api/admin/user/ban?id=${userId}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ reason })
            });
            
            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to ban user');
            }
            
            return { success: true };
        } catch (error) {
            console.error('Ban user error:', error);
            return { success: false, error: error.message };
        }
    },

    // Разбанить пользователя
    async unbanUser(userId) {
        try {
            const response = await fetch(`/api/admin/user/unban?id=${userId}`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error('Failed to unban user');
            }
            
            return { success: true };
        } catch (error) {
            console.error('Unban user error:', error);
            return { success: false, error: error.message };
        }
    },

    // Удалить пользователя
    async deleteUser(userId) {
        try {
            const response = await fetch(`/api/admin/user/delete?id=${userId}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error('Failed to delete user');
            }
            
            return { success: true };
        } catch (error) {
            console.error('Delete user error:', error);
            return { success: false, error: error.message };
        }
    },

    // Получить активность
    async fetchActivity(limit = 50) {
        try {
            const response = await fetch(`/api/admin/activity?limit=${limit}`);
            
            if (!response.ok) {
                throw new Error('Failed to fetch activity');
            }
            
            const data = await response.json();
            this.state.activity = data.activity;
            return data.activity;
        } catch (error) {
            console.error('Activity fetch error:', error);
            return [];
        }
    },

    // Геттеры
    getStats() {
        return this.state.stats;
    },

    getUsers() {
        return this.state.users;
    },

    getActivity() {
        return this.state.activity;
    },

    isLoading() {
        return this.state.isLoading;
    },

    clearError() {
        this.state.error = null;
    }
};

// Экспортируем глобально
window.AdminStore = AdminStore;
