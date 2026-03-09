// Go Quiz Web Application - JavaScript

let userId = localStorage.getItem('goquiz_user_id');
if (!userId) {
    userId = 'user_' + Math.random().toString(36).substr(2, 9);
    localStorage.setItem('goquiz_user_id', userId);
}
document.cookie = "user_id=" + userId + "; path=/; max-age=31536000";

function getHeaders() {
    return {'Content-Type': 'application/json', 'X-User-ID': userId};
}

let currentQuestion = null;
let answered = false;

// Theme
const savedTheme = localStorage.getItem('goquiz_theme');
if (savedTheme === 'light') {
    document.body.classList.add('light-theme');
    document.querySelector('.theme-toggle').textContent = '🌙';
}
function toggleTheme() {
    document.body.classList.toggle('light-theme');
    const icon = document.querySelector('.theme-toggle');
    if (document.body.classList.contains('light-theme')) {
        icon.textContent = '🌙';
        localStorage.setItem('goquiz_theme', 'light');
    } else {
        icon.textContent = '☀️';
        localStorage.setItem('goquiz_theme', 'dark');
    }
}

// Navigation
function showPage(pageId) {
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
    document.getElementById(pageId).classList.add('active');
    
    if (pageId === 'stats') loadStats();
    if (pageId === 'leaderboard') loadLeaderboard();
    if (pageId === 'skills') loadSkills();
    if (pageId === 'quests') loadQuests();
    if (pageId === 'achievements') loadAchievements();
}

// Quiz
function startQuiz() {
    fetch('/api/quiz', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            currentQuestion = data.question;
            document.getElementById('question-counter').textContent = 'Вопрос ' + (data.answered + 1) + ' из ' + data.total;
            document.getElementById('question-text').textContent = data.question.Question;
            const opts = document.getElementById('options-container');
            opts.innerHTML = '';
            data.question.Options.forEach((opt, i) => {
                const btn = document.createElement('button');
                btn.className = 'option-btn';
                btn.textContent = opt;
                btn.onclick = () => answerQuestion(i);
                opts.appendChild(btn);
            });
            document.getElementById('next-btn').classList.remove('visible');
            answered = false;
            showPage('quiz');
        });
}

function answerQuestion(optionIndex) {
    if (answered) return;
    answered = true;

    fetch('/api/answer', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({question_id: currentQuestion.ID, option_index: optionIndex})
    })
    .then(r => r.json())
    .then(data => {
        const opts = document.querySelectorAll('.option-btn');
        opts.forEach((btn, i) => {
            btn.classList.add('disabled');
            if (i === data.correct_option) btn.classList.add('correct');
            if (i === optionIndex && !data.correct) btn.classList.add('wrong');
        });
        document.getElementById('exp-display').textContent = 'EXP: ' + data.new_exp + ' (Ур. ' + data.new_level + ')';
        document.getElementById('level-display').textContent = 'Уровень ' + data.new_level;
        document.getElementById('next-btn').classList.add('visible');
        
        if (data.level_up) {
            alert('🎉 Уровень повышен: ' + data.new_level + '!');
        }
    });
}

function nextQuestion() {
    startQuiz();
}

// Stats
function loadStats() {
    fetch('/api/stats', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            document.getElementById('stat-level').textContent = data.player.level;
            document.getElementById('stat-exp').textContent = data.player.experience;
            document.getElementById('stat-correct').textContent = data.player.correct_answers;
            document.getElementById('stat-wrong').textContent = data.player.wrong_answers;
            document.getElementById('stat-knowledge').textContent = data.player.go_knowledge + '/100';
            document.getElementById('stat-focus').textContent = data.player.focus + '%';
            document.getElementById('stat-willpower').textContent = data.player.willpower + '%';
            document.getElementById('stat-rating').textContent = data.player.rating;
        });
}

// Leaderboard
function loadLeaderboard() {
    fetch('/api/leaderboard', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            const tbody = document.getElementById('leaderboard-body');
            tbody.innerHTML = '';
            data.entries.forEach((entry, i) => {
                const tr = document.createElement('tr');
                if (i < 3) tr.className = 'rank-' + (i+1);
                tr.innerHTML = 
                    '<td><span class="rank-badge">' + (i+1) + '</span></td>' +
                    '<td>' + entry.name + '</td>' +
                    '<td>' + entry.level + '</td>' +
                    '<td>' + entry.rating + '</td>' +
                    '<td>' + entry.correct + '</td>';
                tbody.appendChild(tr);
            });
        });
}

