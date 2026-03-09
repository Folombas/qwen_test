// Vue.js Tutorial Button компонент

const TutorialButton = {
    name: 'TutorialButton',
    props: {
        type: {
            type: String,
            default: 'main'
        },
        label: {
            type: String,
            default: '📚 Обучение'
        }
    },
    setup(props) {
        const { ref, onMounted } = Vue;
        
        const isCompleted = ref(false);
        const showPrompt = ref(false);
        
        onMounted(() => {
            isCompleted.value = TutorialStore.isTutorialCompleted(props.type);
            
            // Показываем подсказку для новых пользователей
            if (!isCompleted.value && props.type === 'main') {
                setTimeout(() => {
                    showPrompt.value = true;
                }, 3000);
            }
        });
        
        function startTutorial() {
            TutorialStore.startTutorial(props.type);
            showPrompt.value = false;
        }
        
        function dismissPrompt() {
            showPrompt.value = false;
        }
        
        return {
            isCompleted,
            showPrompt,
            startTutorial,
            dismissPrompt
        };
    },
    template: `
        <div class="tutorial-button-wrapper">
            <button 
                class="tutorial-btn-trigger"
                :class="{ completed: isCompleted }"
                @click="startTutorial"
            >
                <span class="btn-icon">{{ isCompleted ? '✅' : '📚' }}</span>
                <span class="btn-label">{{ label }}</span>
            </button>
            
            <!-- Подсказка для новых пользователей -->
            <transition name="prompt-fade">
                <div v-if="showPrompt" class="tutorial-prompt">
                    <div class="prompt-arrow"></div>
                    <div class="prompt-content">
                        <strong>👋 Новичок?</strong>
                        <p>Пройди обучение чтобы узнать как играть!</p>
                    </div>
                </div>
            </transition>
        </div>
    `
};

// Helper компонент для кнопки в хедере
const TutorialHelpButton = {
    name: 'TutorialHelpButton',
    setup() {
        function startMainTutorial() {
            TutorialStore.startTutorial('main');
        }
        
        return {
            startMainTutorial
        };
    },
    template: `
        <button class="tutorial-help-btn" @click="startMainTutorial" title="Пройти обучение">
            ❓
        </button>
    `
};

// Экспортируем глобально
window.TutorialButton = TutorialButton;
window.TutorialHelpButton = TutorialHelpButton;
