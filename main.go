package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

// Holiday represents a holiday event
type Holiday struct {
	Name        string
	Description string
	Type        string // international, national, professional
}

// getHolidaysForMarch5 returns holidays for March 5th
func getHolidaysForMarch5() []Holiday {
	return []Holiday{
		{Name: "Всемирный день эффективности", Description: "Международный праздник, посвященный повышению личной и профессиональной эффективности", Type: "international"},
		{Name: "День архивариуса", Description: "Профессиональный праздник работников архивов в России", Type: "professional"},
		{Name: "День физкультурника", Description: "Праздник здорового образа жизни и спорта", Type: "professional"},
		{Name: "День рождения А. С. Пушкина", Description: "День рождения великого русского поэта (1799 год)", Type: "cultural"},
		{Name: "День обретения мощей блаженной Ксении Петербургской", Description: "Православный праздник", Type: "religious"},
	}
}

var tmpl = template.Must(template.New("time").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Current Time</title>
    <style>
        * {
            -webkit-tap-highlight-color: transparent;
            box-sizing: border-box;
        }
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #1a1a2e, #16213e);
            font-family: 'Roboto', -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            transition: background 0.3s ease;
            overflow-x: hidden;
        }
        body.light-theme {
            background: linear-gradient(135deg, #f5f5f5, #e0e0e0);
        }
        .clock {
            text-align: center;
            color: #00d9ff;
            text-shadow: 0 0 20px rgba(0, 217, 255, 0.5);
            transition: color 0.3s ease, text-shadow 0.3s ease;
            padding: 20px;
        }
        body.light-theme .clock {
            color: #333;
            text-shadow: none;
        }
        .time {
            font-size: 5rem;
            font-weight: bold;
            letter-spacing: 0.2rem;
            user-select: none;
            -webkit-user-select: none;
        }
        .label {
            font-size: 1.2rem;
            color: #888;
            margin-top: 1rem;
            font-weight: 300;
        }
        body.light-theme .label {
            color: #555;
        }
        .date-display {
            font-size: 1.5rem;
            color: #00d9ff;
            text-shadow: 0 0 15px rgba(0, 217, 255, 0.4);
            margin-top: 0.5rem;
            font-weight: 300;
            transition: all 0.3s ease;
            opacity: 0;
            transform: translateY(20px);
        }
        body.light-theme .date-display {
            color: #333;
            text-shadow: none;
        }
        .date-display.visible {
            opacity: 1;
            transform: translateY(0);
        }
        .date-btn {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 25px;
            padding: 10px 20px;
            cursor: pointer;
            color: #00d9ff;
            font-size: 0.9rem;
            transition: all 0.3s ease;
            z-index: 100;
            -webkit-tap-highlight-color: transparent;
            white-space: nowrap;
        }
        body.light-theme .date-btn {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
            color: #333;
        }
        .date-btn:hover {
            background: rgba(255, 255, 255, 0.2);
            transform: scale(1.05);
            box-shadow: 0 0 20px rgba(0, 217, 255, 0.5);
        }
        body.light-theme .date-btn:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .date-btn:active {
            transform: scale(0.95);
        }
        .holidays-btn {
            position: fixed;
            bottom: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 25px;
            padding: 10px 20px;
            cursor: pointer;
            color: #00d9ff;
            font-size: 0.9rem;
            transition: all 0.3s ease;
            z-index: 100;
            -webkit-tap-highlight-color: transparent;
            white-space: nowrap;
            text-decoration: none;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }
        body.light-theme .holidays-btn {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
            color: #333;
        }
        .holidays-btn:hover {
            background: rgba(255, 255, 255, 0.2);
            transform: scale(1.05);
            box-shadow: 0 0 20px rgba(0, 217, 255, 0.5);
        }
        body.light-theme .holidays-btn:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .holidays-btn:active {
            transform: scale(0.95);
        }
        .theme-toggle {
            position: fixed;
            top: 20px;
            right: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 50%;
            width: 50px;
            height: 50px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            z-index: 100;
            -webkit-tap-highlight-color: transparent;
        }
        body.light-theme .theme-toggle {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
        }
        .theme-toggle:hover {
            transform: scale(1.1);
            background: rgba(255, 255, 255, 0.2);
        }
        body.light-theme .theme-toggle:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .theme-toggle:active {
            transform: scale(0.95);
        }
        .theme-toggle span {
            font-size: 1.5rem;
        }
        #date {
            transition: color 0.3s ease;
            text-align: center;
            padding: 10px;
        }
        .bubble-btn {
            position: fixed;
            top: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 25px;
            padding: 10px 20px;
            cursor: pointer;
            color: #00d9ff;
            font-size: 0.9rem;
            transition: all 0.3s ease;
            z-index: 100;
            -webkit-tap-highlight-color: transparent;
            white-space: nowrap;
        }
        body.light-theme .bubble-btn {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
            color: #333;
        }
        .bubble-btn:hover {
            background: rgba(255, 255, 255, 0.2);
            transform: scale(1.05);
        }
        body.light-theme .bubble-btn:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .bubble-btn:active {
            transform: scale(0.95);
        }
        .bubble {
            position: absolute;
            border-radius: 50%;
            background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.8), rgba(0, 217, 255, 0.4));
            box-shadow: 0 0 10px rgba(0, 217, 255, 0.3);
            cursor: pointer;
            animation: float 3s ease-in-out infinite;
            transition: transform 0.1s ease;
            user-select: none;
            -webkit-user-select: none;
        }
        body.light-theme .bubble {
            background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.9), rgba(100, 150, 200, 0.5));
            box-shadow: 0 0 10px rgba(100, 150, 200, 0.4);
        }
        .bubble:hover {
            transform: scale(1.1);
        }
        .bubble:active {
            transform: scale(0.9);
        }
        .bubble.pop {
            animation: pop 0.2s ease-out forwards;
        }
        .droplet {
            position: absolute;
            border-radius: 50%;
            background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.9), rgba(0, 217, 255, 0.5));
            box-shadow: 0 0 8px rgba(0, 217, 255, 0.4);
            pointer-events: none;
            user-select: none;
            -webkit-user-select: none;
        }
        @keyframes float {
            0%, 100% { transform: translateY(0px); }
            50% { transform: translateY(-20px); }
        }
        @keyframes pop {
            0% { transform: scale(1); opacity: 1; }
            100% { transform: scale(1.5); opacity: 0; }
        }
        #bubbles-container {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            overflow: hidden;
            z-index: 1;
        }
        #bubbles-container .bubble {
            pointer-events: auto;
        }
        /* Material Design Breakpoints - Mobile First */
        @media (max-width: 480px) {
            .time {
                font-size: 2.5rem;
                letter-spacing: 0.1rem;
            }
            .label {
                font-size: 0.9rem;
            }
            .theme-toggle {
                top: 10px;
                right: 10px;
                width: 44px;
                height: 44px;
            }
            .theme-toggle span {
                font-size: 1.2rem;
            }
            .bubble-btn {
                top: 10px;
                left: 10px;
                padding: 8px 14px;
                font-size: 0.75rem;
            }
            .date-btn {
                bottom: 10px;
                right: 10px;
                padding: 8px 14px;
                font-size: 0.75rem;
            }
            .holidays-btn {
                bottom: 10px;
                left: 10px;
                padding: 8px 14px;
                font-size: 0.75rem;
            }
            #date {
                bottom: 10px;
                right: 10px;
                left: 10px;
                font-size: 0.75rem;
            }
            .clock {
                padding: 15px;
            }
        }
        @media (min-width: 481px) and (max-width: 768px) {
            .time {
                font-size: 3.5rem;
                letter-spacing: 0.15rem;
            }
            .label {
                font-size: 1rem;
            }
            .theme-toggle {
                width: 48px;
                height: 48px;
            }
            .theme-toggle span {
                font-size: 1.3rem;
            }
            .bubble-btn {
                padding: 9px 16px;
                font-size: 0.8rem;
            }
            #date {
                font-size: 0.85rem;
            }
        }
        @media (min-width: 769px) and (max-width: 1024px) {
            .time {
                font-size: 4rem;
            }
            .label {
                font-size: 1.1rem;
            }
        }
        /* Landscape orientation on mobile */
        @media (max-height: 500px) and (orientation: landscape) {
            .time {
                font-size: 3rem;
            }
            .label {
                font-size: 0.85rem;
            }
            .theme-toggle {
                top: 5px;
                right: 5px;
            }
            .bubble-btn {
                top: 5px;
                left: 5px;
                padding: 6px 12px;
                font-size: 0.7rem;
            }
        }
        /* Touch device optimizations */
        @media (hover: none) and (pointer: coarse) {
            .theme-toggle,
            .bubble-btn {
                min-width: 48px;
                min-height: 48px;
            }
            .bubble {
                min-width: 40px;
                min-height: 40px;
            }
        }
    </style>