// Skills
function loadSkills() {
    fetch('/api/skills', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            document.getElementById('skill-points-display').textContent = 
                '✨ Очки навыков: ' + data.tree.skill_points + ' (всего: ' + data.tree.total_points + ')';
            
            const container = document.getElementById('skills-container');
            const categories = {
                '📚 GO-НАВЫКИ': ['go_basics', 'concurrency', 'interfaces', 'web_frameworks', 'databases', 'microservices'],
                '🎯 ФОКУС': ['focus_master', 'meditation', 'anti_procrastination'],
                '💪 СИЛА ВОЛИ': ['willpower', 'discipline'],
                '💰 ФИНАНСЫ': ['money_management']
            };

            let html = '';
            for (const catName in categories) {
                const skillIds = categories[catName];
                html += '<div class="skill-category"><h3>' + catName + '</h3>';
                for (const id of skillIds) {
                    const skill = data.tree.skills[id];
                    if (!skill) continue;
                    const barPercent = (skill.level / skill.max_level) * 100;
                    const bonusTotal = data.bonuses[skill.bonus_type] || 0;
                    html += 
                        '<div class="skill-item">' +
                            '<div class="skill-header">' +
                                '<div class="skill-name">' + skill.icon + ' ' + skill.name + '</div>' +
                                '<div class="skill-level">Ур. ' + skill.level + '/' + skill.max_level + '</div>' +
                            '</div>' +
                            '<div class="skill-bar"><div class="skill-bar-fill" style="width: ' + barPercent + '%"></div></div>' +
                            '<div class="skill-description">' + skill.description + '</div>' +
                            '<div class="skill-bonus">+' + (skill.bonus_value * skill.level) + ' к ' + getBonusName(skill.bonus_type) + ' (всего: +' + bonusTotal + ')</div>' +
                            '<button class="upgrade-btn" onclick="upgradeSkill(\'' + id + '\')" ' + 
                                (!skill.unlocked || skill.level >= skill.max_level ? 'disabled' : '') + '>' +
                                '⬆️ Улучшить (' + skill.cost + ' очк.)' +
                            '</button>' +
                        '</div>';
                }
                html += '</div>';
            }
            container.innerHTML = html;
        });
}

function getBonusName(type) {
    const names = {focus: 'Фокус', willpower: 'Сила воли', knowledge: 'Знание Go', money: 'Деньги', dopamine: 'Дофамин'};
    return names[type] || type;
}

function upgradeSkill(skillId) {
    fetch('/api/skills/upgrade', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({skill_id: skillId})
    })
    .then(r => r.json())
    .then(data => {
        alert(data.message);
        loadSkills();
        loadStats();
    });
}

// Quests
function loadQuests() {
    fetch('/api/quests', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            const container = document.getElementById('quests-container');
            let html = '';
            data.system.quests.forEach(quest => {
                const percent = (quest.progress / quest.goal) * 100;
                const status = quest.completed ? (quest.claimed ? '✅' : '🎁') : '⏳';
                html += 
                    '<div class="quest-item">' +
                        '<div class="quest-header">' +
                            '<div class="quest-name">' + status + ' ' + quest.name + '</div>' +
                            '<div class="quest-status">' + quest.progress + '/' + quest.goal + '</div>' +
                        '</div>' +
                        '<div class="quest-progress-bar"><div class="quest-progress-fill" style="width: ' + percent + '%"></div></div>' +
                        '<div>' + quest.description + '</div>' +
                        (quest.completed && !quest.claimed ? 
                            '<button class="claim-btn" onclick="claimQuest(\'' + quest.id + '\')">🎁 Забрать (' + quest.reward + ' очк.)</button>' : 
                            '') +
                    '</div>';
            });
            html += '<p>🔥 Серия дней: ' + data.system.streak + '</p>';
            container.innerHTML = html;
        });
}

function claimQuest(questId) {
    alert('Награда будет начислена автоматически!');
    loadQuests();
}

// Achievements
function loadAchievements() {
    fetch('/api/achievements', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            const container = document.getElementById('achievements-container');
            let html = '<p style="margin-bottom: 20px;">Всего разблокировано: ' + data.unlocked_count + '/' + data.total_count + '</p>';
            
            const achievements = Object.values(data.system.achievements);
            achievements.forEach(ach => {
                html += 
                    '<div class="achievement-item ' + (ach.unlocked ? 'unlocked' : '') + '">' +
                        '<div class="achievement-icon">' + (ach.unlocked ? ach.icon : '🔒') + '</div>' +
                        '<div class="achievement-info">' +
                            '<div class="achievement-name">' + ach.name + '</div>' +
                            '<div class="achievement-description">' + ach.description + '</div>' +
                        '</div>' +
                    '</div>';
            });
            container.innerHTML = html;
        });
}

// Study & Rest
function studyGo(minutes) {
    fetch('/api/study', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({minutes: minutes})
    })
    .then(r => r.json())
    .then(data => {
        alert(data.message);
        loadStats();
    });
}

function rest(minutes) {
    fetch('/api/rest', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({minutes: minutes})
    })
    .then(r => r.json())
    .then(data => {
        alert(data.message);
        loadStats();
    });
}

function createBackup() {
    fetch('/api/backup', {headers: getHeaders()})
        .then(r => r.json())
        .then(data => {
            alert('✅ ' + data.message);
        });
}

// Reset
function resetProgress() {
    if (confirm('Вы уверены? Весь прогресс будет сброшен!')) {
        fetch('/api/reset', {method: 'POST', headers: getHeaders()})
            .then(() => {
                alert('Прогресс сброшен');
                location.reload();
            });
    }
}

// Auto-load stats on first visit
loadStats();
