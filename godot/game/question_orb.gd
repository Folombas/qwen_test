extends Area2D

# Question Orb - Орб с вопросом
# Игрок подбирает и отвечает на вопрос

signal orb_collected(orb: Node2D, question: Dictionary)

@export var question_data: Dictionary = {}
@export var float_speed: float = 50.0
@export var float_amplitude: float = 10.0

@onready var sprite: Sprite2D = $Sprite2D
@onready var animation_player: AnimationPlayer = $AnimationPlayer
@onready var label: Label = $Label

var base_y: float = 0.0
var time: float = 0.0

func _ready() -> void:
	base_y = position.y
	setup_visuals()
	
	# Анимация появления
	var tween = create_tween()
	tween.from_property(sprite, "scale", Vector2(0, 0), 0.5)
	tween.set_ease(Tween.EASE_OUT)

func _process(delta: float) -> void:
	# Плавающее движение
	time += delta
	position.y = base_y + sin(time * float_speed) * float_amplitude
	
	# Вращение
	sprite.rotation += delta * 0.5

func setup_visuals() -> void:
	# Настройка внешнего вида
	if label:
		label.text = "?"
	
	if animation_player:
		animation_player.play("glow")

func _on_body_entered(body: Node2D) -> void:
	# Игрок коснулся орба
	if body.is_in_group("player"):
		collect()

func collect() -> void:
	# Сбор орба
	orb_collected.emit(self, question_data)
	
	# Анимация сбора
	var tween = create_tween()
	tween.parallel().tween_property(sprite, "scale", Vector2(1.5, 1.5), 0.2)
	tween.parallel().tween_property(sprite, "modulate:a", 0, 0.2)
	tween.tween_callback(queue_free)

func _on_body_exited(body: Node2D) -> void:
	pass
