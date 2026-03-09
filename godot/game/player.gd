extends CharacterBody2D

# Go Quiz Player - Главный герой игры
# Собирает знания (орбы), отвечает на вопросы, избегает ошибок

@export var speed: float = 300.0
@export var jump_velocity: float = -400.0
@export var gravity: float = 980.0

# Игровые переменные
var player_id: String = ""
var experience: int = 0
var level: int = 1
var go_knowledge: int = 0
var combo: int = 0
var max_combo: int = 0

# Ноды
@onready var sprite: Sprite2D = $Sprite2D
@onready var animation_player: AnimationPlayer = $AnimationPlayer
@onready var name_label: Label = $NameLabel
@onready var exp_bar: ProgressBar = $ExpBar

# Состояние
var is_answer_mode: bool = false
var current_question: Dictionary = {}

func _ready() -> void:
	# Инициализация игрока
	player_id = get_player_id()
	load_player_data()
	setup_javascript_bridge()

func _physics_process(delta: float) -> void:
	# Гравитация
	if not is_on_floor():
		velocity.y += gravity * delta
	
	# Прыжок
	if Input.is_action_just_pressed("jump") and is_on_floor():
		velocity.y = jump_velocity
		play_animation("jump")
	
	# Движение
	var direction := Input.get_axis("move_left", "move_right")
	if direction:
		velocity.x = direction * speed
		# Поворот спрайта
		sprite.flip_h = direction < 0
		play_animation("run")
	else:
		velocity.x = move_toward(velocity.x, 0, speed * 0.2)
		play_animation("idle")
	
	move_and_slide()

func play_animation(anim_name: String) -> void:
	if animation_player:
		if animation_player.current_animation != anim_name:
			animation_player.play(anim_name)

func get_player_id() -> String:
	# Получаем user_id из JavaScript (если запущено в браузере)
	if JavaScriptBridge:
		return JavaScriptBridge.eval("localStorage.getItem('goquiz_user_id') || 'godot_player'")
	return "godot_player"

func load_player_data() -> void:
	# Загружаем данные игрока с сервера
	var http = HTTPRequest.new()
	add_child(http)
	http.request_completed.connect(_on_player_data_loaded)
	
	var url = "http://localhost:8080/api/stats"
	var headers = ["X-User-ID: " + player_id]
	http.request(url, headers, HTTPClient.METHOD_GET)

func _on_player_data_loaded(result: int, response_code: int, headers: PackedStringArray, body: PackedByteArray) -> void:
	if result == HTTPRequest.RESULT_SUCCESS and response_code == 200:
		var json = JSON.new()
		var parse_result = json.parse(body.get_string_from_utf8())
		if parse_result == OK:
			var data = json.get_data()
			if data.has("player"):
				var player_data = data["player"]
				experience = player_data.get("experience", 0)
				level = player_data.get("level", 1)
				go_knowledge = player_data.get("go_knowledge", 0)
				update_ui()

func update_ui() -> void:
	if name_label:
		name_label.text = "Ур.%d | EXP: %d" % [level, experience]
	if exp_bar:
		exp_bar.max_value = level * 100
		exp_bar.value = experience % (level * 100)

func setup_javascript_bridge() -> void:
	# Настраиваем мост между Godot и JavaScript
	if JavaScriptBridge:
		JavaScriptBridge.eval("""
			window.godotBridge = {
				onQuestionAnswered: function(exp, correct) {
					console.log('Question answered:', exp, correct);
				},
				onLevelUp: function(newLevel) {
					console.log('Level up!', newLevel);
				}
			};
		""")

func collect_knowledge(amount: int) -> void:
	# Сбор знания (орба)
	go_knowledge = mini(go_knowledge + amount, 100)
	experience += amount * 10
	check_level_up()
	update_ui()
	
	# Визуальный эффект
	create_collect_effect()

