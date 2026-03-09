// Vue.js Sound Store для управления звуками в игре

const SoundStore = {
    state: {
        enabled: true,
        volume: 0.7,
        muted: false,
        sounds: {},
        audioContext: null
    },

    // Конфигурация звуков
    soundEffects: {
        correct: { volume: 0.8, preload: true },
        wrong: { volume: 0.8, preload: true },
        levelup: { volume: 1.0, preload: false },
        coin: { volume: 0.7, preload: true },
        combo: { volume: 0.8, preload: false },
        click: { volume: 0.5, preload: true },
        victory: { volume: 1.0, preload: false }
    },

    // Инициализация
    init() {
        const saved = localStorage.getItem('sound_settings');
        if (saved) {
            const s = JSON.parse(saved);
            this.state.enabled = s.enabled ?? true;
            this.state.volume = s.volume ?? 0.7;
            this.state.muted = s.muted ?? false;
        }
        this.initAudioContext();
        console.log('🔊 SoundStore initialized');
    },

    // AudioContext для синтезатора
    initAudioContext() {
        try {
            this.state.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        } catch (e) {
            this.state.audioContext = null;
        }
    },

    // Синтезировать звук (генерация через Web Audio API)
    playSynth(frequency, duration, type = 'sine') {
        if (!this.state.audioContext || !this.state.enabled || this.state.muted) return;
        
        const osc = this.state.audioContext.createOscillator();
        const gain = this.state.audioContext.createGain();
        
        osc.connect(gain);
        gain.connect(this.state.audioContext.destination);
        osc.frequency.value = frequency;
        osc.type = type;
        
        gain.gain.setValueAtTime(this.state.volume * 0.3, this.state.audioContext.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.01, this.state.audioContext.currentTime + duration);
        
        osc.start();
        osc.stop(this.state.audioContext.currentTime + duration);
    },

    // Звук правильного ответа
    playCorrect() {
        this.playSynth(523.25, 0.1);
        setTimeout(() => this.playSynth(659.25, 0.1), 100);
        setTimeout(() => this.playSynth(783.99, 0.2), 200);
    },

    // Звук неправильного ответа
    playWrong() {
        this.playSynth(150, 0.3, 'sawtooth');
        setTimeout(() => this.playSynth(100, 0.3, 'sawtooth'), 150);
    },

    // Звук повышения уровня
    playLevelUp() {
        [523.25, 659.25, 783.99, 1046.50].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.2, 'square'), i * 150);
        });
    },

    // Звук монеты/награды
    playCoin() {
        this.playSynth(987.77, 0.1);
        setTimeout(() => this.playSynth(1318.51, 0.2), 100);
    },

    // Звук комбо
    playCombo(combo) {
        const freq = 440 + (combo * 50);
        this.playSynth(freq, 0.15, 'triangle');
    },

    // Звук клика
    playClick() {
        this.playSynth(800, 0.05, 'sine');
    },

    // Звук победы
    playVictory() {
        [523.25, 659.25, 783.99, 1046.50, 1318.51].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.3, 'square'), i * 200);
        });
    },

    // 🎵 Фоновая музыка (ambient)
    startBackgroundMusic() {
        if (!this.state.enabled || this.state.muted) return;
        
        // Простая амбиент мелодия
        const melody = [
            523.25, 0, 659.25, 0, 783.99, 0, 659.25, 0,
            783.99, 0, 987.77, 0, 1046.50, 0, 987.77, 0
        ];
        
        let noteIndex = 0;
        const playNote = () => {
            if (!this.state.backgroundMusicPlaying) return;
            
            const freq = melody[noteIndex];
            if (freq > 0) {
                this.playSynth(freq * 0.5, 0.3, 'sine'); // Октава ниже
            }
            
            noteIndex = (noteIndex + 1) % melody.length;
            setTimeout(playNote, 400);
        };
        
        this.state.backgroundMusicPlaying = true;
        playNote();
    },

    stopBackgroundMusic() {
        this.state.backgroundMusicPlaying = false;
    },

    // 🔔 Звук достижения (achievement)
    playAchievement() {
        // Победный звук с арпеджио
        [523.25, 659.25, 783.99, 987.77, 1046.50].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.2, 'square'), i * 100);
        });
        setTimeout(() => this.playVictory(), 500);
    },

    // 💫 Звук телепортации (переход между страницами)
    playTeleport() {
        // Скользящий звук вверх
        if (!this.state.audioContext) return;
        
        const osc = this.state.audioContext.createOscillator();
        const gain = this.state.audioContext.createGain();
        
        osc.connect(gain);
        gain.connect(this.state.audioContext.destination);
        
        osc.frequency.setValueAtTime(200, this.state.audioContext.currentTime);
        osc.frequency.exponentialRampToValueAtTime(800, this.state.audioContext.currentTime + 0.3);
        osc.type = 'sine';
        
        gain.gain.setValueAtTime(this.state.volume * 0.2, this.state.audioContext.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.01, this.state.audioContext.currentTime + 0.3);
        
        osc.start();
        osc.stop(this.state.audioContext.currentTime + 0.3);
    },

    // 🎰 Звук вращения (колесо, рандом)
    playSpin() {
        // Быстрые клики
        for (let i = 0; i < 8; i++) {
            setTimeout(() => this.playSynth(600 + (i * 50), 0.05, 'triangle'), i * 80);
        }
    },

    // 🎁 Звук получения награды
    playReward() {
        this.playCoin();
        setTimeout(() => this.playCoin(), 150);
        setTimeout(() => this.playSynth(1046.50, 0.2, 'sine'), 300);
    },

    // ⭐ Звук разблокировки (новый навык, достижение)
    playUnlock() {
        [659.25, 783.99, 987.77, 1046.50].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.15, 'square'), i * 120);
        });
    },

    // 💪 Звук улучшения (апгрейд навыка)
    playUpgrade() {
        this.playSynth(440, 0.1, 'sine');
        setTimeout(() => this.playSynth(554.37, 0.1, 'sine'), 100);
        setTimeout(() => this.playSynth(659.25, 0.2, 'sine'), 200);
    },

    // ⏱️ Звук таймера (время истекает)
    playTimer() {
        this.playSynth(880, 0.1, 'square');
    },

    // 🚫 Звук ошибки (нельзя сделать)
    playError() {
        this.playSynth(150, 0.2, 'sawtooth');
        setTimeout(() => this.playSynth(100, 0.2, 'sawtooth'), 150);
    },

    // ✨ Звук магии (special effect)
    playMagic() {
        [783.99, 987.77, 1174.66, 1318.51].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.15, 'sine'), i * 80);
        });
    },

    // 🎯 Звук фокуса (концентрация)
    playFocus() {
        this.playSynth(392, 0.3, 'sine'); // G4
    },

    // 💀 Звук поражения (босс, игра окончена)
    playDefeat() {
        [392, 349.23, 329.63, 261.63].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.4, 'sawtooth'), i * 300);
        });
    },

    // 🎪 Звук мини-игры
    playMiniGame() {
        [523.25, 659.25, 783.99, 659.25].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.1, 'square'), i * 150);
        });
    },

    // 📊 Звук статистики (открытие)
    playStats() {
        this.playSynth(523.25, 0.15, 'sine');
        setTimeout(() => this.playSynth(659.25, 0.15, 'sine'), 150);
    },

    // 🏆 Звук рекорда
    playRecord() {
        this.playVictory();
        setTimeout(() => this.playAchievement(), 800);
    },

    // 🎂 Звук дня рождения / праздника
    playCelebration() {
        [523.25, 523.25, 587.33, 523.25, 659.25, 659.25].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.2, 'square'), i * 200);
        });
    },

    // 🔔 Звук уведомления
    playNotification() {
        this.playSynth(1046.50, 0.1, 'sine');
        setTimeout(() => this.playSynth(1318.51, 0.2, 'sine'), 100);
    },

    // 💤 Звук отдыха / сна
    playRest() {
        [329.63, 392, 493.88].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.3, 'sine'), i * 300);
        });
    },

    // 📚 Звук обучения
    playStudy() {
        [440, 493.88, 523.25, 587.33].forEach((f, i) => {
            setTimeout(() => this.playSynth(f, 0.15, 'sine'), i * 150);
        });
    },

    // 🎮 Общий play по имени
    play(name, options = {}) {
        const methods = {
            correct: () => this.playCorrect(),
            wrong: () => this.playWrong(),
            levelup: () => this.playLevelUp(),
            coin: () => this.playCoin(),
            combo: () => this.playCombo(options.combo || 1),
            click: () => this.playClick(),
            victory: () => this.playVictory(),
            achievement: () => this.playAchievement(),
            teleport: () => this.playTeleport(),
            spin: () => this.playSpin(),
            reward: () => this.playReward(),
            unlock: () => this.playUnlock(),
            upgrade: () => this.playUpgrade(),
            timer: () => this.playTimer(),
            error: () => this.playError(),
            magic: () => this.playMagic(),
            focus: () => this.playFocus(),
            defeat: () => this.playDefeat(),
            miniGame: () => this.playMiniGame(),
            stats: () => this.playStats(),
            record: () => this.playRecord(),
            celebration: () => this.playCelebration(),
            notification: () => this.playNotification(),
            rest: () => this.playRest(),
            study: () => this.playStudy()
        };
        
        if (methods[name] && this.state.enabled && !this.state.muted) {
            methods[name]();
            return true;
        }
        return false;
    },

    // Настройки
    toggle() {
        this.state.enabled = !this.state.enabled;
        this.saveSettings();
        return this.state.enabled;
    },

    toggleMute() {
        this.state.muted = !this.state.muted;
        this.saveSettings();
        return this.state.muted;
    },

    setVolume(vol) {
        this.state.volume = Math.max(0, Math.min(1, vol));
        this.saveSettings();
    },

    saveSettings() {
        localStorage.setItem('sound_settings', JSON.stringify({
            enabled: this.state.enabled,
            volume: this.state.volume,
            muted: this.state.muted
        }));
    },

    getSettings() {
        return {
            enabled: this.state.enabled,
            volume: this.state.volume,
            muted: this.state.muted
        };
    }
};

window.SoundStore = SoundStore;
