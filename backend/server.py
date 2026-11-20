"""Small standard-library HTTP API for the Astralis MVP.

Run with: python -m backend.server
"""

import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse, parse_qs

from .engine import RuleError, move, interact, attack, end_turn
from .store import Store

ROOT = Path(__file__).resolve().parents[1]
WEB = ROOT / "Astralis-Game-Table"
store = Store(os.getenv("ASTRALIS_DB", str(ROOT / "astralis.db")))
tokens = {}


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_): pass
    def send_json(self, status, value):
        body=json.dumps(value).encode(); self.send_response(status); self.send_header("Content-Type","application/json"); self.send_header("Content-Length",str(len(body))); self.send_header("Access-Control-Allow-Origin","*"); self.end_headers(); self.wfile.write(body)
    def body(self): return json.loads(self.rfile.read(int(self.headers.get("Content-Length",0))) or b"{}")
    def user(self):
        token=self.headers.get("Authorization","").removeprefix("Bearer "); return tokens.get(token)
    def do_OPTIONS(self): self.send_response(204); self.send_header("Access-Control-Allow-Origin","*"); self.send_header("Access-Control-Allow-Headers","Content-Type, Authorization"); self.send_header("Access-Control-Allow-Methods","GET, POST, OPTIONS"); self.end_headers()
    def do_GET(self):
        path=urlparse(self.path).path
        if path == "/health": return self.send_json(200,{"status":"ok","service":"astralis-api"})
        if path == "/" or path == "/index.html": return self.file(WEB/"astralis-home-v2.html")
        if path == "/astralis-tabletop.html": return self.file(WEB/"astralis-tabletop-v2.html")
        if path == "/astralis-onboarding-v2.html": return self.file(WEB/"astralis-onboarding-v2.html")
        if path.endswith(".html"): return self.file(WEB/path.lstrip("/"))
        if path.startswith("/design/"): return self.file(WEB/path.removeprefix("/design/"))
        if path.startswith("/assets/") or path.endswith((".jpg", ".png", ".webp", ".mp4", ".svg")): return self.file(WEB/path.lstrip("/"), binary=True)
        if path.startswith("/api/v1/sessions/") and path.endswith("/snapshot"):
            return self.send_json(200,store.snapshot(path.split("/")[4]))
        if path.startswith("/api/v1/sessions/") and "/events" in path:
            parts=path.split("/"); after=int(parse_qs(urlparse(self.path).query).get("after",[0])[0]); return self.send_json(200,{"events":store.events(parts[4],after)})
        self.send_json(404,{"error":"NOT_FOUND"})
    def file(self,path, binary=False):
        if not path.is_file() or WEB not in path.parents: return self.send_error(404)
        data=path.read_bytes(); suffix=path.suffix.lower(); types={".html":"text/html; charset=utf-8",".jpg":"image/jpeg",".png":"image/png",".webp":"image/webp",".mp4":"video/mp4",".svg":"image/svg+xml"}; self.send_response(200); self.send_header("Content-Type",types.get(suffix,"application/octet-stream")); self.send_header("Content-Length",str(len(data))); self.end_headers(); self.wfile.write(data)
    def do_POST(self):
        path=urlparse(self.path).path; data=self.body()
        try:
            if path == "/api/v1/auth/register":
                return self.send_json(201,{"user":store.register(data["email"],data["password"],data.get("display_name",data["email"]))})
            if path == "/api/v1/auth/login":
                user=store.authenticate(data["email"],data["password"])
                if not user: return self.send_json(401,{"error":"AUTH_REQUIRED"})
                import secrets; token=secrets.token_urlsafe(24); tokens[token]=user["id"]; return self.send_json(200,{"token":token,"user":user})
            user_id=self.user()
            if not user_id: return self.send_json(401,{"error":"AUTH_REQUIRED"})
            if path == "/api/v1/campaigns": return self.send_json(201,{"campaign":store.create_campaign(user_id,data["name"],data.get("description",""))})
            if path.startswith("/api/v1/campaigns/") and path.endswith("/invitations"):
                campaign_id=path.split("/")[4]; return self.send_json(201,{"token":store.invite(campaign_id,user_id)})
            if path == "/api/v1/invitations/redeem": return self.send_json(200,{"redeemed":store.redeem(data["token"],user_id)})
            if path == "/api/v1/characters": return self.send_json(201,{"character":store.create_character(user_id,data.get("campaign_id"),data["name"],data.get("class_name","Ranger"))})
            if path == "/api/v1/maps": return self.send_json(201,{"map":store.create_map(user_id,data.get("campaign_id"),data["name"],data.get("width",20),data.get("height",14),data.get("blocked"))})
            if path.startswith("/api/v1/campaigns/") and path.endswith("/sessions"):
                return self.send_json(201,{"session_id":store.create_session(path.split("/")[4],data["map_id"])})
            if path.startswith("/api/v1/sessions/") and path.endswith("/commands"):
                session_id=path.split("/")[4]; command=data["payload"]; world=store.world(session_id); kind=command["type"]; actor_id=command.get("character_id",command.get("attacker_id","mayank"))
                if kind == "character.move": world,payload=move(world,actor_id,(command["target"]["x"],command["target"]["y"])); event="character.moved"
                elif kind == "object.interact": world,payload=interact(world,actor_id,command["object_id"]); event="object.interacted"
                elif kind == "combat.attack": world,payload=attack(world,command.get("attacker_id",actor_id),command["target_id"]); event="attack.resolved"
                elif kind == "turn.end": world,payload=end_turn(world); event="turn.advanced"
                else: return self.send_json(400,{"error":"INVALID_COMMAND"})
                sequence,version=store.commit(session_id,world,event,actor_id,payload); return self.send_json(200,{"status":"accepted","command_id":data.get("command_id"),"sequence_number":sequence,"state_version":version,"event_type":event,"payload":payload,"snapshot":store.snapshot(session_id)})
            self.send_json(404,{"error":"NOT_FOUND"})
        except RuleError as error: self.send_json(409,{"error":error.code,"message":error.message})
        except (KeyError, ValueError, json.JSONDecodeError) as error: self.send_json(400,{"error":"INVALID_COMMAND","message":str(error)})


if __name__ == "__main__":
    port=int(os.getenv("PORT","8000")); print(f"Astralis API running at http://localhost:{port}"); ThreadingHTTPServer(("127.0.0.1",port),Handler).serve_forever()
