/**
 * Godot ↔ JavaScript ↔ Vue.js Bridge
 * Связь между Godot игрой и веб-приложением
 */

class GodotBridge {
    constructor() {
        this.godotInstance = null;
        this.userId = localStorage.getItem('goquiz_user_id') || '';
        this.callbacks = {};
        
        this.init();
    }

    init() {
        console.log('🌉 GodotBridge initialized');
        
        // Регистрируем глобальные функции для Godot
        window.godotBridge = {
            onGameReady: this.onGameReady.bind(this),
            onQuestionAnswered: this.onQuestionAnswered.bind(this),
            onLevelUp: this.onLevelUp.bind(this),
            onKnowledgeCollected: this.onKnowledgeCollected.bind(this),
            onComboChanged: this.onComboChanged.bind(this),
            onPlayerDataLoaded: this.onPlayerDataLoaded.bind(this)
        };

        // Загружаем данные игрока
        this.loadPlayerData();
    }

    /**
     * Инициализация Godot игры
     */
    async initGodot(canvasId = 'godot-canvas') {
        if (!this.godotInstance) {
            console.log('🎮 Initializing Godot...');
            
            // Godot должен быть экспортирован как модуль
            if (typeof loadGodotGame === 'function') {
                this.godotInstance = await loadGodotGame(canvasId);
            } else {
                console.warn('⚠️ Godot export not found, using fallback');
                this.useFallbackMode();
            }
        }
        
        return this.godotInstance;
    }

    /**
     * Fallback режим (если Godot не загружен)
     */
    useFallbackMode() {
        console.log('🕹️ Using fallback canvas mode');
        this.setupFallbackCanvas();
    }

    setupFallbackCanvas() {
        // Создаём простой canvas для демонстрации
        const canvas = document.getElementById('godot-canvas');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        let playerX = 400;
        let playerY = 500;
        let orbs = [];

        // Генерируем орбы
        for (let i = 0; i < 5; i++) {
            orbs.push({
                x: 100 + i * 150,
                y: 200 + Math.random() * 100,
                collected: false
            });
        }

        // Игровой цикл
        const gameLoop = () => {
            ctx.fillStyle = '#1a1a2e';
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            // Рисуем орбы
            orbs.forEach(orb => {
                if (!orb.collected) {
                    ctx.beginPath();
                    ctx.arc(orb.x, orb.y + Math.sin(Date.now() / 500) * 10, 20, 0, Math.PI * 2);
                    ctx.fillStyle = '#6366f1';
                    ctx.fill();
                    ctx.strokeStyle = '#818cf8';
                    ctx.stroke();
                    
                    // Свечение
                    ctx.shadowColor = '#6366f1';
                    ctx.shadowBlur = 20;
                }
            });

            // Игрок
            ctx.beginPath();
            ctx.arc(playerX, playerY, 30, 0, Math.PI * 2);
            ctx.fillStyle = '#10b981';
            ctx.fill();

            // Управление
            requestAnimationFrame(gameLoop);
        };

        gameLoop();

        // Обработка клавиатуры
        document.addEventListener('keydown', (e) => {
            if (e.key === 'a' || e.key === 'ArrowLeft') playerX -= 20;
            if (e.key === 'd' || e.key === 'ArrowRight') playerX += 20;
            
            // Проверка сбора орбов
            orbs.forEach(orb => {
                if (!orb.collected && Math.abs(playerX - orb.x) < 50 && Math.abs(playerY - orb.y) < 50) {
                    orb.collected = true;
                    this.onKnowledgeCollected(10);
                }
            });
        });
    }

    /**
     * Godot вызываем эту функцию при готовности
     */
    onGameReady(data) {
        console.log('🎮 Godot game ready:', data);
        this.emit('gameReady', data);
        
        // Передаём данные игрока в Godot
        this.sendPlayerDataToGodot();
    }

    /**
     * Ответ на вопрос получен
     */
    onQuestionAnswered(exp, correct) {
        console.log('📝 Question answered:', exp, correct);
        
        // Обновляем статистику в Vue
        this.emit('questionAnswered', { exp, correct });
        
        // Показывем toast
        if (correct) {
            this.showToast(`✅ Правильно! +${exp} EXP`, 'success');
        } else {
            this.showToast('❌ Неправильно!', 'error');
        }
    }

    /**
     * Повышение уровня
     */
    onLevelUp(newLevel) {
        console.log('🎉 LEVEL UP!', newLevel);
        
        this.emit('levelUp', { level: newLevel });
        this.showToast(`🎉 Уровень повышен: ${newLevel}!`, 'success');
        
        // Конфетти
        this.triggerConfetti();
    }

    /**
     * Сбор знания
     */
    onKnowledgeCollected(amount) {
        console.log('📚 Knowledge collected:', amount);
        this.emit('knowledgeCollected', { amount });
    }

    /**
     * Изменение комбо
     */
    onComboChanged(combo) {
        console.log('🔥 Combo:', combo);
        this.emit('comboChanged', { combo });
    }

    /**
     * Данные игрока загружены
     */
    onPlayerDataLoaded(data) {
        console.log('📊 Player data loaded:', data);
        this.emit('playerDataLoaded', data);
    }

    /**
     * Загрузка данных игрока
     */
    async loadPlayerData() {
        try {
            const response = await fetch('/api/stats', {
                headers: { 'X-User-ID': this.userId }
            });
            const data = await response.json();
            
            this.onPlayerDataLoaded(data.player);
            
            // Если Godot загружен, передаём данные
            if (this.godotInstance) {
                this.sendPlayerDataToGodot();
            }
            
            return data;
        } catch (error) {
            console.error('Failed to load player data:', error);
        }
    }

    /**
     * Отправка данных игрока в Godot
     */
    sendPlayerDataToGodot() {
        if (this.godotInstance && this.godotInstance.receivePlayerData) {
            this.godotInstance.receivePlayerData({
                userId: this.userId,
                // ... данные игрока
            });
        }
    }

    /**
     * Показать toast уведомление
     */
    showToast(message, type = 'info') {
        const event = new CustomEvent('showToast', { detail: { message, type } });
        window.dispatchEvent(event);
    }

    /**
     * Запуск конфетти
     */
    triggerConfetti() {
        const event = new CustomEvent('triggerConfetti');
        window.dispatchEvent(event);
    }

    /**
     * Подписка на события
     */
    on(event, callback) {
        if (!this.callbacks[event]) {
            this.callbacks[event] = [];
        }
        this.callbacks[event].push(callback);
    }

    /**
     * Отписка от событий
     */
    off(event, callback) {
        if (this.callbacks[event]) {
            this.callbacks[event] = this.callbacks[event].filter(cb => cb !== callback);
        }
    }

    /**
     * Эмиссия события
     */
    emit(event, data) {
        if (this.callbacks[event]) {
            this.callbacks[event].forEach(cb => cb(data));
        }
    }
}

// Создаём глобальный экземпляр
window.godotBridge = new GodotBridge();

// Экспорт для модулей
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GodotBridge;
}