</head>
<body>
    <button class="bubble-btn" onclick="toggleBubbles()">🫧 Активировать пузырьки</button>
    <button class="theme-toggle" onclick="toggleTheme()" title="Переключить тему">
        <span id="theme-icon">☀️</span>
    </button>
    <button class="date-btn" onclick="toggleDateDisplay()">📅 Показать дату</button>
    <a href="/holidays" class="holidays-btn">🎉 Какой праздник сегодня</a>
    <div id="bubbles-container"></div>
    <div class="clock">
        <div class="time" id="time">{{ .Time }}</div>
        <div class="label">Московское время сейчас</div>
        <div class="date-display" id="date"></div>
    </div>
    <script>
        let bubblesActive = false;
        let bubblesInterval = null;
        
        function toggleBubbles() {
            const btn = document.querySelector('.bubble-btn');
            if (bubblesActive) {
                bubblesActive = false;
                btn.textContent = '🫧 Активировать пузырьки';
                clearInterval(bubblesInterval);
                // Remove all bubbles
                const container = document.getElementById('bubbles-container');
                container.innerHTML = '';
            } else {
                bubblesActive = true;
                btn.textContent = '🫧 Отключить пузырьки';
                createBubble();
                bubblesInterval = setInterval(createBubble, 500);
            }
        }
        
        function createBubble() {
            const container = document.getElementById('bubbles-container');
            const bubble = document.createElement('div');
            bubble.className = 'bubble';
            
            // Random size between 20 and 60px
            const size = Math.random() * 40 + 20;
            bubble.style.width = size + 'px';
            bubble.style.height = size + 'px';
            
            // Random position
            bubble.style.left = Math.random() * window.innerWidth + 'px';
            bubble.style.top = Math.random() * window.innerHeight + 'px';
            
            // Random animation delay
            bubble.style.animationDelay = Math.random() * 2 + 's';
            
            // Pop on click
            bubble.onclick = function() {
                createDroplets(bubble);
                bubble.classList.add('pop');
                setTimeout(() => bubble.remove(), 200);
            };
            
            // Remove bubble after 10 seconds
            setTimeout(() => {
                if (bubble.parentNode) {
                    bubble.style.transition = 'opacity 1s ease';
                    bubble.style.opacity = '0';
                    setTimeout(() => bubble.remove(), 1000);
                }
            }, 10000);
            
            container.appendChild(bubble);
        }

        function createDroplets(bubble) {
            const container = document.getElementById('bubbles-container');
            const bubbleRect = bubble.getBoundingClientRect();
            const centerX = bubbleRect.left + bubbleRect.width / 2;
            const centerY = bubbleRect.top + bubbleRect.height / 2;
            const numDroplets = Math.floor(Math.random() * 6) + 8;

            for (let i = 0; i < numDroplets; i++) {
                const droplet = document.createElement('div');
                droplet.className = 'droplet';

                const size = Math.random() * 8 + 4;
                droplet.style.width = size + 'px';
                droplet.style.height = size + 'px';
                droplet.style.left = centerX + 'px';
                droplet.style.top = centerY + 'px';

                const angle = (Math.PI * 2 * i) / numDroplets + Math.random() * 0.5;
                const velocity = Math.random() * 60 + 40;
                const deltaX = Math.cos(angle) * velocity;
                const deltaY = Math.sin(angle) * velocity;

                droplet.style.transition = 'all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)';
                container.appendChild(droplet);

                requestAnimationFrame(() => {
                    droplet.style.transform = 'translate(' + deltaX + 'px, ' + deltaY + 'px)';
                    droplet.style.opacity = '0';
                });

                setTimeout(() => droplet.remove(), 600);
            }
        }

        function toggleTheme() {
            document.body.classList.toggle('light-theme');
            const icon = document.getElementById('theme-icon');
            if (document.body.classList.contains('light-theme')) {
                icon.textContent = '🌙';
                localStorage.setItem('theme', 'light');
            } else {
                icon.textContent = '☀️';
                localStorage.setItem('theme', 'dark');
            }
        }
        // Load saved theme on page load
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme === 'light') {
            document.body.classList.add('light-theme');
            document.getElementById('theme-icon').textContent = '🌙';
        }
        let dateVisible = false;
        function toggleDateDisplay() {
            dateVisible = !dateVisible;
            const btn = document.querySelector('.date-btn');
            const dateEl = document.getElementById('date');
            if (dateVisible) {
                btn.textContent = '📅 Скрыть дату';
                dateEl.classList.add('visible');
                updateDate();
            } else {
                btn.textContent = '📅 Показать дату';
                dateEl.classList.remove('visible');
            }
        }
        function updateTime() {
            const now = new Date();
            const hours = String(now.getHours()).padStart(2, '0');
            const minutes = String(now.getMinutes()).padStart(2, '0');
            const seconds = String(now.getSeconds()).padStart(2, '0');
            document.getElementById('time').textContent = hours + '-' + minutes + '-' + seconds;
        }
        setInterval(updateTime, 1000);

        function updateDate() {
            const now = new Date();
            const months = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'];
            const dayNum = now.getDate();
            const monthName = months[now.getMonth()];
            const year = now.getFullYear();
            document.getElementById('date').textContent = 'сегодня ' + dayNum + ' ' + monthName + ' ' + year + ' года';
        }
        updateDate();
        setInterval(updateDate, 1000);
    </script>
