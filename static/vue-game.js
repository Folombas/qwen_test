// Vue.js компонент Godot Game Page

const VueGodotGame = {
    name: 'GodotGame',
    setup() {
        const { ref, reactive, onMounted, onUnmounted } = Vue;

        // Состояние
        const gameLoaded = ref(false);
        const gameData = reactive({
            player: null,
            questions: [],
            orbs: []
        });
        const godotInstance = ref(null);
        const canvasRef = ref(null);

        // Инициализация
        onMounted(() => {
            initGame();
        });

        onUnmounted(() => {
            cleanupGame();
        });

        // Инициализация игры
        async function initGame() {
            try {
                // Загружаем данные игры
                await loadGameData();
                
                // Инициализируем Godot bridge
                if (window.godotBridge) {
                    window.godotBridge.on('gameReady', onGameReady);
                    window.godotBridge.on('questionAnswered', onQuestionAnswered);
                    window.godotBridge.on('levelUp', onLevelUp);
                    
                    // Инициализируем Godot canvas
                    await initGodotCanvas();
                }
                
                gameLoaded.value = true;
            } catch (error) {
                console.error('Failed to init game:', error);
            }
        }

        // Загрузка данных игры
        async function loadGameData() {
            const userId = localStorage.getItem('goquiz_user_id');
            const response = await fetch('/api/game', {
                headers: { 'X-User-ID': userId }
            });
            const data = await response.json();
            
            gameData.player = data.player;
            gameData.questions = data.questions || [];
        }

        // Инициализация Godot canvas
        async function initGodotCanvas() {
            if (!canvasRef.value) return;

            // Если Godot экспортирован, загружаем его
            if (typeof loadGodotGame === 'function') {
                godotInstance.value = await loadGodotGame(canvasRef.value);
            } else {
                // Fallback режим уже инициализирован в godot-bridge.js
                console.log('🕹️ Using fallback canvas mode');
            }
        }

        // Очистка
        function cleanupGame() {
            if (window.godotBridge) {
                window.godotBridge.off('gameReady', onGameReady);
                window.godotBridge.off('questionAnswered', onQuestionAnswered);
                window.godotBridge.off('levelUp', onLevelUp);
            }
        }

        // Обработчики событий от Godot
        function onGameReady(data) {
            console.log('🎮 Godot game ready:', data);
        }

        function onQuestionAnswered(data) {
            console.log('📝 Question answered:', data);
            // Обновляем статистику
            loadGameData();
        }

        function onLevelUp(data) {
            console.log('🎉 Level up!', data);
            // Показываем уведомление
            showToast('🎉 Уровень повышен: ' + data.level + '!', 'success');
        }

        // Утилита для toast
        function showToast(message, type = 'info') {
            const event = new CustomEvent('showToast', { detail: { message, type } });
            window.dispatchEvent(event);
        }

        // Методы для UI
        function startGame() {
            if (godotInstance.value) {
                godotInstance.value.start();
            }
        }

        function pauseGame() {
            if (godotInstance.value) {
                godotInstance.value.pause();
            }
        }

        function restartGame() {
            location.reload();
        }

        return {
            gameLoaded,
            gameData,
            canvasRef,
            startGame,
            pauseGame,
            restartGame
        };
    },
    template: `
        <div class="godot-game-page">
            <div class="game-header">
                <h2>🎮 Go Quiz Game</h2>
                <div class="game-controls">
                    <button @click="startGame" class="btn-game">▶️ Старт</button>
                    <button @click="pauseGame" class="btn-game">⏸️ Пауза</button>
                    <button @click="restartGame" class="btn-game btn-reset">🔄 Заново</button>
                </div>
            </div>

            <div class="game-container" v-if="gameLoaded">
                <!-- Godot Canvas -->
                <div class="canvas-wrapper">
                    <canvas ref="canvasRef" id="godot-canvas" width="800" height="600"></canvas>
                    
                    <!-- Overlay UI -->
                    <div class="game-overlay">
                        <div class="player-stats" v-if="gameData.player">
                            <span>🏆 Ур. {{ gameData.player.level }}</span>
                            <span>⚡ EXP: {{ gameData.player.experience }}</span>
                            <span>📚 Знание: {{ gameData.player.go_knowledge }}/100</span>
                        </div>
                        
                        <div class="controls-hint">
                            <span>🎮 A/D - Движение</span>
                            <span>⬆️ Space - Прыжок</span>
                            <span>📦 E - Взаимодействие</span>
                        </div>
                    </div>
                </div>

                <!-- Info Panel -->
                <div class="info-panel">
                    <h3>📖 Как играть</h3>
                    <ul>
                        <li>🟣 Собирай орбы с вопросами</li>
                        <li>✅ Отвечай правильно для получения EXP</li>
                        <li>🔥 Поддерживай комбо для бонусов</li>
                        <li>🎯 Избегай ошибок для сохранения комбо</li>
                        <li>📈 Повышай уровень и открывай навыки</li>
                    </ul>
                    
                    <div class="lua-info">
                        <h4>🔧 Lua Логика</h4>
                        <p>Игровая логика работает на Lua (lua/game_logic.lua)</p>
                        <p>Godot использует GDScript для рендеринга</p>
                    </div>
                </div>
            </div>

            <div class="loading-game" v-else>
                <div class="spinner"></div>
                <p>Загрузка Godot игры...</p>
            </div>
        </div>
    `
};

// Экспортируем глобально
window.VueGodotGame = VueGodotGame;