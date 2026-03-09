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

    // Воспроизвести по имени
    play(name) {
        const methods = {
            correct: () => this.playCorrect(),
            wrong: () => this.playWrong(),
            levelup: () => this.playLevelUp(),
            coin: () => this.playCoin(),
            combo: () => this.playCombo(1),
            click: () => this.playClick(),
            victory: () => this.playVictory()
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
