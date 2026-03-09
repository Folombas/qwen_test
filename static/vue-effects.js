// Vue.js 3 Effects Composables для Go Quiz

// === Confetti Effect ===
export function useConfetti() {
    const confettiParticles = ref([]);
    const isActive = ref(false);

    function createConfetti(x = 50, y = 50) {
        isActive.value = true;
        const colors = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];
        
        for (let i = 0; i < 60; i++) {
            confettiParticles.value.push({
                id: Date.now() + i,
                x: x,
                y: y,
                vx: (Math.random() - 0.5) * 20,
                vy: (Math.random() - 0.5) * 20 - 5,
                size: Math.random() * 10 + 5,
                color: colors[Math.floor(Math.random() * colors.length)],
                rotation: Math.random() * 360,
                rotationSpeed: (Math.random() - 0.5) * 10,
                gravity: 0.5,
                drag: 0.96
            });
        }

        requestAnimationFrame(animateConfetti);
        
        setTimeout(() => {
            isActive.value = false;
            confettiParticles.value = [];
        }, 3000);
    }

    function animateConfetti() {
        if (!isActive.value) return;

        confettiParticles.value.forEach(p => {
            p.x += p.vx;
            p.y += p.vy;
            p.vy += p.gravity;
            p.vx *= p.drag;
            p.vy *= p.drag;
            p.rotation += p.rotationSpeed;
        });

        requestAnimationFrame(animateConfetti);
    }

    return { confettiParticles, isActive, createConfetti };
}

// === Particle System ===
export function useParticles() {
    const particles = ref([]);

    function emitParticles(x, y, count = 20, color = '#6366f1') {
        const newParticles = [];
        for (let i = 0; i < count; i++) {
            const angle = (Math.PI * 2 * i) / count;
            const speed = Math.random() * 5 + 3;
            newParticles.push({
                id: Date.now() + i,
                x,
                y,
                vx: Math.cos(angle) * speed,
                vy: Math.sin(angle) * speed,
                size: Math.random() * 6 + 3,
                color: color,
                life: 1,
                decay: Math.random() * 0.03 + 0.02
            });
        }
        particles.value.push(...newParticles);
        animateParticles();
    }

    function animateParticles() {
        if (particles.value.length === 0) return;

        particles.value = particles.value
            .map(p => ({
                ...p,
                x: p.x + p.vx,
                y: p.y + p.vy,
                life: p.life - p.decay,
                vy: p.vy + 0.2
            }))
            .filter(p => p.life > 0);

        if (particles.value.length > 0) {
            requestAnimationFrame(animateParticles);
        }
    }

    return { particles, emitParticles };
}

// === Toast Notifications ===
export function useToast() {
    const toasts = ref([]);

    function addToast(message, type = 'info', duration = 3000) {
        const id = Date.now();
        toasts.value.push({ id, message, type });

        setTimeout(() => {
            removeToast(id);
        }, duration);
    }

    function removeToast(id) {
        const index = toasts.value.findIndex(t => t.id === id);
        if (index > -1) {
            toasts.value.splice(index, 1);
        }
    }

    function success(message) {
        addToast(message, 'success');
    }

    function error(message) {
        addToast(message, 'error');
    }

    function info(message) {
        addToast(message, 'info');
    }

    function warning(message) {
        addToast(message, 'warning');
    }

    return { toasts, addToast, success, error, info, warning };
}

// === Level Up Animation ===
export function useLevelUp() {
    const isAnimating = ref(false);
    const level = ref(1);
    const stars = ref([]);

    function triggerLevelUp(newLevel) {
        isAnimating.value = true;
        level.value = newLevel;
        
        stars.value = [];
        for (let i = 0; i < 10; i++) {
            stars.value.push({
                id: i,
                x: 50 + (Math.random() - 0.5) * 80,
                y: 50 + (Math.random() - 0.5) * 80,
                scale: 0,
                rotation: Math.random() * 360,
                delay: i * 0.1
            });
        }

        setTimeout(() => {
            isAnimating.value = false;
        }, 2500);
    }

    return { isAnimating, level, stars, triggerLevelUp };
}

// === Combo Counter ===
export function useCombo() {
    const combo = ref(0);
    const maxCombo = ref(0);
    const showCombo = ref(false);
    const comboTimeout = ref(null);

    function addCombo() {
        combo.value++;
        if (combo.value > maxCombo.value) {
            maxCombo.value = combo.value;
        }
        showCombo.value = true;

        if (comboTimeout.value) {
            clearTimeout(comboTimeout.value);
        }

        comboTimeout.value = setTimeout(() => {
            combo.value = 0;
            showCombo.value = false;
        }, 3000);
    }

    function resetCombo() {
        combo.value = 0;
        showCombo.value = false;
    }

    return { combo, maxCombo, showCombo, addCombo, resetCombo };
}

// === Shake Effect ===
export function useShake() {
    const isShaking = ref(false);
    const shakeIntensity = ref(10);

    function shake(intensity = 10, duration = 500) {
        isShaking.value = true;
        shakeIntensity.value = intensity;
        
        setTimeout(() => {
            isShaking.value = false;
        }, duration);
    }

    return { isShaking, shakeIntensity, shake };
}

// === Ripple Effect ===
export function useRipple() {
    const ripples = ref([]);

    function addRipple(event, element) {
        if (!element) return;
        const rect = element.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        const ripple = {
            id: Date.now(),
            x,
            y,
            scale: 0,
            opacity: 1
        };
        
        ripples.value.push(ripple);
        
        const startTime = Date.now();
        const duration = 600;
        
        function animate() {
            const elapsed = Date.now() - startTime;
            const progress = elapsed / duration;
            
            if (progress < 1) {
                ripple.scale = progress * 3;
                ripple.opacity = 1 - progress;
                requestAnimationFrame(animate);
            } else {
                ripples.value = ripples.value.filter(r => r.id !== ripple.id);
            }
        }
        
        requestAnimationFrame(animate);
    }

    return { ripples, addRipple };
}

// === Floating Text ===
export function useFloatingText() {
    const texts = ref([]);

    function showFloatingText(text, x, y, color = '#6366f1') {
        const id = Date.now();
        texts.value.push({ id, text, x, y, color, opacity: 1, yOffset: 0 });
        
        const startTime = Date.now();
        const duration = 1500;
        
        function animate() {
            const elapsed = Date.now() - startTime;
            const progress = elapsed / duration;
            
            const textObj = texts.value.find(t => t.id === id);
            if (textObj) {
                textObj.opacity = 1 - progress;
                textObj.yOffset = -progress * 100;
            }
            
            if (progress < 1) {
                requestAnimationFrame(animate);
            } else {
                texts.value = texts.value.filter(t => t.id !== id);
            }
        }
        
        requestAnimationFrame(animate);
    }

    return { texts, showFloatingText };
}

// Экспортируем глобально для window.VueEffects
if (typeof window !== 'undefined') {
    window.VueEffects = {
        useConfetti,
        useParticles,
        useToast,
        useLevelUp,
        useCombo,
        useShake,
        useRipple,
        useFloatingText
    };
}
