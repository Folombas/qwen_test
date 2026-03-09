// Vue.js Sound Settings компонент

const SoundSettings = {
    name: 'SoundSettings',
    setup() {
        const { ref, reactive, onMounted } = Vue;
        
        const settings = reactive({
            enabled: true,
            volume: 0.7,
            muted: false
        });
        
        const showPopup = ref(false);
        
        onMounted(() => {
            loadSettings();
        });
        
        function loadSettings() {
            const s = SoundStore.getSettings();
            settings.enabled = s.enabled;
            settings.volume = s.volume;
            settings.muted = s.muted;
        }
        
        function toggle() {
            SoundStore.toggle();
            loadSettings();
        }
        
        function toggleMute() {
            SoundStore.toggleMute();
            loadSettings();
        }
        
        function setVolume(val) {
            SoundStore.setVolume(val);
            loadSettings();
        }
        
        function testSound() {
            SoundStore.playCoin();
        }
        
        return {
            settings,
            showPopup,
            toggle,
            toggleMute,
            setVolume,
            testSound
        };
    },
    template: `
        <div class="sound-settings">
            <button class="sound-btn" @click="showPopup = !showPopup" title="Настройки звука">
                🔊
            </button>
            
            <transition name="fade">
                <div v-if="showPopup" class="sound-popup" @click.self="showPopup = false">
                    <div class="sound-card">
                        <div class="sound-header">
                            <h4>🔊 Звук</h4>
                            <button class="close" @click="showPopup = false">×</button>
                        </div>
                        
                        <div class="sound-control">
                            <button @click="toggle" class="toggle-btn" :class="{ active: settings.enabled }">
                                {{ settings.enabled ? '🔊 Вкл' : '🔇 Выкл' }}
                            </button>
                        </div>
                        
                        <div class="sound-control">
                            <button @click="toggleMute" class="mute-btn" :class="{ active: settings.muted }">
                                {{ settings.muted ? '🔇 Unmute' : '🔈 Mute' }}
                            </button>
                        </div>
                        
                        <div class="sound-control">
                            <label>Громкость: {{ Math.round(settings.volume * 100) }}%</label>
                            <input 
                                type="range" 
                                min="0" 
                                max="100" 
                                :value="settings.volume * 100"
                                @input="setVolume($event.target.value / 100)"
                                class="volume-slider"
                            />
                        </div>
                        
                        <button @click="testSound" class="test-btn">
                            🔔 Тест звука
                        </button>
                    </div>
                </div>
            </transition>
        </div>
    `
};

window.SoundSettings = SoundSettings;
