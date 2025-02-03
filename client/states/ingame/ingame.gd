extends Node

const packets := preload("res://packets.gd")

const Actor := preload("res://objects/actor/actor.gd")
const Spore := preload("res://objects/spore/spore.gd")

var _players: Dictionary
var _spores: Dictionary

@onready var _logout_button: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/LogoutButton
@onready var _send_button: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/SendButton
@onready var _line_edit: LineEdit = $UI/MarginContainer/VBoxContainer/HBoxContainer/LineEdit
@onready var _log: Log = $UI/MarginContainer/VBoxContainer/Log
@onready var _hiscores: Hiscores = $UI/MarginContainer/VBoxContainer/Hiscores
@onready var _world: Node2D = $World

func _ready() -> void:
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_received)
	
	_logout_button.pressed.connect(_on_logout_button_pressed)
	_send_button.pressed.connect(_on_send_button_pressed)
	_line_edit.text_submitted.connect(_on_line_edit_text_submitted)

func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	if sender_id in _players:
		var actor: Actor = _players[sender_id]
		_log.chat(actor.actor_name, chat_msg.get_msg())
	
func _on_logout_button_pressed() -> void:
	var packet := packets.Packet.new()
	var disconnect_msg := packet.new_disconnect()
	disconnect_msg.set_reason("logged out")
	WS.send(packet)
	GameManager.set_state(GameManager.State.CONNECTED)
	
func _on_send_button_pressed() -> void:
	_on_line_edit_text_submitted(_line_edit.text)
	
func _on_line_edit_text_submitted(new_text: String) -> void:
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat()
	chat_msg.set_msg(new_text)
	
	var err := WS.send(packet)
	if err:
		_log.error("Error sending chat message")
	else:
		_log.chat("You", new_text)
	_line_edit.clear()
	
	
func _on_ws_connection_closed() -> void:
	_log.warning("Connection closed")

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat():
		_handle_chat_msg(sender_id, packet.get_chat())
	elif packet.has_player():
		_handle_player_msg(sender_id, packet.get_player())
	elif packet.has_spore():
		_handle_spore_msg(sender_id, packet.get_spore())
	elif packet.has_spores_batch():
		_handle_spores_batch_msg(sender_id, packet.get_spores_batch())
	elif packet.has_spore_consumed():
		_handle_spore_consumed_msg(sender_id, packet.get_spore_consumed())
	elif packet.has_disconnect():
		_handle_disconnect_msg(sender_id, packet.get_disconnect())
	
func _handle_player_msg(sender_id: int, player_msg: packets.PlayerMessage) -> void:
	var actor_id := player_msg.get_id()
	var actor_name := player_msg.get_name()
	var x := player_msg.get_x()
	var y := player_msg.get_y()
	var radius := player_msg.get_radius()
	var speed := player_msg.get_speed()
	var color_hex := player_msg.get_color()
	var color := Color.hex(color_hex)
	
	var is_player := actor_id == GameManager.client_id
	
	if actor_id not in _players:
		_add_actor(actor_id, actor_name, x, y, radius, speed, color, is_player)
	else:
		var direction := player_msg.get_direction()
		_update_actor(actor_id, actor_name, x, y, direction, radius, speed, is_player)

func _add_actor(actor_id: int, actor_name: String, x: float, y: float, radius: float, speed: float, color: Color, is_player: bool) -> void:
	# This is a new player, so we need to create a new actor
	var actor := Actor.instatiate(actor_id, actor_name, x, y, radius, speed, color, is_player)
	_world.add_child(actor)
	actor.z_index = 1
	_set_actor_mass(actor, _rad_to_mass(radius))
	_players[actor_id] = actor
	
	if is_player:
		actor.area_entered.connect(_on_player_area_entered)
	
func _update_actor(actor_id: int, actor_name: String, x: float, y: float, direction: float, radius: float, speed: float, is_player: bool) -> void:
	# This is an existing player, so we need to update their position
	var actor: Actor = _players[actor_id]
	
	_set_actor_mass(actor, _rad_to_mass(radius))
	
	var server_position := Vector2(x, y)
	if actor.position.distance_squared_to(server_position) > 100:
		actor.server_position = server_position
	
	if not is_player:
		actor.velocity = speed * Vector2.from_angle(direction)

