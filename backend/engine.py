"""Pure, dependency-free rules for the Astralis MVP vertical slice."""

from dataclasses import dataclass, replace
import random


class RuleError(Exception):
    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code
        self.message = message


@dataclass(frozen=True)
class Actor:
    id: str
    name: str
    x: int
    y: int
    hp: int
    max_hp: int
    movement: int = 6
    attack_bonus: int = 5
    damage: int = 8
    hostile: bool = False


@dataclass(frozen=True)
class World:
    width: int
    height: int
    blocked: frozenset[tuple[int, int]]
    actors: tuple[Actor, ...]
    active_actor_id: str
    round: int = 1


def distance(a: tuple[int, int], b: tuple[int, int]) -> int:
    return abs(a[0] - b[0]) + abs(a[1] - b[1])


def actor(world: World, actor_id: str) -> Actor:
    for item in world.actors:
        if item.id == actor_id:
            return item
    raise RuleError("NOT_FOUND", "Actor was not found in this session")


def _replace_actor(world: World, updated: Actor) -> World:
    return replace(world, actors=tuple(updated if a.id == updated.id else a for a in world.actors))


def move(world: World, actor_id: str, target: tuple[int, int]) -> tuple[World, dict]:
    subject = actor(world, actor_id)
    if world.active_actor_id != actor_id:
        raise RuleError("NOT_YOUR_TURN", "Only the active participant can move")
    x, y = target
    if x < 0 or y < 0 or x >= world.width or y >= world.height:
        raise RuleError("OUT_OF_RANGE", "Target is outside the map")
    if (x, y) in world.blocked:
        raise RuleError("BLOCKED_PATH", "That cell is blocked")
    if any(a.id != actor_id and a.x == x and a.y == y for a in world.actors):
        raise RuleError("BLOCKED_PATH", "That cell is occupied")
    cost = distance((subject.x, subject.y), target)
    if cost == 0 or cost > subject.movement:
        raise RuleError("OUT_OF_RANGE", "Target is outside the movement budget")
    updated = replace(subject, x=x, y=y)
    return _replace_actor(world, updated), {"character_id": actor_id, "from": {"x": subject.x, "y": subject.y}, "to": {"x": x, "y": y}, "cost": cost}


def interact(world: World, actor_id: str, object_id: str) -> tuple[World, dict]:
    subject = actor(world, actor_id)
    if object_id != "old-bridge":
        raise RuleError("NOT_FOUND", "That object is not interactable")
    if distance((subject.x, subject.y), (8, 5)) > 2:
        raise RuleError("OUT_OF_RANGE", "Move closer to the old bridge")
    return world, {"character_id": actor_id, "object_id": object_id, "result": "The bridge bears fresh wagon marks."}


def attack(world: World, attacker_id: str, target_id: str, rng: random.Random | None = None) -> tuple[World, dict]:
    rng = rng or random.SystemRandom()
    attacker = actor(world, attacker_id)
    target = actor(world, target_id)
    if world.active_actor_id != attacker_id:
        raise RuleError("NOT_YOUR_TURN", "Only the active participant can attack")
    if not target.hostile or distance((attacker.x, attacker.y), (target.x, target.y)) > 5:
        raise RuleError("OUT_OF_RANGE", "Target is not in attack range")
    roll = rng.randint(1, 20)
    total = roll + attacker.attack_bonus
    hit = total >= 12
    damage = attacker.damage if hit else 0
    remaining = max(0, target.hp - damage)
    updated = replace(target, hp=remaining)
    result = {"attacker_id": attacker_id, "target_id": target_id, "roll": roll, "total": total, "hit": hit, "damage": damage, "remaining_hp": remaining}
    return _replace_actor(world, updated), result


def end_turn(world: World) -> tuple[World, dict]:
    order = [a for a in world.actors if not a.hostile and a.hp > 0] + [a for a in world.actors if a.hostile and a.hp > 0]
    if not order:
        raise RuleError("SESSION_NOT_ACTIVE", "There are no active actors")
    ids = [a.id for a in order]
    index = ids.index(world.active_actor_id) if world.active_actor_id in ids else -1
    next_index = (index + 1) % len(ids)
    next_round = world.round + 1 if next_index == 0 else world.round
    next_id = ids[next_index]
    return replace(world, active_actor_id=next_id, round=next_round), {"active_actor_id": next_id, "round": next_round}
