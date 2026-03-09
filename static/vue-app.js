// Vue.js 3 Application для Go Quiz с эффектами

const { createApp, ref, reactive, computed, onMounted, watch } = Vue;

// Импортируем эффекты (глобально из vue-effects.js)
const { useConfetti, useParticles, useToast, useLevelUp, useCombo, useShake, useRipple, useFloatingText } = window.VueEffects || {};

// Инициализируем AuthStore
if (typeof AuthStore !== 'undefined') {
    AuthStore.init();
}

createApp({
    setup() {
        // === Состояние приложения ===
        const currentPage = ref('home');
        const theme = ref('dark');
        const userId = ref('');
        const isLoading = ref(false);

        // === Подключаем эффекты ===
        const confetti = useConfetti ? useConfetti() : { confettiParticles: ref([]), isActive: ref(false), createConfetti: () => {} };
        const particles = useParticles ? useParticles() : { particles: ref([]), emitParticles: () => {} };
        const toast = useToast ? useToast() : { toasts: [], success: () => {}, error: () => {}, info: () => {} };
        const levelUp = useLevelUp ? useLevelUp() : { isAnimating: ref(false), level: ref(1), stars: ref([]), triggerLevelUp: () => {} };
        const combo = useCombo ? useCombo() : { combo: ref(0), maxCombo: ref(0), showCombo: ref(false), addCombo: () => {}, resetCombo: () => {} };
        const shake = useShake ? useShake() : { isShaking: ref(false), shakeIntensity: ref(10), shake: () => {} };
        const ripple = useRipple ? useRipple() : { ripples: ref([]), addRipple: () => {} };
        const floatingText = useFloatingText ? useFloatingText() : { texts: ref([]), showFloatingText: () => {} };

        // Данные игрока
        const player = reactive({
            id: '',
            name: '',
            level: 1,
            experience: 0,
            go_knowledge: 0,
            focus: 100,
            willpower: 100,
            money: 0,
            dopamine: 100,
            correct_answers: 0,
            wrong_answers: 0,
            rating: 0
        });

        // Викторина
        const currentQuestion = ref(null);
        const quizTotal = ref(120);
        const quizAnswered = ref(0);
        const answered = ref(false);
        const selectedOption = ref(-1);
        const lastCorrect = ref(false);

        // Навыки
        const skillTree = reactive({
            skill_points: 0,
            total_points: 0,
            skills: {}
        });
        const bonuses = reactive({});

        // Квесты
        const questSystem = reactive({
            quests: [],
            streak: 0,
            total_completed: 0
        });

        // Достижения
        const achievements = reactive({
            system: {},
            unlocked_count: 0,
            total_count: 0
        });

        // Лидерборд
        const leaderboard = ref([]);

        // === Инициализация ===
        onMounted(() => {
            initUserId();
            initTheme();
            loadStats();
        });

        function initUserId() {
            let id = localStorage.getItem('goquiz_user_id');
            if (!id) {
                id = 'user_' + Math.random().toString(36).substr(2, 9);
                localStorage.setItem('goquiz_user_id', id);
            }
            userId.value = id;
            document.cookie = `user_id=${id}; path=/; max-age=31536000`;
        }

        function initTheme() {
            const savedTheme = localStorage.getItem('goquiz_theme');
            if (savedTheme === 'light') {
                theme.value = 'light';
                document.body.classList.add('light-theme');
            }
        }

        function toggleTheme() {
            theme.value = theme.value === 'dark' ? 'light' : 'dark';
            if (theme.value === 'light') {
                document.body.classList.add('light-theme');
                localStorage.setItem('goquiz_theme', 'light');
            } else {
                document.body.classList.remove('light-theme');
                localStorage.setItem('goquiz_theme', 'dark');
            }
        }

        // === Навигация ===
        function navigate(page) {
            currentPage.value = page;
            if (page === 'stats') loadStats();
            if (page === 'leaderboard') loadLeaderboard();
            if (page === 'skills') loadSkills();
            if (page === 'quests') loadQuests();
            if (page === 'achievements') loadAchievements();
        }

        // === API запросы ===
        async function apiRequest(endpoint, options = {}) {
            const headers = {
                'Content-Type': 'application/json',
                'X-User-ID': userId.value,
                ...options.headers
            };

            try {
                const response = await fetch(endpoint, { ...options, headers });
                return await response.json();
            } catch (error) {
                console.error('API Error:', error);
                toast.error('Ошибка соединения с сервером');
                throw error;
            }
        }

        // === Загрузка статистики ===
        async function loadStats() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/stats');
                Object.assign(player, data.player);
            } catch (error) {
                console.error('Ошибка загрузки статистики:', error);
            }
            isLoading.value = false;
        }

        // === Викторина с эффектами ===
        async function startQuiz() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/quiz');
                currentQuestion.value = data.question;
                quizTotal.value = data.total;
                quizAnswered.value = data.answered;
                answered.value = false;
                selectedOption.value = -1;
                lastCorrect.value = false;
                navigate('quiz');
            } catch (error) {
                console.error('Ошибка загрузки вопроса:', error);
            }
            isLoading.value = false;
        }

        async function answerQuestion(optionIndex, event) {
            if (answered.value) return;
            answered.value = true;
            selectedOption.value = optionIndex;

            // Ripple эффект на кнопке
            if (event && ripple.addRipple) {
                ripple.addRipple(event, event.currentTarget);
            }

            try {
                const data = await apiRequest('/api/answer', {
                    method: 'POST',
                    body: JSON.stringify({
                        question_id: currentQuestion.value.ID,
                        option_index: optionIndex
                    })
                });

                lastCorrect.value = data.correct;

                if (data.correct) {
                    // Эффекты для правильного ответа
                    toast.success('✅ Правильно! +' + data.exp + ' EXP');
                    
                    if (particles.emitParticles) {
                        const rect = event?.target?.getBoundingClientRect();
                        const x = rect ? (rect.left + rect.width / 2) : 50;
                        const y = rect ? (rect.top + rect.height / 2) : 50;
                        particles.emitParticles(x, y, 30, '#10b981');
                    }
                    
                    if (floatingText.showFloatingText) {
                        floatingText.showFloatingText('+' + data.exp + ' EXP', 
                            event?.clientX || window.innerWidth / 2, 
                            event?.clientY || window.innerHeight / 2, 
                            '#10b981');
                    }

                    // Combo
                    if (combo.addCombo) {
                        combo.addCombo();
                        if (combo.combo.value >= 3) {
                            toast.info('🔥 Combo x' + combo.combo.value + '!');
                        }
                    }
                } else {
                    // Эффекты для неправильного ответа
                    toast.error('❌ Неправильно!');
                    if (shake.shake) {
                        shake.shake(15, 500);
                    }
                    combo.resetCombo();
                }

                // Обновляем статистику
                player.experience = data.new_exp;
                player.level = data.new_level;

                if (data.level_up) {
                    // LEVEL UP!
                    setTimeout(() => {
                        toast.success('🎉 Уровень повышен: ' + data.new_level + '!');
                        if (levelUp.triggerLevelUp) {
                            levelUp.triggerLevelUp(data.new_level);
                        }
                        if (confetti.createConfetti) {
                            confetti.createConfetti(50, 50);
                        }
                        loadStats();
                    }, 500);
                }

            } catch (error) {
                console.error('Ошибка отправки ответа:', error);
                shake.shake(10, 300);
            }
        }

        function nextQuestion() {
            startQuiz();
        }

        // === Навыки ===
        async function loadSkills() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/skills');
                skillTree.skill_points = data.tree.skill_points;
                skillTree.total_points = data.tree.total_points;
                skillTree.skills = data.tree.skills;
                Object.assign(bonuses, data.bonuses);
            } catch (error) {
                console.error('Ошибка загрузки навыков:', error);
            }
            isLoading.value = false;
        }

        async function upgradeSkill(skillId, event) {
            try {
                const data = await apiRequest('/api/skills/upgrade', {
                    method: 'POST',
                    body: JSON.stringify({ skill_id: skillId })
                });
                
                if (data.message.includes('✅')) {
                    toast.success(data.message);
                    if (confetti.createConfetti) {
                        confetti.createConfetti(50, 30);
                    }
                } else {
                    toast.info(data.message);
                }
                
                loadSkills();
                loadStats();
            } catch (error) {
                console.error('Ошибка улучшения навыка:', error);
                toast.error('Ошибка улучшения навыка');
            }
        }

        // === Квесты ===
        async function loadQuests() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/quests');
                questSystem.quests = data.system.quests;
                questSystem.streak = data.system.streak;
                questSystem.total_completed = data.system.total_completed;
            } catch (error) {
                console.error('Ошибка загрузки квестов:', error);
            }
            isLoading.value = false;
        }

        async function claimQuest(questId) {
            toast.success('🎁 Награда получена!');
            if (confetti.createConfetti) {
                confetti.createConfetti(50, 50);
            }
            loadQuests();
        }

        // === Достижения ===
        async function loadAchievements() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/achievements');
                achievements.system = data.system;
                achievements.unlocked_count = data.unlocked_count;
                achievements.total_count = data.total_count;
            } catch (error) {
                console.error('Ошибка загрузки достижений:', error);
            }
            isLoading.value = false;
        }

        // === Лидерборд ===
        async function loadLeaderboard() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/leaderboard');
                leaderboard.value = data.entries;
            } catch (error) {
                console.error('Ошибка загрузки лидерборда:', error);
            }
            isLoading.value = false;
        }

        // === Обучение и отдых ===
        async function studyGo(minutes) {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/study', {
                    method: 'POST',
                    body: JSON.stringify({ minutes })
                });
                toast.success(data.message);
                if (floatingText.showFloatingText) {
                    floatingText.showFloatingText('+EXP', window.innerWidth / 2, window.innerHeight / 2, '#6366f1');
                }
                loadStats();
                loadQuests();
            } catch (error) {
                console.error('Ошибка изучения Go:', error);
            }
            isLoading.value = false;
        }

        async function rest(minutes) {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/rest', {
                    method: 'POST',
                    body: JSON.stringify({ minutes })
                });
                toast.success(data.message);
                loadStats();
            } catch (error) {
                console.error('Ошибка отдыха:', error);
            }
            isLoading.value = false;
        }

        async function createBackup() {
            try {
                const data = await apiRequest('/api/backup');
                toast.success('✅ ' + data.message);
                if (confetti.createConfetti) {
                    confetti.createConfetti(50, 50);
                }
            } catch (error) {
                console.error('Ошибка бэкапа:', error);
                toast.error('Ошибка создания бэкапа');
            }
        }

        async function resetProgress() {
            if (!confirm('Вы уверены? Весь прогресс будет сброшен!')) return;
            try {
                await apiRequest('/api/reset', { method: 'POST' });
                toast.info('Прогресс сброшен');
                setTimeout(() => location.reload(), 1000);
            } catch (error) {
                console.error('Ошибка сброса:', error);
            }
        }

        // === Вычисляемые свойства ===
        const skillCategories = computed(() => ({
            '📚 GO-НАВЫКИ': ['go_basics', 'concurrency', 'interfaces', 'web_frameworks', 'databases', 'microservices'],
            '🎯 ФОКУС': ['focus_master', 'meditation', 'anti_procrastination'],
            '💪 СИЛА ВОЛИ': ['willpower', 'discipline'],
            '💰 ФИНАНСЫ': ['money_management']
        }));

        const bonusNames = {
            focus: 'Фокус',
            willpower: 'Сила воли',
            knowledge: 'Знание Go',
            money: 'Деньги',
            dopamine: 'Дофамин'
        };

        function getBonusName(type) {
            return bonusNames[type] || type;
        }

        function getSkillProgress(skill) {
            return skill ? (skill.level / skill.max_level) * 100 : 0;
        }

        function getQuestProgress(quest) {
            return quest ? (quest.progress / quest.goal) * 100 : 0;
        }

        function handleRipple(event) {
            ripple.addRipple(event, event.currentTarget);
        }

        // === Возвращаем все данные и методы ===
        return {
            // Состояние
            currentPage,
            theme,
            userId,
            isLoading,
            player,
            currentQuestion,
            quizTotal,
            quizAnswered,
            answered,
            selectedOption,
            lastCorrect,
            skillTree,
            bonuses,
            questSystem,
            achievements,
            leaderboard,
            skillCategories,
            
            // Эффекты
            ...confetti,
            ...particles,
            ...toast,
            ...levelUp,
            ...combo,
            ...shake,
            ...ripple,
            ...floatingText,
            
            // Методы
            toggleTheme,
            navigate,
            startQuiz,
            answerQuestion,
            nextQuestion,
            loadSkills,
            upgradeSkill,
            loadQuests,
            claimQuest,
            loadAchievements,
            loadLeaderboard,
            studyGo,
            rest,
            createBackup,
            resetProgress,
            getBonusName,
            getSkillProgress,
            getQuestProgress,
            handleRipple
        };
    }
}).mount('#app');

