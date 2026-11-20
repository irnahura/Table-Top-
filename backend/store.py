"""SQLite persistence and event log for local Astralis development."""

import hashlib
import json
import secrets
import sqlite3
import uuid
from datetime import datetime, timezone
from pathlib import Path

from .engine import Actor, World


def now() -> str:
    return datetime.now(timezone.utc).isoformat()


class Store:
    def __init__(self, path: str = "astralis.db"):
        self.db = sqlite3.connect(path, check_same_thread=False)
        self.db.row_factory = sqlite3.Row
        self.db.executescript("""
        PRAGMA foreign_keys = ON;
        CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, display_name TEXT NOT NULL, created_at TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS campaigns (id TEXT PRIMARY KEY, owner_id TEXT NOT NULL, name TEXT NOT NULL, description TEXT NOT NULL, created_at TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS members (campaign_id TEXT NOT NULL, user_id TEXT NOT NULL, role TEXT NOT NULL, PRIMARY KEY (campaign_id,user_id));
        CREATE TABLE IF NOT EXISTS invitations (token TEXT PRIMARY KEY, campaign_id TEXT NOT NULL, creator_id TEXT NOT NULL, expires_at TEXT NOT NULL, used INTEGER NOT NULL DEFAULT 0);
        CREATE TABLE IF NOT EXISTS characters (id TEXT PRIMARY KEY, owner_id TEXT NOT NULL, campaign_id TEXT, name TEXT NOT NULL, class_name TEXT NOT NULL, level INTEGER NOT NULL, hp INTEGER NOT NULL, max_hp INTEGER NOT NULL, x INTEGER NOT NULL, y INTEGER NOT NULL);
        CREATE TABLE IF NOT EXISTS maps (id TEXT PRIMARY KEY, owner_id TEXT NOT NULL, campaign_id TEXT, name TEXT NOT NULL, width INTEGER NOT NULL, height INTEGER NOT NULL, blocked TEXT NOT NULL, created_at TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS sessions (id TEXT PRIMARY KEY, campaign_id TEXT NOT NULL, map_id TEXT NOT NULL, status TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL, sequence INTEGER NOT NULL, created_at TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS events (session_id TEXT NOT NULL, sequence INTEGER NOT NULL, event_type TEXT NOT NULL, actor_id TEXT, payload TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY(session_id,sequence));
        """)
        self.db.commit()

    def _id(self): return str(uuid.uuid4())
    def _hash(self, password): return hashlib.sha256(password.encode()).hexdigest()

    def user(self, user_id): return self.db.execute("SELECT * FROM users WHERE id=?", (user_id,)).fetchone()
    def register(self, email, password, display_name):
        user_id = self._id()
        self.db.execute("INSERT INTO users VALUES (?,?,?,?,?)", (user_id, email.lower(), self._hash(password), display_name, now()))
        self.db.commit()
        return dict(self.user(user_id))
    def authenticate(self, email, password):
        row = self.db.execute("SELECT * FROM users WHERE email=? AND password_hash=?", (email.lower(), self._hash(password))).fetchone()
        return dict(row) if row else None

    def campaign(self, campaign_id): return self.db.execute("SELECT * FROM campaigns WHERE id=?", (campaign_id,)).fetchone()
    def is_member(self, campaign_id, user_id): return self.db.execute("SELECT 1 FROM members WHERE campaign_id=? AND user_id=?", (campaign_id,user_id)).fetchone() is not None
    def create_campaign(self, user_id, name, description):
        campaign_id = self._id(); self.db.execute("INSERT INTO campaigns VALUES (?,?,?,?,?)", (campaign_id,user_id,name,description,now())); self.db.execute("INSERT INTO members VALUES (?,?,?)", (campaign_id,user_id,"owner")); self.db.commit(); return dict(self.campaign(campaign_id))
    def invite(self, campaign_id, creator_id):
        token = secrets.token_urlsafe(24); self.db.execute("INSERT INTO invitations VALUES (?,?,?,?,0)", (token,campaign_id,creator_id,now())); self.db.commit(); return token
    def redeem(self, token, user_id):
        invite = self.db.execute("SELECT * FROM invitations WHERE token=? AND used=0", (token,)).fetchone()
        if not invite: return False
        self.db.execute("INSERT OR REPLACE INTO members VALUES (?,?,?)", (invite["campaign_id"],user_id,"player")); self.db.execute("UPDATE invitations SET used=1 WHERE token=?", (token,)); self.db.commit(); return True
    def create_character(self, user_id, campaign_id, name, class_name):
        char_id=self._id(); self.db.execute("INSERT INTO characters VALUES (?,?,?,?,?,?,?,?,?,?)", (char_id,user_id,campaign_id,name,class_name,1,100,100,2,2)); self.db.commit(); return dict(self.db.execute("SELECT * FROM characters WHERE id=?",(char_id,)).fetchone())
    def create_map(self, owner_id, campaign_id, name, width=20, height=14, blocked=None):
        map_id=self._id(); blocked=blocked or [[7,4],[7,5],[7,6],[8,6],[9,6]]; self.db.execute("INSERT INTO maps VALUES (?,?,?,?,?,?,?,?)", (map_id,owner_id,campaign_id,name,width,height,json.dumps(blocked),now())); self.db.commit(); return dict(self.db.execute("SELECT * FROM maps WHERE id=?",(map_id,)).fetchone())
    def create_session(self, campaign_id, map_id):
        session_id=self._id(); actors=[Actor("mayank","Mayank",2,2,82,100),Actor("lira","Lira",4,3,100,100),Actor("bandit-scout","Bandit Scout",10,5,30,30,hostile=True),Actor("bandit-archer","Bandit Archer",13,4,24,24,hostile=True)]
        state={"width":20,"height":14,"blocked":[[7,4],[7,5],[7,6],[8,6],[9,6]],"actors":[a.__dict__ for a in actors],"active_actor_id":"mayank","round":1}
        self.db.execute("INSERT INTO sessions VALUES (?,?,?,?,?,?,?,?)",(session_id,campaign_id,map_id,"active",json.dumps(state),1,0,now())); self.db.commit(); return session_id
    def session(self, session_id): return self.db.execute("SELECT * FROM sessions WHERE id=?",(session_id,)).fetchone()
    def world(self, session_id):
        row=self.session(session_id); data=json.loads(row["state"]); return World(data["width"],data["height"],frozenset(tuple(x) for x in data["blocked"]),tuple(Actor(**a) for a in data["actors"]),data["active_actor_id"],data["round"])
    def commit(self, session_id, world: World, event_type, actor_id, payload):
        row=self.session(session_id); sequence=row["sequence"]+1; version=row["version"]+1; state={"width":world.width,"height":world.height,"blocked":[list(x) for x in world.blocked],"actors":[a.__dict__ for a in world.actors],"active_actor_id":world.active_actor_id,"round":world.round}
        self.db.execute("UPDATE sessions SET state=?,version=?,sequence=? WHERE id=?",(json.dumps(state),version,sequence,session_id)); self.db.execute("INSERT INTO events VALUES (?,?,?,?,?,?)",(session_id,sequence,event_type,actor_id,json.dumps(payload),now())); self.db.commit(); return sequence,version
    def snapshot(self, session_id):
        row=self.session(session_id); return {"session_id":session_id,"state_version":row["version"],"sequence":row["sequence"],"status":row["status"],"state":json.loads(row["state"])}
    def events(self, session_id, after=0):
        return [dict(x) for x in self.db.execute("SELECT * FROM events WHERE session_id=? AND sequence>? ORDER BY sequence",(session_id,after)).fetchall()]
