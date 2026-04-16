package game

import (
	"testing"
)

func TestNewQuestSystem(t *testing.T) {
	qs := NewQuestSystem("user123")

	if qs.UserID != "user123" {
		t.Errorf("Ожидаемый UserID 'user123', получен '%s'", qs.UserID)
	}
	if len(qs.Quests) != 5 {
		t.Errorf("Ожидаемое количество квестов 5, получено %d", len(qs.Quests))
	}
	if qs.Day != 1 {
		t.Errorf("Ожидаемый день 1, получено %d", qs.Day)
	}
	if qs.Streak != 0 {
		t.Errorf("Ожидаемая серия 0, получено %d", qs.Streak)
	}
	if qs.TotalCompleted != 0 {
		t.Errorf("Ожидаемое всего выполнено 0, получено %d", qs.TotalCompleted)
	}
}

func TestQuestSystem_GenerateDailyQuests(t *testing.T) {
	qs := NewQuestSystem("user1")
	qs.GenerateDailyQuests()

	if len(qs.Quests) != 5 {
		t.Errorf("Ожидаемое количество квестов 5, получено %d", len(qs.Quests))
	}

	expectedQuestIDs := []string{
		"study_30",
		"temptation_fighter",
		"code_practice",
		"morning_ritual",
		"digital_detox",
	}

	for i, expectedID := range expectedQuestIDs {
		if i >= len(qs.Quests) {
			t.Errorf("Квест %d не существует", i)
			continue
		}
		if qs.Quests[i].ID != expectedID {
			t.Errorf("Ожидаемый ID квеста '%s', получен '%s'", expectedID, qs.Quests[i].ID)
		}
	}
}

func TestQuestTemplates(t *testing.T) {
	if len(DailyQuestsTemplates) != 5 {
		t.Errorf("Ожидаемое количество шаблонов 5, получено %d", len(DailyQuestsTemplates))
	}

	tests := []struct {
		id       string
		name     string
		goal     int
		reward   int
	}{
		{"study_30", "30 минут Go", 30, 2},
		{"temptation_fighter", "Борец с искушениями", 3, 3},
		{"code_practice", "Практика кода", 50, 3},
		{"morning_ritual", "Утренний ритуал", 1, 1},
		{"digital_detox", "Цифровой детокс", 240, 2},
	}

	for _, tt := range tests {
		found := false
		for _, tmpl := range DailyQuestsTemplates {
			if tmpl.ID == tt.id {
				found = true
				if tmpl.Name != tt.name {
					t.Errorf("Шаблон %s: ожидаемое имя '%s', получено '%s'", tt.id, tt.name, tmpl.Name)
				}
				if tmpl.Goal != tt.goal {
					t.Errorf("Шаблон %s: ожидаемая цель %d, получена %d", tt.id, tt.goal, tmpl.Goal)
				}
				if tmpl.Reward != tt.reward {
					t.Errorf("Шаблон %s: ожидаемая награда %d, получена %d", tt.id, tt.reward, tmpl.Reward)
				}
				break
			}
		}
		if !found {
			t.Errorf("Шаблон квеста %s не найден", tt.id)
		}
	}
}

func TestUpdateProgress_Complete(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Находим квест study_30
	qs.UpdateProgress("study_30", 15)

	quest := qs.Quests[0]
	if quest.Progress != 15 {
		t.Errorf("Ожидаемый прогресс 15, получено %d", quest.Progress)
	}
	if quest.Completed {
		t.Error("Квест не должен быть завершён")
	}

	// Добавляем ещё прогресса для завершения
	qs.UpdateProgress("study_30", 20)

	if quest.Progress != 30 {
		t.Errorf("Ожидаемый прогресс 30 (кап), получено %d", quest.Progress)
	}
	if !quest.Completed {
		t.Error("Квест должен быть завершён")
	}
}

func TestUpdateProgress_NegativeProgress(t *testing.T) {
	qs := NewQuestSystem("user1")
	initialProgress := qs.Quests[0].Progress

	qs.UpdateProgress("study_30", -10)

	if qs.Quests[0].Progress != initialProgress {
		t.Errorf("Ожидаемый прогресс без изменений (%d), получено %d", initialProgress, qs.Quests[0].Progress)
	}
}

func TestUpdateProgress_NonExistentQuest(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Не должно паниковать
	qs.UpdateProgress("nonexistent_quest", 10)

	// Проверяем, что ни один квест не изменился
	for _, quest := range qs.Quests {
		if quest.Progress != 0 {
			t.Errorf("Ожидаемый прогресс 0, получено %d", quest.Progress)
		}
	}
}

func TestUpdateProgress_AlreadyCompleted(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем квест
	qs.UpdateProgress("study_30", 30)
	qs.Quests[0].Completed = true
	qs.Quests[0].Claimed = true

	// Пытаемся добавить прогресс
	qs.UpdateProgress("study_30", 10)

	if qs.Quests[0].Progress != 30 {
		t.Errorf("Ожидаемый прогресс 30 (без изменений), получено %d", qs.Quests[0].Progress)
	}
}

func TestGetCompletedCount(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем 2 квеста
	qs.UpdateProgress("study_30", 30)
	qs.UpdateProgress("morning_ritual", 1)

	count := qs.GetCompletedCount()

	if count != 2 {
		t.Errorf("Ожидаемое количество завершённых 2, получено %d", count)
	}
}

func TestGetCompletedCount_Claimed(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем и забираем награду
	qs.UpdateProgress("study_30", 30)
	qs.Quests[0].Claimed = true

	count := qs.GetCompletedCount()

	if count != 0 {
		t.Errorf("Ожидаемое количество 0 (награда забрана), получено %d", count)
	}
}

func TestClaimRewards_Success(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем квест
	qs.UpdateProgress("study_30", 30)

	// Забираем награду
	reward := qs.ClaimRewards()

	if reward != 2 {
		t.Errorf("Ожидаемая награда 2, получено %d", reward)
	}
	if !qs.Quests[0].Claimed {
		t.Error("Квест должен быть помечен как полученный")
	}
}

func TestClaimRewards_MultipleQuests(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем 2 квеста
	qs.UpdateProgress("study_30", 30)
	qs.UpdateProgress("morning_ritual", 1)

	reward := qs.ClaimRewards()

	if reward != 3 {
		t.Errorf("Ожидаемая награда 3 (2+1), получено %d", reward)
	}
}

func TestClaimRewards_NoCompleted(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Пытаемся забрать награду незавершённых квестов
	reward := qs.ClaimRewards()

	if reward != 0 {
		t.Errorf("Ожидаемая награда 0 (квесты не завершены), получено %d", reward)
	}
}

func TestClaimRewards_AlreadyClaimed(t *testing.T) {
	qs := NewQuestSystem("user1")

	// Завершаем и забираем награду
	qs.UpdateProgress("study_30", 30)
	qs.ClaimRewards()

	// Пытаемся забрать снова
	reward2 := qs.ClaimRewards()

	if reward2 != 0 {
		t.Errorf("Ожидаемая награда 0 (уже забрана), получено %d", reward2)
	}
}
