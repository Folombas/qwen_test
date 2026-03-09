# 🎮 Go Quiz Godot Game

**2D платформер-викторина на Godot Engine с экспортом в HTML5**

---

## 📁 Структура проекта

```
godot/
├── project.godot          # Конфигурация проекта Godot 4.2
├── icon.svg               # Иконка проекта
├── game/
│   ├── main.tscn         # Главная сцена
│   ├── main.gd           # Скрипт главной сцены
│   ├── player.tscn       # Сцена игрока
│   ├── player.gd         # Скрипт игрока (GDScript)
│   ├── question_orb.tscn # Сцена орба с вопросом
│   └── question_orb.gd   # Скрипт орба
└── scripts/
    └── ...               # Дополнительные скрипты
```

---

## 🛠️ Требования

- **Godot Engine 4.2+**
- **HTML5 Export Template** (для экспорта в веб)

---

## 🚀 Запуск в Godot

1. Откройте Godot Engine
2. Нажмите "Import" и выберите `project.godot`
3. Нажмите "Run" (F5) для запуска

---

## 🌐 Экспорт в HTML5

### Шаги экспорта:

1. **Установите HTML5 Export Template:**
   - В Godot: Project → Install Export Templates
   - Скачайте с https://godotengine.org/download/

2. **Настройте экспорт:**
   - Project → Export → Add → HTML5
   - Настройте размер: 800x600
   - Включите "Debug" для отладки

3. **Экспортируйте:**
   - Export → Export Project
   - Выберите папку: `../static/godot-export/`
   - Файл: `index.html`

4. **Результат:**
   ```
   static/godot-export/
   ├── index.html
   ├── project.html
   ├── project.js
   └── project.wasm
   ```

---

## 🔗 Интеграция с веб-приложением

### Godot ↔ JavaScript Bridge

```gdscript
# В GDScript
func _ready():
    if JavaScriptBridge:
        JavaScriptBridge.eval("window.godotBridge.onGameReady({})")
```

```javascript
// В JavaScript
window.godotBridge.on('gameReady', (data) => {
    console.log('Godot game ready:', data);
});
```

### API Endpoints

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/game` | GET | Данные для игры |
| `/api/answer` | POST | Ответ на вопрос |
| `/api/stats` | GET | Статистика игрока |

---

## 🎮 Геймплей

### Управление
- **A / D** - Движение влево/вправо
- **Space** - Прыжок
- **E** - Взаимодействие

### Механика
1. 🟣 Собирайте орбы с вопросами
2. ✅ Отвечайте правильно для получения EXP
3. 🔥 Поддерживайте комбо для бонусов
4. 📈 Повышайте уровень

---

## 🔧 Lua Логика

Игровая логика дублируется на Lua для возможности использования вне Godot:

```lua
-- lua/game_logic.lua
local player = GameLogic.initPlayer("player1")
local exp, msg = GameLogic.answerQuestion(player, true, 10)
```

---

## 📊 Архитектура

```
┌─────────────────────────────────────────────────┐
│              Веб-браузер                        │
│  ┌─────────────────────────────────────────┐   │
│  │         Vue.js Приложение               │   │
│  │  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ GodotGame   │  │  Quiz/Vue       │  │   │
│  │  │ Component   │  │  Components     │  │   │
│  │  └──────┬──────┘  └────────┬────────┘  │   │
│  │         │                  │            │   │
│  │  ┌──────▼──────────────────▼────────┐  │   │
│  │  │      Godot Bridge (JS)           │  │   │
│  │  └──────┬───────────────────────────┘  │   │
│  └─────────│──────────────────────────────┘   │
│            │                                   │
│  ┌─────────▼──────────────────────────────┐   │
│  │         Godot HTML5 (Canvas)           │   │
│  │  ┌─────────────┐  ┌─────────────────┐ │   │
│  │  │   GDScript  │  │   Lua Logic     │ │   │
│  │  │   (Player)  │  │   (game_logic)  │ │   │
│  │  └─────────────┘  └─────────────────┘ │   │
│  └─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
            ↕ HTTP API
┌─────────────────────────────────────────────────┐
│         Go Backend (qwen_test)                  │
│  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Handlers  │  │    Game Logic           │  │
│  │  (REST API) │  │    (internal/game)      │  │
│  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

---

## 🎯 Планы

- [ ] Полный экспорт Godot игры
- [ ] Интеграция с Lua через WASM
- [ ] Мультиплеер режим
- [ ] Новые уровни и боссы
- [ ] Система достижений в игре

---

## 📄 Лицензия

MIT

---

**Создано в рамках Go365 челленджа — 77 день (9 марта 2026)** 🚀
