#!/usr/bin/env python3
"""
Speech Coach - Raspberry Pi Web Client
Interface web para conectar ao Speech Coach Agent no servidor studio
"""
import os
import json
import uuid
from datetime import datetime
from flask import Flask, render_template, request, jsonify, send_from_directory
from flask_cors import CORS
import requests

app = Flask(__name__)
CORS(app)

# Configuration
AGENT_URL = os.getenv(
    "AGENT_URL",
    "http://mobile-api.homelab-services.svc.cluster.local:8080/api/v1/cloudevents"
)
AGENT_ID = "speech-coach-agent"
USER_ID = os.getenv("USER_ID", f"pi-user-{uuid.uuid4().hex[:8]}")
THEME = os.getenv("THEME", "default")

# In-memory storage (persistência pode ser adicionada depois)
conversations = {}
current_conversation_id = None


def send_to_agent(message: str, exercise_type: str = None, session_id: str = None):
    """Enviar mensagem para o agente via CloudEvents"""
    conversation_id = current_conversation_id or str(uuid.uuid4())
    
    event = {
        "specversion": "1.0",
        "type": "agent.message",
        "source": "pi-client/speech-coach",
        "id": str(uuid.uuid4()),
        "time": datetime.utcnow().isoformat() + "Z",
        "datacontenttype": "application/json",
        "data": {
            "conversationId": conversation_id,
            "agentId": AGENT_ID,
            "userId": USER_ID,
            "content": message,
            "timestamp": datetime.utcnow().isoformat(),
            "exercise_type": exercise_type,
            "session_id": session_id,
        }
    }
    
    try:
        # Enviar CloudEvent
        response = requests.post(
            AGENT_URL,
            json=event,
            headers={
                "Content-Type": "application/json",
                "ce-specversion": "1.0",
                "ce-type": "agent.message",
                "ce-source": "pi-client/speech-coach",
                "ce-id": event["id"],
                "ce-time": event["time"],
            },
            timeout=30
        )
        response.raise_for_status()
        
        # Processar resposta
        result = response.json()
        
        # Extrair resposta do CloudEvent
        if isinstance(result, dict) and "data" in result:
            response_data = result["data"].get("response", {})
            if isinstance(response_data, dict):
                return {
                    "response": response_data.get("response", ""),
                    "exercise": response_data.get("exercise"),
                    "progress": response_data.get("progress"),
                    "suggestions": response_data.get("suggestions", []),
                }
            return {"response": response_data if isinstance(response_data, str) else str(response_data)}
        
        return {"response": "Resposta recebida", "raw": result}
        
    except requests.exceptions.RequestException as e:
        return {"error": f"Erro ao conectar com o agente: {str(e)}"}


@app.route("/")
def index():
    """Página principal"""
    return render_template("index.html", theme=THEME, agent_url=AGENT_URL)


@app.route("/api/message", methods=["POST"])
def send_message():
    """Enviar mensagem para o agente"""
    global current_conversation_id
    
    data = request.json
    message = data.get("message", "")
    exercise_type = data.get("exercise_type")
    session_id = data.get("session_id")
    
    if not message:
        return jsonify({"error": "Mensagem é obrigatória"}), 400
    
    # Criar conversa se não existir
    if not current_conversation_id:
        current_conversation_id = str(uuid.uuid4())
        conversations[current_conversation_id] = {
            "id": current_conversation_id,
            "messages": [],
            "created_at": datetime.utcnow().isoformat(),
        }
    
    # Adicionar mensagem do usuário
    user_message = {
        "id": str(uuid.uuid4()),
        "content": message,
        "is_from_user": True,
        "timestamp": datetime.utcnow().isoformat(),
    }
    conversations[current_conversation_id]["messages"].append(user_message)
    
    # Enviar para o agente
    result = send_to_agent(message, exercise_type, session_id)
    
    # Adicionar resposta do agente
    if "error" not in result:
        agent_message = {
            "id": str(uuid.uuid4()),
            "content": result.get("response", "Sem resposta"),
            "is_from_user": False,
            "timestamp": datetime.utcnow().isoformat(),
            "exercise": result.get("exercise"),
            "progress": result.get("progress"),
            "suggestions": result.get("suggestions", []),
        }
        conversations[current_conversation_id]["messages"].append(agent_message)
    
    return jsonify(result)


@app.route("/api/conversations", methods=["GET"])
def get_conversations():
    """Listar conversas"""
    return jsonify(list(conversations.values()))


@app.route("/api/conversations/<conversation_id>", methods=["GET"])
def get_conversation(conversation_id):
    """Obter conversa específica"""
    if conversation_id in conversations:
        return jsonify(conversations[conversation_id])
    return jsonify({"error": "Conversa não encontrada"}), 404


@app.route("/api/health", methods=["GET"])
def health():
    """Health check"""
    try:
        # Testar conexão com agente
        test_response = requests.get(AGENT_URL.replace("/cloudevents", "/health"), timeout=5)
        agent_ok = test_response.status_code == 200
    except:
        agent_ok = False
    
    return jsonify({
        "status": "healthy",
        "agent_connected": agent_ok,
        "agent_url": AGENT_URL,
        "user_id": USER_ID,
    })


@app.route("/static/<path:path>")
def serve_static(path):
    """Servir arquivos estáticos"""
    return send_from_directory("static", path)


if __name__ == "__main__":
    port = int(os.getenv("PORT", 8080))
    host = os.getenv("HOST", "0.0.0.0")
    
    print(f"🚀 Speech Coach Pi Client iniciando...")
    print(f"   Agent URL: {AGENT_URL}")
    print(f"   User ID: {USER_ID}")
    print(f"   Theme: {THEME}")
    print(f"   Acesse em: http://{host}:{port}")
    
    app.run(host=host, port=port, debug=False)