</body>
</html>
`))

var holidaysTmpl = template.Must(template.New("holidays").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Праздники сегодня</title>
    <link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;600;700&family=Open+Sans:wght@400;600&display=swap" rel="stylesheet">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Open Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            min-height: 100vh;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #fff;
            padding: 20px;
            transition: background 0.3s ease, color 0.3s ease;
        }
        body.light-theme {
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 50%, #e8eaf6 100%);
            color: #333;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            padding-top: 20px;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            animation: slideDown 0.5s ease-out;
        }
        .header h1 {
            font-family: 'Montserrat', sans-serif;
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 10px;
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        body.light-theme .header h1 {
            background: linear-gradient(135deg, #1a1a2e, #0f3460);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .header .date {
            font-size: 1.2rem;
            color: #888;
            font-weight: 400;
        }
        body.light-theme .header .date {
            color: #555;
        }
        .holidays-list {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }
        .holiday-card {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 20px;
            padding: 25px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            transition: all 0.3s ease;
            animation: fadeInUp 0.5s ease-out backwards;
            cursor: pointer;
        }
        body.light-theme .holiday-card {
            background: rgba(255, 255, 255, 0.8);
            border-color: rgba(0, 0, 0, 0.1);
        }
        .holiday-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 30px rgba(0, 217, 255, 0.3);
            border-color: #00d9ff;
        }
        body.light-theme .holiday-card:hover {
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
        }
        .holiday-card:nth-child(1) { animation-delay: 0.1s; }
        .holiday-card:nth-child(2) { animation-delay: 0.15s; }
        .holiday-card:nth-child(3) { animation-delay: 0.2s; }
        .holiday-card:nth-child(4) { animation-delay: 0.25s; }
        .holiday-card:nth-child(5) { animation-delay: 0.3s; }
        .holiday-name {
            font-family: 'Montserrat', sans-serif;
            font-size: 1.3rem;
            font-weight: 600;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .holiday-type {
            font-size: 0.75rem;
            padding: 4px 12px;
            border-radius: 20px;
            text-transform: uppercase;
            font-weight: 600;
            letter-spacing: 0.5px;
        }
        .type-international {
            background: linear-gradient(135deg, #667eea, #764ba2);
            color: #fff;
        }
        .type-national {
            background: linear-gradient(135deg, #f093fb, #f5576c);
            color: #fff;
        }
        .type-professional {
            background: linear-gradient(135deg, #4facfe, #00f2fe);
            color: #fff;
        }
        .type-cultural {
            background: linear-gradient(135deg, #fa709a, #fee140);
            color: #fff;
        }
        .type-religious {
            background: linear-gradient(135deg, #a8edea, #fed6e3);
            color: #333;
        }
        .holiday-description {
            font-size: 0.95rem;
            line-height: 1.6;
            color: #aaa;
        }
        body.light-theme .holiday-description {
            color: #555;
        }
        .back-btn {
            position: fixed;
            top: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 50%;
            width: 50px;
            height: 50px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            z-index: 100;
            text-decoration: none;
            -webkit-tap-highlight-color: transparent;
        }
        body.light-theme .back-btn {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
        }
        .back-btn:hover {
            transform: scale(1.1) rotate(-10deg);
            background: rgba(255, 255, 255, 0.2);
        }
        body.light-theme .back-btn:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .back-btn:active {
            transform: scale(0.95);
        }
        .back-btn span {
            font-size: 1.5rem;
            color: #00d9ff;
        }
        body.light-theme .back-btn span {
            color: #333;
        }
        .theme-toggle {
            position: fixed;
            top: 20px;
            right: 20px;
            background: rgba(255, 255, 255, 0.1);
            border: 2px solid #00d9ff;
            border-radius: 50%;
            width: 50px;
            height: 50px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            z-index: 100;
            -webkit-tap-highlight-color: transparent;
        }
        body.light-theme .theme-toggle {
            background: rgba(0, 0, 0, 0.1);
            border-color: #333;
        }
        .theme-toggle:hover {
            transform: scale(1.1);
            background: rgba(255, 255, 255, 0.2);
        }
        body.light-theme .theme-toggle:hover {
            background: rgba(0, 0, 0, 0.15);
        }
        .theme-toggle:active {
            transform: scale(0.95);
        }
        .theme-toggle span {
            font-size: 1.5rem;
        }
        @keyframes slideDown {
            from {
                opacity: 0;
                transform: translateY(-30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        @keyframes fadeInUp {
            from {
                opacity: 0;
                transform: translateY(30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        /* Mobile First - Responsive Design */
        @media (max-width: 480px) {
            body {
                padding: 15px;
            }
            .header h1 {
                font-size: 1.8rem;
            }
            .header .date {
                font-size: 1rem;
            }
            .holiday-card {
                padding: 20px;
            }
            .holiday-name {
                font-size: 1.1rem;
            }
            .holiday-description {
                font-size: 0.9rem;
            }
            .back-btn,
            .theme-toggle {
                width: 44px;
                height: 44px;
                top: 10px;
            }
            .back-btn {
                left: 10px;
            }
            .theme-toggle {
                right: 10px;
            }
        }
        @media (min-width: 481px) and (max-width: 768px) {
            .header h1 {
                font-size: 2rem;
            }
            .holiday-name {
                font-size: 1.2rem;
            }
        }
        @media (min-width: 769px) {
            .container {
                padding-top: 40px;
            }
            .header h1 {
                font-size: 3rem;
            }
        }
        /* Touch device optimizations */
        @media (hover: none) and (pointer: coarse) {
            .back-btn,
            .theme-toggle {
                min-width: 48px;
                min-height: 48px;
            }
            .holiday-card {
                min-height: 80px;
            }
        }
    </style>
</head>
<body>
    <a href="/" class="back-btn" title="На главную">
        <span>←</span>
    </a>
    <button class="theme-toggle" onclick="toggleTheme()" title="Переключить тему">
        <span id="theme-icon">☀️</span>
    </button>
    <div class="container">
        <div class="header">
            <h1>🎉 Праздники сегодня</h1>
            <p class="date">{{ .Date }}</p>
        </div>
        <div class="holidays-list">
            {{ range .Holidays }}
            <div class="holiday-card">
                <div class="holiday-name">
                    {{ .Name }}
                    <span class="holiday-type type-{{ .Type }}">{{ .Type }}</span>
                </div>
                <p class="holiday-description">{{ .Description }}</p>
            </div>
            {{ end }}
        </div>
    </div>
    <script>
        // Load saved theme
        const savedTheme = localStorage.getItem('holidays-theme');
        if (savedTheme === 'light') {
            document.body.classList.add('light-theme');
            document.getElementById('theme-icon').textContent = '🌙';
        }

        function toggleTheme() {
            document.body.classList.toggle('light-theme');
            const icon = document.getElementById('theme-icon');
            if (document.body.classList.contains('light-theme')) {
                icon.textContent = '🌙';
                localStorage.setItem('holidays-theme', 'light');
            } else {
                icon.textContent = '☀️';
                localStorage.setItem('holidays-theme', 'dark');
            }
        }
    </script>
</body>
</html>
`))

func timeHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	timeStr := fmt.Sprintf("%02d-%02d-%02d", now.Hour(), now.Minute(), now.Second())

	data := struct {
		Time string
	}{
		Time: timeStr,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", timeHandler)
	http.HandleFunc("/holidays", holidaysHandler)

	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// holidaysHandler renders the holidays page
func holidaysHandler(w http.ResponseWriter, r *http.Request) {
	holidays := getHolidaysForMarch5()
	
	data := struct {
		Holidays []Holiday
		Date     string
	}{
		Holidays: holidays,
		Date:     "5 марта",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := holidaysTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