// Регистрируем компоненты
if (typeof window !== 'undefined') {
    if (typeof VueGodotGame !== 'undefined') {
        window.app.component('godot-game', VueGodotGame);
    }
    if (typeof LoginComponent !== 'undefined') {
        window.app.component('login-component', LoginComponent);
    }
    if (typeof RegisterComponent !== 'undefined') {
        window.app.component('register-component', RegisterComponent);
    }
    if (typeof ProfileComponent !== 'undefined') {
        window.app.component('profile-component', ProfileComponent);
    }
    if (typeof AdminLayout !== 'undefined') {
        window.app.component('admin-layout', AdminLayout);
    }
    if (typeof AdminDashboard !== 'undefined') {
        window.app.component('admin-dashboard', AdminDashboard);
    }
    if (typeof AdminUsers !== 'undefined') {
        window.app.component('admin-users', AdminUsers);
    }
    if (typeof AdminQuestions !== 'undefined') {
        window.app.component('admin-questions', AdminQuestions);
    }
    if (typeof AdminActivity !== 'undefined') {
        window.app.component('admin-activity', AdminActivity);
    }
    if (typeof AdminSettings !== 'undefined') {
        window.app.component('admin-settings', AdminSettings);
    }
    if (typeof TutorialOverlay !== 'undefined') {
        window.app.component('tutorial-overlay', TutorialOverlay);
    }
    if (typeof TutorialButton !== 'undefined') {
        window.app.component('tutorial-button', TutorialButton);
    }
    if (typeof TutorialHelpButton !== 'undefined') {
        window.app.component('tutorial-help-button', TutorialHelpButton);
    }
}

// Инициализируем TutorialStore
if (typeof TutorialStore !== 'undefined') {
    TutorialStore.init();
}