func answer_question(question: Dictionary, answer_index: int) -> void:
	# Ответ на вопрос
	is_answer_mode = true
	current_question = question
	
	var http = HTTPRequest.new()
	add_child(http)
	http.request_completed.connect(_on_question_answered.bind(http))
	
	var url = "http://localhost:8080/api/answer"
	var headers = [
		"Content-Type: application/json",
		"X-User-ID: " + player_id
	]
	var body = JSON.stringify({
		"question_id": question.get("id", 0),
		"option_index": answer_index
	})
	
	http.request(url, headers, HTTPClient.METHOD_POST, body)

func _on_question_answered(result: int, response_code: int, headers: PackedStringArray, body: PackedByteArray, http: HTTPRequest) -> void:
	http.queue_free()
	is_answer_mode = false
	
	if result == HTTPRequest.RESULT_SUCCESS and response_code == 200:
		var json = JSON.new()
		var parse_result = json.parse(body.get_string_from_utf8())
		if parse_result == OK:
			var data = json.get_data()
			var correct = data.get("correct", false)
			var exp = data.get("exp", 0)
			var new_level = data.get("new_level", level)
			
			if correct:
				# Правильный ответ
				combo += 1
				if combo > max_combo:
					max_combo = combo
				experience = data.get("new_exp", experience)
				level = new_level
				
				# Эффекты
				create_success_effect()
				
				# Уведомляем JavaScript
				if JavaScriptBridge:
					JavaScriptBridge.eval("window.godotBridge.onQuestionAnswered(%d, true)" % exp)
					
					if data.get("level_up", false):
						JavaScriptBridge.eval("window.godotBridge.onLevelUp(%d)" % new_level)
			else:
				# Неправильный ответ
				combo = 0
				create_error_effect()
				
				if JavaScriptBridge:
					JavaScriptBridge.eval("window.godotBridge.onQuestionAnswered(0, false)")
			
			update_ui()

func check_level_up() -> void:
	# Проверка повышения уровня
	var required_exp = level * 100
	if experience >= required_exp:
		level += 1
		create_level_up_effect()

func create_collect_effect() -> void:
	# Эффект сбора знания
	var particles = GPUParticles2D.new()
	add_child(particles)
	particles.position = position
	particles.emitting = true
	
	# Настройка частиц (упрощённо)
	var material = ParticleProcessMaterial.new()
	material.emission_shape = ParticleProcessMaterial.EMISSION_SHAPE_POINT
	material.lifetime = 1.0
	particles.process_material = material
	
	particles.finished.connect(func(): particles.queue_free())

func create_success_effect() -> void:
	# Эффект правильного ответа
	var label = Label.new()
	label.text = "+%d EXP 🔥" % (10 * (combo + 1))
	label.add_theme_color_override("font_color", Color(0, 1, 0))
	label.position = Vector2(position.x - 50, position.y - 100)
	add_child(label)
	
	var tween = create_tween()
	tween.tween_property(label, "position:y", position.y - 200, 1.0)
	tween.tween_property(label, "modulate:a", 0, 1.0)
	tween.tween_callback(label.queue_free)

func create_error_effect() -> void:
	# Эффект неправильного ответа
	var label = Label.new()
	label.text = "❌ Combo сброшен!"
	label.add_theme_color_override("font_color", Color(1, 0, 0))
	label.position = Vector2(position.x - 80, position.y - 100)
	add_child(label)
	
	var tween = create_tween()
	tween.tween_property(label, "position:y", position.y - 200, 1.0)
	tween.tween_property(label, "modulate:a", 0, 1.0)
	tween.tween_callback(label.queue_free)

func create_level_up_effect() -> void:
	# Эффект повышения уровня
	if JavaScriptBridge:
		JavaScriptBridge.eval("""
			alert('🎉 LEVEL UP! Новый уровень: %d')
		""" % level)
	
	# Полное восстановление
	velocity.y = -200  # Прыжок радости
