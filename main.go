package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

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
                    droplet.style.transform = `translate(${deltaX}px, ${deltaY}px)`;
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

	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