func _handle_spore_msg(sender_id: int, spore_msg: packets.SporeMessage) -> void:
	var spore_id := spore_msg.get_id()
	var x := spore_msg.get_x()
	var y := spore_msg.get_y()
	var radius := spore_msg.get_radius()
	
	var underneath_player := false
	if GameManager.client_id in _players:
		var player: Actor = _players[GameManager.client_id]
		var player_pos := Vector2(player.position.x, player.position.y)
		var spore_pos := Vector2(x, y)
		underneath_player = player_pos.distance_squared_to(spore_pos) < player.radius * player.radius
	
	if spore_id not in _spores:
		var spore: Spore = Spore.instantiate(spore_id, x, y, radius, underneath_player)
		_world.add_child(spore)
		_spores[spore_id] = spore

func _handle_spores_batch_msg(sender_id: int, spores_batch_msg: packets.SporesBatchMessage) -> void:
	for spore_msg in spores_batch_msg.get_spores():
		_handle_spore_msg(sender_id, spore_msg)
		
func _handle_spore_consumed_msg(sender_id: int, spore_consumed_msg: packets.SporeConsumedMessage) -> void:
	if sender_id in _players:
		var actor: Actor = _players[sender_id]
		var actor_mass := _rad_to_mass(actor.radius)
		
		var spore_id := spore_consumed_msg.get_spore_id()
		if spore_id in _spores:
			var spore: Spore = _spores[spore_id]
			var spore_mass := _rad_to_mass(spore.radius)
			
			_set_actor_mass(actor, actor_mass + spore_mass)
			_remove_spore(spore)
		
func _handle_disconnect_msg(sender_id: int, disconnect_msg: packets.DisconnectMessage) -> void:
	if sender_id in _players:
		var actor: Actor = _players[sender_id]
		var reason := disconnect_msg.get_reason()
		_log.info("%s disconnected because %s" % [actor.actor_name, reason])
		_remove_actor(actor)
		
func _rad_to_mass(radius: float) -> float:
	return radius * radius * PI

func _set_actor_mass(actor: Actor, new_mass: float) -> void:
	actor.radius = sqrt(new_mass / PI)
	_hiscores.set_hiscore(actor.actor_name, roundi(new_mass))

func _on_player_area_entered(area: Area2D) -> void:
	if area is Spore:
		_consume_spore(area as Spore)
	elif area is Actor:
		_collide_actor(area as Actor)
		
func _consume_spore(spore: Spore) -> void:
	if spore.underneath_player:
		return
	
	var packet := packets.Packet.new()
	var spore_consumed_msg := packet.new_spore_consumed()
	spore_consumed_msg.set_spore_id(spore.spore_id)
	WS.send(packet)
	_remove_spore(spore)
	
func _collide_actor(actor: Actor) -> void:
	var player: Actor = _players[GameManager.client_id]
	var player_mass := _rad_to_mass(player.radius)
	var actor_mass := _rad_to_mass(actor.radius)
	
	if player_mass > actor_mass * 1.5:
		_consume_actor(actor)
	
func _consume_actor(actor: Actor) -> void:
	var player: Actor = _players[GameManager.client_id]
	var player_mass := _rad_to_mass(player.radius)
	var actor_mass := _rad_to_mass(actor.radius)
	_set_actor_mass(player, player_mass + actor_mass)
	
	var packet := packets.Packet.new()
	var player_consumed_msg := packet.new_player_consumed()
	player_consumed_msg.set_player_id(actor.actor_id)
	WS.send(packet)
	
	_remove_actor(actor)

func _remove_spore(spore: Spore) -> void:
	_spores.erase(spore.spore_id)
	spore.queue_free()

func _remove_actor(actor: Actor) -> void:
	_players.erase(actor.actor_id)
	actor.queue_free()
	_hiscores.remove_hiscore(actor.actor_name)
