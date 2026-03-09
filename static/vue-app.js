// Vue.js 3 Application для Go Quiz

const { createApp, ref, reactive, computed, onMounted } = Vue;

createApp({
    setup() {
        // Состояние приложения
        const currentPage = ref('home');
        const theme = ref('dark');
        const userId = ref('');
        const isLoading = ref(false);

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
        const quizTotal = ref(0);
        const quizAnswered = ref(0);
        const answered = ref(false);
        const selectedOption = ref(-1);

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

        // Инициализация
        onMounted(() => {
            initUserId();
            initTheme();
            loadStats();
        });

        // Инициализация user_id
        function initUserId() {
            let id = localStorage.getItem('goquiz_user_id');
            if (!id) {
                id = 'user_' + Math.random().toString(36).substr(2, 9);
                localStorage.setItem('goquiz_user_id', id);
            }
            userId.value = id;
            document.cookie = `user_id=${id}; path=/; max-age=31536000`;
        }

        // Инициализация темы
        function initTheme() {
            const savedTheme = localStorage.getItem('goquiz_theme');
            if (savedTheme === 'light') {
                theme.value = 'light';
                document.body.classList.add('light-theme');
            }
        }

        // Переключение темы
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

        // Навигация
        function navigate(page) {
            currentPage.value = page;
            if (page === 'stats') loadStats();
            if (page === 'leaderboard') loadLeaderboard();
            if (page === 'skills') loadSkills();
            if (page === 'quests') loadQuests();
            if (page === 'achievements') loadAchievements();
        }

        // API запросы
        async function apiRequest(endpoint, options = {}) {
            const headers = {
                'Content-Type': 'application/json',
                'X-User-ID': userId.value,
                ...options.headers
            };

            const response = await fetch(endpoint, { ...options, headers });
            return response.json();
        }

        // Загрузка статистики
        async function loadStats() {
            try {
                const data = await apiRequest('/api/stats');
                Object.assign(player, data.player);
            } catch (error) {
                console.error('Ошибка загрузки статистики:', error);
            }
        }

        // Викторина
        async function startQuiz() {
            isLoading.value = true;
            try {
                const data = await apiRequest('/api/quiz');
                currentQuestion.value = data.question;
                quizTotal.value = data.total;
                quizAnswered.value = data.answered;
                answered.value = false;
                selectedOption.value = -1;
                navigate('quiz');
            } catch (error) {
                console.error('Ошибка загрузки вопроса:', error);
            }
            isLoading.value = false;
        }

        async function answerQuestion(optionIndex) {
            if (answered.value) return;
            answered.value = true;
            selectedOption.value = optionIndex;

            try {
                const data = await apiRequest('/api/answer', {
                    method: 'POST',
                    body: JSON.stringify({
                        question_id: currentQuestion.value.ID,
                        option_index: optionIndex
                    })
                });

                if (data.level_up) {
                    alert(`🎉 Уровень повышен: ${data.new_level}!`);
                    loadStats();
                }
            } catch (error) {
                console.error('Ошибка отправки ответа:', error);
            }
        }

        function nextQuestion() {
            startQuiz();
        }

        // Навыки
        async function loadSkills() {
            try {
                const data = await apiRequest('/api/skills');
                skillTree.skill_points = data.tree.skill_points;
                skillTree.total_points = data.tree.total_points;
                skillTree.skills = data.tree.skills;
                Object.assign(bonuses, data.bonuses);
            } catch (error) {
                console.error('Ошибка загрузки навыков:', error);
            }
        }

        async function upgradeSkill(skillId) {
            try {
                const data = await apiRequest('/api/skills/upgrade', {
                    method: 'POST',
                    body: JSON.stringify({ skill_id: skillId })
                });
                alert(data.message);
                loadSkills();
                loadStats();
            } catch (error) {
                console.error('Ошибка улучшения навыка:', error);
            }
        }

        // Квесты
        async function loadQuests() {
            try {
                const data = await apiRequest('/api/quests');
                questSystem.quests = data.system.quests;
                questSystem.streak = data.system.streak;
                questSystem.total_completed = data.system.total_completed;
            } catch (error) {
                console.error('Ошибка загрузки квестов:', error);
            }
        }

        async function claimQuest(questId) {
            alert('Награда будет начислена автоматически!');
            loadQuests();
        }

        // Достижения
        async function loadAchievements() {
            try {
                const data = await apiRequest('/api/achievements');
                achievements.system = data.system;
                achievements.unlocked_count = data.unlocked_count;
                achievements.total_count = data.total_count;
            } catch (error) {
                console.error('Ошибка загрузки достижений:', error);
            }
        }

        // Лидерборд
        async function loadLeaderboard() {
            try {
                const data = await apiRequest('/api/leaderboard');
                leaderboard.value = data.entries;
            } catch (error) {
                console.error('Ошибка загрузки лидерборда:', error);
            }
        }

        // Обучение и отдых
        async function studyGo(minutes) {
            try {
                const data = await apiRequest('/api/study', {
                    method: 'POST',
                    body: JSON.stringify({ minutes })
                });
                alert(data.message);
                loadStats();
                loadQuests();
            } catch (error) {
                console.error('Ошибка изучения Go:', error);
            }
        }

        async function rest(minutes) {
            try {
                const data = await apiRequest('/api/rest', {
                    method: 'POST',
                    body: JSON.stringify({ minutes })
                });
                alert(data.message);
                loadStats();
            } catch (error) {
                console.error('Ошибка отдыха:', error);
            }
        }

        async function createBackup() {
            try {
                const data = await apiRequest('/api/backup');
                alert('✅ ' + data.message);
            } catch (error) {
                console.error('Ошибка бэкапа:', error);
            }
        }

        async function resetProgress() {
            if (!confirm('Вы уверены? Весь прогресс будет сброшен!')) return;
            try {
                await apiRequest('/api/reset', { method: 'POST' });
                alert('Прогресс сброшен');
                location.reload();
            } catch (error) {
                console.error('Ошибка сброса:', error);
            }
        }

        // Вычисляемые свойства
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
            return (skill.level / skill.max_level) * 100;
        }

        function getQuestProgress(quest) {
            return (quest.progress / quest.goal) * 100;
        }

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
            skillTree,
            bonuses,
            questSystem,
            achievements,
            leaderboard,
            skillCategories,
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
            getQuestProgress
        };
    }
}).mount('#app');
