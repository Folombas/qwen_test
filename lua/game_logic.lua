-- Go Quiz Game Logic (Lua)
-- Логика игры на Lua для Godot

-- Модуль игровой логики
GameLogic = {}

-- Константы
GameLogic.MAX_KNOWLEDGE = 100
GameLogic.BASE_EXP = 100
GameLogic.COMBO_MULTIPLIER = 1.5

-- Состояние игрока
PlayerState = {
    id = "",
    level = 1,
    experience = 0,
    go_knowledge = 0,
    focus = 100,
    willpower = 100,
    combo = 0,
    max_combo = 0,
    correct_answers = 0,
    wrong_answers = 0
}

-- Инициализация игрока
function GameLogic.initPlayer(playerId)
    local player = {
        id = playerId,
        level = 1,
        experience = 0,
        go_knowledge = 0,
        focus = 100,
        willpower = 100,
        combo = 0,
        max_combo = 0,
        correct_answers = 0,
        wrong_answers = 0
    }
    return player
end

-- Расчёт опыта для уровня
function GameLogic.expForLevel(level)
    return level * GameLogic.BASE_EXP
end

-- Проверка повышения уровня
function GameLogic.checkLevelUp(player)
    local required = GameLogic.expForLevel(player.level)
    local leveledUp = false
    
    while player.experience >= required do
        player.level = player.level + 1
        required = GameLogic.expForLevel(player.level)
        leveledUp = true
    end
    
    return leveledUp
end

-- Добавление опыта
function GameLogic.addExperience(player, amount)
    player.experience = player.experience + amount
    return GameLogic.checkLevelUp(player)
end

-- Ответ на вопрос
function GameLogic.answerQuestion(player, isCorrect, baseExp)
    local expGained = 0
    local message = ""
    
    if isCorrect then
        -- Правильный ответ
        player.combo = player.combo + 1
        if player.combo > player.max_combo then
            player.max_combo = player.combo
        end
        player.correct_answers = player.correct_answers + 1
        
        -- Опыт с комбо-множителем
        local comboBonus = math.min(player.combo, 10) * 0.1
        expGained = math.floor(baseExp * (1 + comboBonus))
        
        player.experience = player.experience + expGained
        
        -- Знание Go
        player.go_knowledge = math.min(player.go_knowledge + 5, GameLogic.MAX_KNOWLEDGE)
        
        message = string.format("✅ Правильно! +%d EXP (combo x%d)", expGained, player.combo)
        
        -- Проверка уровня
        local leveledUp = GameLogic.checkLevelUp(player)
        if leveledUp then
            message = message .. string.format("\n🎉 LEVEL UP! Уровень %d", player.level)
        end
    else
        -- Неправильный ответ
        player.combo = 0
        player.wrong_answers = player.wrong_answers + 1
        
        -- Штраф к опыту (но не ниже 0)
        local expLoss = math.min(baseExp / 2, player.experience)
        player.experience = player.experience - expLoss
        
        message = string.format("❌ Неправильно! Комбо сброшено", expGained)
    end
    
    return expGained, message
end

-- Сбор знания (орб)
function GameLogic.collectKnowledge(player, amount)
    local oldKnowledge = player.go_knowledge
    player.go_knowledge = math.min(player.go_knowledge + amount, GameLogic.MAX_KNOWLEDGE)
    
    local gained = player.go_knowledge - oldKnowledge
    local expGained = gained * 10
    
    player.experience = player.experience + expGained
    
    return expGained, player.go_knowledge
end

-- Расчёт рейтинга игрока
function GameLogic.calculateRating(player)
    return player.go_knowledge * 10 + 
           player.focus * 5 + 
           player.willpower * 3 + 
           player.level * 100
end

-- Получение ранга по рейтингу
function GameLogic.getRank(rating)
    if rating < 500 then
        return "🌱 Начинающий гофер"
    elseif rating < 1500 then
        return "🌿 Ученик разработчика"
    elseif rating < 3000 then
        return "🌳 Junior Go Developer"
    elseif rating < 5000 then
        return "🏢 Middle Go Developer"
    else
        return "🚀 Senior Go Master"
    end
end

-- Восстановление фокуса
function GameLogic.rest(player, minutes)
    local focusGain = math.floor(minutes / 2)
    local dopamineGain = math.floor(minutes / 3)
    
    player.focus = math.min(player.focus + focusGain, 100)
    
    return focusGain, dopamineGain
end

-- Изучение Go
function GameLogic.studyGo(player, minutes)
    local expGain = math.floor(minutes / 2)
    local knowledgeGain = math.floor(minutes / 5)
    local dopamineGain = math.floor(minutes / 3)
    
    player.experience = player.experience + expGain
    player.go_knowledge = math.min(player.go_knowledge + knowledgeGain, GameLogic.MAX_KNOWLEDGE)
    
    return expGain, knowledgeGain, dopamineGain
end

-- Проверка достижений
function GameLogic.checkAchievements(player)
    local unlocked = {}
    
    -- Достижения за уровни
    local levelAchievements = {
        [2] = "level_2",
        [5] = "level_5",
        [10] = "level_10",
        [20] = "level_20",
        [30] = "level_30"
    }
    
    if levelAchievements[player.level] then
        table.insert(unlocked, levelAchievements[player.level])
    end
    
    -- Достижения за комбо
    if player.max_combo >= 5 then
        table.insert(unlocked, "combo_5")
    end
    if player.max_combo >= 10 then
        table.insert(unlocked, "combo_10")
    end
    
    -- Достижения за знание Go
    if player.go_knowledge >= 50 then
        table.insert(unlocked, "knowledge_50")
    end
    if player.go_knowledge >= 100 then
        table.insert(unlocked, "knowledge_100")
    end
    
    return unlocked
end

-- Сериализация состояния игрока в JSON (упрощённая)
function GameLogic.playerToJSON(player)
    return string.format(
        '{"id":"%s","level":%d,"experience":%d,"go_knowledge":%d,"focus":%d,"willpower":%d,"combo":%d,"max_combo":%d,"correct_answers":%d,"wrong_answers":%d}',
        player.id,
        player.level,
        player.experience,
        player.go_knowledge,
        player.focus,
        player.willpower,
        player.combo,
        player.max_combo,
        player.correct_answers,
        player.wrong_answers
    )
end

-- Тесты
function GameLogic.runTests()
    print("🧪 Running GameLogic tests...")
    
    local player = GameLogic.initPlayer("test_player")
    assert(player.level == 1, "Initial level should be 1")
    assert(player.experience == 0, "Initial EXP should be 0")
    
    -- Тест ответа на вопрос
    local exp, msg = GameLogic.answerQuestion(player, true, 10)
    assert(exp > 0, "Should gain EXP for correct answer")
    assert(player.combo == 1, "Combo should be 1")
    
    -- Тест повышения уровня
    player.experience = 500
    local leveledUp = GameLogic.checkLevelUp(player)
    assert(leveledUp, "Should level up with 500 EXP")
    assert(player.level > 1, "Level should increase")
    
    print("✅ All tests passed!")
end

-- Запуск тестов при загрузке
GameLogic.runTests()

return GameLogic
