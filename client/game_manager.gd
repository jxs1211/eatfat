extends Node

enum State {
	ENTERED,
	INGAME,
}

var _states_scenes: Dictionary = {
	State.ENTERED: "res://states/entered/entered.tscn",
	State.INGAME: "res://states/ingame/ingame.tscn",
}

var client_id: int
var _current_scene_root: Node

func set_state(state: State) -> void:
	if _current_scene_root != null:
		_current_scene_root.queue_free()

	var scene_path: String = _states_scenes[state]
	if scene_path == null:
		printerr("Invalid state: %s" % state)
		return
	var scene: PackedScene = load(scene_path)
	_current_scene_root = scene.instantiate()

	add_child(_current_scene_root)
