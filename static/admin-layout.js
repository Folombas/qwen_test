// Vue.js Admin Layout компонент

const AdminLayout = {
    name: 'AdminLayout',
    setup() {
        const { ref, reactive, onMounted } = Vue;
        
        const currentView = ref('dashboard');
        const sidebarOpen = ref(true);
        const isLoading = ref(false);
        const adminName = ref('Admin');
        
        // Навигация
        const menuItems = [
            { id: 'dashboard', icon: '📊', label: 'Дашборд' },
            { id: 'users', icon: '👥', label: 'Пользователи' },
            { id: 'questions', icon: '📝', label: 'Вопросы' },
            { id: 'activity', icon: '📜', label: 'Активность' },
            { id: 'settings', icon: '⚙️', label: 'Настройки' }
        ];
        
        onMounted(() => {
            // Проверяем что пользователь админ
            const user = AuthStore.getUser();
            if (!user || user.role !== 'admin') {
                alert('❌ Доступ запрещён. Требуются права администратора.');
                window.location.hash = '';
            }
        });
        
        function navigate(view) {
            currentView.value = view;
        }
        
        function toggleSidebar() {
            sidebarOpen.value = !sidebarOpen.value;
        }
        
        function goBack() {
            window.location.hash = '';
        }
        
        return {
            currentView,
            sidebarOpen,
            isLoading,
            adminName,
            menuItems,
            navigate,
            toggleSidebar,
            goBack
        };
    },
    template: `
        <div class="admin-layout">
            <!-- Sidebar -->
            <aside class="admin-sidebar" :class="{ 'sidebar-closed': !sidebarOpen }">
                <div class="sidebar-header">
                    <h1 class="sidebar-title">👨‍💼 Admin Panel</h1>
                    <button class="sidebar-toggle" @click="toggleSidebar">
                        {{ sidebarOpen ? '◀' : '▶' }}
                    </button>
                </div>
                
                <nav class="sidebar-nav">
                    <button 
                        v-for="item in menuItems" 
                        :key="item.id"
                        class="nav-item"
                        :class="{ active: currentView === item.id }"
                        @click="navigate(item.id)"
                    >
                        <span class="nav-icon">{{ item.icon }}</span>
                        <span v-if="sidebarOpen" class="nav-label">{{ item.label }}</span>
                    </button>
                </nav>
                
                <div class="sidebar-footer">
                    <button class="nav-item" @click="goBack">
                        <span class="nav-icon">🏠</span>
                        <span v-if="sidebarOpen">На сайт</span>
                    </button>
                </div>
            </aside>
            
            <!-- Main Content -->
            <main class="admin-content">
                <header class="admin-header">
                    <h2 class="page-title">
                        {{ menuItems.find(i => i.id === currentView)?.label || 'Admin' }}
                    </h2>
                    <div class="admin-user">
                        <span class="user-name">{{ adminName }}</span>
                        <span class="user-badge">Admin</span>
                    </div>
                </header>
                
                <div class="admin-body">
                    <!-- Dashboard View -->
                    <admin-dashboard v-if="currentView === 'dashboard'"></admin-dashboard>
                    
                    <!-- Users View -->
                    <admin-users v-if="currentView === 'users'"></admin-users>
                    
                    <!-- Questions View -->
                    <admin-questions v-if="currentView === 'questions'"></admin-questions>
                    
                    <!-- Activity View -->
                    <admin-activity v-if="currentView === 'activity'"></admin-activity>
                    
                    <!-- Settings View -->
                    <admin-settings v-if="currentView === 'settings'"></admin-settings>
                </div>
            </main>
        </div>
    `
};

// Экспортируем глобально
window.AdminLayout = AdminLayout;
