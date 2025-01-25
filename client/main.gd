extends Node

const packets := preload("res://packets.gd")

var client_id: int
@onready var _line_edit := $LineEdit as LineEdit
@onready var logger := $Log as Log

func _on_line_edit_text_submitted(text: String) -> void:
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat()
	chat_msg.set_msg(text)
	
	var err := WS.send(packet)
	if err:
			logger.error("Error sending chat message")
	else:
			logger.chat("You", text)
	_line_edit.text = ""

func _ready() -> void:
	WS.connected_to_server.connect(_on_ws_connected_to_server)
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_received)
	
	_line_edit.text_submitted.connect(_on_line_edit_text_submitted)
	logger.info("Connecting to server...")
	WS.connect_to_url("ws://127.0.0.1:8080/ws")

func _on_ws_connected_to_server() -> void:
	logger.success("Connected successfully")
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat()
	chat_msg.set_msg("Hello, Golang!")
	
	var err := WS.send(packet)
	if err:
		logger.error("Error sending packet")
	else:
		logger.success("Sent packet")
	
func _on_ws_connection_closed() -> void:
	logger.warning("Connection closed")
	
func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_id(): # id is sent when the client connects to the server
			_handle_id_msg(sender_id, packet.get_id())
	elif packet.has_chat():
			_handle_chat_msg(sender_id, packet.get_chat())

func _handle_id_msg(_sender_id: int, id_msg: packets.IdMessage) -> void:
	var client_id := id_msg.get_id()
	logger.info("Received client ID: %d" % client_id)

func _handle_chat_msg(sender_id: int, chat_msg: packets.ChatMessage) -> void:
	logger.chat("Client %d" % sender_id, chat_msg.get_msg())
