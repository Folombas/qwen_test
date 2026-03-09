// Vue.js Tutorial Store для обучающего гайда

const TutorialStore = {
    state: {
        isActive: false,
        currentStep: 0,
        completedSteps: [],
        totalSteps: 0,
        tutorialType: 'main' // main, quiz, skills, quests
    },

    // Шаги основного обучения
    mainSteps: [
        {
            id: 'welcome',
            title: '🎮 Добро пожаловать в Go Quiz!',
            content: 'Привет! Я помогу тебе разобраться в игре. Давай пройдёмся по основным функциям!',
            target: null,
            position: 'center'
        },
        {
            id: 'navigation',
            title: '🧭 Навигация',
            content: 'Вверху страницы ты видишь кнопки навигации. С их помощью можно перемещаться между разделами игры.',
            target: '.header-actions',
            position: 'bottom'
        },
        {
            id: 'quiz',
            title: '🎯 Викторина',
            content: 'Здесь ты будешь отвечать на вопросы по Go. За правильные ответы получаешь опыт (EXP) и повышаешь уровень!',
            target: '.nav-btn:nth-child(2)',
            position: 'bottom'
        },
        {
            id: 'study',
            title: '📚 Обучение',
            content: 'В этом разделе можно изучать Go и отдыхать. Это восстанавливает фокус и даёт дополнительные очки!',
            target: '.nav-btn:nth-child(3)',
            position: 'bottom'
        },
        {
            id: 'skills',
            title: '🌳 Навыки',
            content: 'Дерево навыков позволяет улучшать характеристики. Получай очки навыков за уровни и улучшай их!',
            target: '.nav-btn:nth-child(4)',
            position: 'bottom'
        },
        {
            id: 'quests',
            title: '📋 Квесты',
            content: 'Ежедневные задания дадут тебе дополнительные очки навыков. Выполняй 5 квестов каждый день!',
            target: '.nav-btn:nth-child(5)',
            position: 'bottom'
        },
        {
            id: 'achievements',
            title: '🏆 Достижения',
            content: 'Коллекционируй достижения! Здесь 23 награды за уровни, квесты, серии дней и другие успехи.',
            target: '.nav-btn:nth-child(6)',
            position: 'bottom'
        },
        {
            id: 'stats',
            title: '📊 Статистика',
            content: 'Здесь отображается вся твоя статистика: уровень, опыт, правильные ответы и рейтинг.',
            target: '.nav-btn:nth-child(7)',
            position: 'bottom'
        },
        {
            id: 'leaderboard',
            title: '👑 Таблица лидеров',
            content: 'Соревнуйся с другими игроками! Твой рейтинг зависит от уровня, знаний Go, фокуса и силы воли.',
            target: '.nav-btn:nth-child(8)',
            position: 'bottom'
        },
        {
            id: 'profile',
            title: '👤 Профиль',
            content: 'Здесь можно посмотреть свой профиль, сменить пароль и настройки.',
            target: '.profile-btn',
            position: 'left'
        },
        {
            id: 'complete',
            title: '🎉 Обучение завершено!',
            content: 'Молодец! Теперь ты готов к игре. Удачи в изучении Go! 🚀',
            target: null,
            position: 'center'
        }
    ],

    // Шаги обучения викторине
    quizSteps: [
        {
            id: 'quiz-welcome',
            title: '🎯 Викторина',
            content: 'Добро пожаловать в викторину! Здесь ты будешь отвечать на вопросы по Go.',
            target: '.quiz-container',
            position: 'center'
        },
        {
            id: 'question',
            title: '❓ Вопрос',
            content: 'Читай вопрос внимательно. Выбери один правильный ответ из четырёх вариантов.',
            target: '.question-text',
            position: 'bottom'
        },
        {
            id: 'options',
            title: '💡 Варианты ответов',
            content: 'Нажми на вариант ответа. Правильный ответ подсветится зелёным, неправильный — красным.',
            target: '.options',
            position: 'right'
        },
        {
            id: 'exp',
            title: '⚡ Опыт (EXP)',
            content: 'За правильный ответ ты получаешь опыт. Количество EXP зависит от сложности вопроса.',
            target: '.exp-badge',
            position: 'top'
        },
        {
            id: 'next',
            title: '➡️ Следующий вопрос',
            content: 'После ответа появится кнопка "Далее". Нажми её чтобы продолжить!',
            target: '.next-btn',
            position: 'top'
        },
        {
            id: 'quiz-complete',
            title: '🎉 Готово!',
            content: 'Теперь ты знаешь как работает викторина. Отвечай на вопросы и получай опыт!',
            target: null,
            position: 'center'
        }
    ],

    // Шаги обучения навыкам
    skillsSteps: [
        {
            id: 'skills-welcome',
            title: '🌳 Дерево навыков',
            content: 'Добро пожаловать в дерево навыков! Здесь ты можешь улучшать свои характеристики.',
            target: '.skills-container',
            position: 'center'
        },
        {
            id: 'skill-points',
            title: '✨ Очки навыков',
            content: 'Очки навыков даются за каждый уровень. 2 + (уровень / 5) очков за уровень.',
            target: '.skill-points-display',
            position: 'bottom'
        },
        {
            id: 'categories',
            title: '📁 Категории',
            content: 'Навыки разделены на 4 категории: GO-НАВЫКИ, ФОКУС, СИЛА ВОЛИ, ФИНАНСЫ.',
            target: '.skill-category:first-child',
            position: 'right'
        },
        {
            id: 'skill-info',
            title: 'ℹ️ Информация о навыке',
            content: 'Каждый навык имеет уровень, стоимость улучшения и бонус к характеристике.',
            target: '.skill-item:first-child',
            position: 'right'
        },
        {
            id: 'upgrade',
            title: '⬆️ Улучшение',
            content: 'Нажми "Улучшить" чтобы повысить уровень навыка. Это даст бонус к характеристике!',
            target: '.upgrade-btn:first-child',
            position: 'top'
        },
        {
            id: 'bonuses',
            title: '🎁 Бонусы',
            content: 'Бонусы от навыков применяются автоматически. Следи за своими характеристиками!',
            target: '.skill-bonus',
            position: 'top'
        },
        {
            id: 'skills-complete',
            title: '🎉 Готово!',
            content: 'Теперь ты знаешь как работать с навыками. Улучшай их strategically!',
            target: null,
            position: 'center'
        }
    ],

    // Шаги обучения квестам
    questsSteps: [
        {
            id: 'quests-welcome',
            title: '📋 Ежедневные квесты',
            content: 'Добро пожаловать в раздел квестов! Здесь ты будешь выполнять ежедневные задания.',
            target: '.quests-container',
            position: 'center'
        },
        {
            id: 'quest-list',
            title: '📝 Список квестов',
            content: 'Каждый день генерируется 5 новых квестов. Выполни их все!',
            target: '.quest-item',
            position: 'right'
        },
        {
            id: 'quest-progress',
            title: '📊 Прогресс',
            content: 'Следи за прогрессом выполнения. Полоска показывает сколько осталось сделать.',
            target: '.quest-progress-bar',
            position: 'top'
        },
        {
            id: 'quest-reward',
            title: '🎁 Награда',
            content: 'После выполнения квеста нажми "Забрать" чтобы получить очки навыков!',
            target: '.claim-btn',
            position: 'top'
        },
        {
            id: 'quest-streak',
            title: '🔥 Серия дней',
            content: 'Выполняй все квесты каждый день чтобы поддерживать серию! Это даёт бонусы.',
            target: '.quest-stats',
            position: 'top'
        },
        {
            id: 'quests-complete',
            title: '🎉 Готово!',
            content: 'Теперь ты знаешь как работать с квестами. Выполняй их ежедневно!',
            target: null,
            position: 'center'
        }
    ],

    // Инициализация
    init() {
        // Загружаем прогресс из localStorage
        const saved = localStorage.getItem('tutorial_completed');
        if (saved) {
            this.state.completedSteps = JSON.parse(saved);
        }
    },

    // Начать обучение
    startTutorial(type = 'main') {
        this.state.tutorialType = type;
        this.state.currentStep = 0;
        this.state.isActive = true;
        
        const steps = this.getSteps(type);
        this.state.totalSteps = steps.length;
        
        // Показываем первый шаг
        this.showStep(0);
    },

    // Получить шаги
    getSteps(type) {
        switch(type) {
            case 'quiz': return this.quizSteps;
            case 'skills': return this.skillsSteps;
            case 'quests': return this.questSteps;
            default: return this.mainSteps;
        }
    },

    // Показать шаг
    showStep(index) {
        const steps = this.getSteps(this.state.tutorialType);
        if (index >= 0 && index < steps.length) {
            this.state.currentStep = index;
            
            // Подсвечиваем элемент
            const step = steps[index];
            if (step.target) {
                this.highlightElement(step.target);
            }
            
            // Событие для компонента
            document.dispatchEvent(new CustomEvent('tutorial-step', { 
                detail: { step, index, total: steps.length }
            }));
        }
    },

    // Следующий шаг
    nextStep() {
        const steps = this.getSteps(this.state.tutorialType);
        if (this.state.currentStep < steps.length - 1) {
            this.showStep(this.state.currentStep + 1);
        } else {
            this.completeTutorial();
        }
    },

    // Предыдущий шаг
    prevStep() {
        if (this.state.currentStep > 0) {
            this.showStep(this.state.currentStep - 1);
        }
    },

    // Завершить обучение
    completeTutorial() {
        this.state.isActive = false;
        this.state.completedSteps.push(this.state.tutorialType);
        localStorage.setItem('tutorial_completed', JSON.stringify(this.state.completedSteps));
        
        document.dispatchEvent(new CustomEvent('tutorial-complete', {
            detail: { type: this.state.tutorialType }
        }));
    },

    // Закрыть обучение
    closeTutorial() {
        this.state.isActive = false;
        this.clearHighlight();
    },

    // Подсветка элемента
    highlightElement(selector) {
        this.clearHighlight();
        
        const element = document.querySelector(selector);
        if (element) {
            element.classList.add('tutorial-highlight');
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    },

    // Очистить подсветку
    clearHighlight() {
        document.querySelectorAll('.tutorial-highlight').forEach(el => {
            el.classList.remove('tutorial-highlight');
        });
    },

    // Пропущено ли обучение
    isTutorialCompleted(type) {
        return this.state.completedSteps.includes(type);
    },

    // Сбросить прогресс
    resetProgress() {
        this.state.completedSteps = [];
        localStorage.removeItem('tutorial_completed');
    },

    // Получить прогресс
    getProgress() {
        const steps = this.getSteps(this.state.tutorialType);
        return {
            current: this.state.currentStep + 1,
            total: steps.length,
            percent: Math.round((this.state.currentStep + 1) / steps.length * 100)
        };
    }
};

// Экспортируем глобально
window.TutorialStore = TutorialStore;
