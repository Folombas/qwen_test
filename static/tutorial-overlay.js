// Vue.js Tutorial Overlay компонент

const TutorialOverlay = {
    name: 'TutorialOverlay',
    setup() {
        const { ref, reactive, onMounted, onUnmounted } = Vue;
        
        const isVisible = ref(false);
        const currentStep = ref(null);
        const stepIndex = ref(0);
        const totalSteps = ref(0);
        const position = ref('center');
        
        onMounted(() => {
            // Слушаем события от TutorialStore
            document.addEventListener('tutorial-step', onStepChange);
            document.addEventListener('tutorial-complete', onComplete);
        });
        
        onUnmounted(() => {
            document.removeEventListener('tutorial-step', onStepChange);
            document.removeEventListener('tutorial-complete', onComplete);
        });
        
        function onStepChange(event) {
            currentStep.value = event.detail.step;
            stepIndex.value = event.detail.index;
            totalSteps.value = event.detail.total;
            position.value = event.detail.step.position || 'center';
            isVisible.value = true;
        }
        
        function onComplete() {
            isVisible.value = false;
        }
        
        function next() {
            TutorialStore.nextStep();
        }
        
        function prev() {
            TutorialStore.prevStep();
        }
        
        function close() {
            TutorialStore.closeTutorial();
            isVisible.value = false;
        }
        
        function getProgress() {
            return ((stepIndex.value + 1) / totalSteps.value) * 100;
        }
        
        return {
            isVisible,
            currentStep,
            stepIndex,
            totalSteps,
            position,
            next,
            prev,
            close,
            getProgress
        };
    },
    template: `
        <transition name="tutorial-fade">
            <div v-if="isVisible" class="tutorial-overlay">
                <!-- Затемнение фона -->
                <div class="tutorial-backdrop"></div>
                
                <!-- Туториал карточка -->
                <div class="tutorial-card" :class="'position-' + position">
                    <!-- Заголовок -->
                    <div class="tutorial-header">
                        <h3 class="tutorial-title">{{ currentStep?.title || 'Обучение' }}</h3>
                        <button class="tutorial-close" @click="close">×</button>
                    </div>
                    
                    <!-- Контент -->
                    <div class="tutorial-content">
                        <p class="tutorial-text">{{ currentStep?.content }}</p>
                    </div>
                    
                    <!-- Прогресс бар -->
                    <div class="tutorial-progress">
                        <div class="progress-bar">
                            <div class="progress-fill" :style="{ width: getProgress() + '%' }"></div>
                        </div>
                        <span class="progress-text">{{ stepIndex + 1 }} / {{ totalSteps }}</span>
                    </div>
                    
                    <!-- Кнопки навигации -->
                    <div class="tutorial-footer">
                        <button 
                            v-if="stepIndex > 0" 
                            class="tutorial-btn tutorial-btn-prev"
                            @click="prev"
                        >
                            ← Назад
                        </button>
                        <button v-else class="tutorial-btn tutorial-btn-spacer"></button>
                        
                        <button 
                            v-if="stepIndex < totalSteps - 1"
                            class="tutorial-btn tutorial-btn-next"
                            @click="next"
                        >
                            Далее →
                        </button>
                        <button 
                            v-else
                            class="tutorial-btn tutorial-btn-complete"
                            @click="next"
                        >
                            🎉 Завершить
                        </button>
                        
                        <button class="tutorial-btn tutorial-btn-skip" @click="close">
                            Пропустить
                        </button>
                    </div>
                </div>
            </div>
        </transition>
    `
};

// Экспортируем глобально
window.TutorialOverlay = TutorialOverlay;
