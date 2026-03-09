extends Node2D

# Go Quiz - Главная сцена игры
# Управляет игровым миром, вопросами и сбором знаний

@onready var player: CharacterBody2D = $Player
@onready var ui: CanvasLayer = $UI
@onready var question_manager: Node = $QuestionManager
@onready var background: ParallaxBackground = $Background

# Игровые переменные
var questions: Array = []
var active_orbs: Array = []
var game_state: String = "exploring"  # exploring, answering, transition

func _ready() -> void:
	# Инициализация игры
	print("🎮 Go Quiz Game запущен!")
	load_questions()
	spawn_question_orbs()
	setup_ui()

func _process(delta: float) -> void:
	# Обновление игры
	match game_state:
		"exploring":
			update_exploring_mode()
		"answering":
			update_answer_mode()

func load_questions() -> void:
	# Загружаем вопросы с сервера
	var http = HTTPRequest.new()
	add_child(http)
	http.request_completed.connect(_on_questions_loaded)
	
	var url = "http://localhost:8080/api/quiz"
	var headers = ["X-User-ID: " + player.player_id]
	http.request(url, headers, HTTPClient.METHOD_GET)

func _on_questions_loaded(result: int, response_code: int, headers: PackedStringArray, body: PackedByteArray) -> void:
	if result == HTTPRequest.RESULT_SUCCESS and response_code == 200:
		var json = JSON.new()
		var parse_result = json.parse(body.get_string_from_utf8())
		if parse_result == OK:
			var data = json.get_data()
			if data.has("question"):
				questions.append(data["question"])

func spawn_question_orbs() -> void:
	# Создаём орбы с вопросами на уровне
	var orb_scene = preload("res://game/question_orb.tscn")
	
	# Спавним 5 орбов в разных местах
	var positions = [
		Vector2(200, 400),
		Vector2(400, 350),
		Vector2(600, 400),
		Vector2(300, 250),
		Vector2(500, 250)
	]
	
	for i in range(min(5, questions.size())):
		var orb = orb_scene.instantiate()
		orb.position = positions[i]
		orb.question_data = questions[i]
		orb.orb_collected.connect(_on_orb_collected)
		add_child(orb)
		active_orbs.append(orb)

func _on_orb_collected(orb: Node2D, question: Dictionary) -> void:
	# Подбор орба с вопросом
	game_state = "answering"
	player.is_answer_mode = true
	question_manager.show_question(question)
	active_orbs.erase(orb)

func update_exploring_mode() -> void:
	# Игрок исследует мир
	pass

func update_answer_mode() -> void:
	# Игрок отвечает на вопрос
	pass

func setup_ui() -> void:
	# Настройка UI
	if ui:
		ui.player_exp = player.experience
		ui.player_level = player.level
		ui.player_knowledge = player.go_knowledge

func on_question_answered(exp: int, correct: bool) -> void:
	# Обработка ответа на вопрос
	if correct:
		game_state = "exploring"
		player.is_answer_mode = false
		spawn_question_orbs()  # Спавним новые орбы
	else:
		# Даём ещё шанс
		game_state = "exploring"
		player.is_answer_mode = false
