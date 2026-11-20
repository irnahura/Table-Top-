import random
import pytest
from backend.engine import Actor, World, RuleError, move, attack


def world():
    return World(10, 10, frozenset({(3, 3)}), (Actor("hero", "Hero", 1, 1, 10, 10), Actor("enemy", "Enemy", 3, 1, 5, 5, hostile=True)), "hero")


def test_move_is_server_validated():
    updated, event = move(world(), "hero", (2, 1))
    assert (updated.actors[0].x, updated.actors[0].y) == (2, 1)
    with pytest.raises(RuleError) as error: move(world(), "hero", (3, 3))
    assert error.value.code == "BLOCKED_PATH"


def test_attack_uses_injected_rng():
    updated, event = attack(world(), "hero", "enemy", random.Random(1))
    assert event["roll"] == 5
    assert event["hit"] is False
    assert updated.actors[1].hp == 5
