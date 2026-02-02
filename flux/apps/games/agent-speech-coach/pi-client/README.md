# 🍓 Speech Coach - Raspberry Pi Client

Interface web para Raspberry Pi que se conecta ao Speech Coach Agent no servidor studio.

## 🎯 Características

- **🌐 Interface Web**: Funciona no navegador do Raspberry Pi
- **📸 Câmera USB**: Suporte para câmera USB para reconhecimento facial
- **🎤 Microfone**: Captura de áudio via microfone USB ou GPIO
- **🎨 Interface Amigável**: Design simples e intuitivo para crianças
- **🔄 Temas Customizáveis**: Crianças podem personalizar cores e temas
- **📊 Progresso**: Visualização de progresso e conquistas

## 📋 Requisitos

- Raspberry Pi 4 ou superior
- Raspbian/Raspberry Pi OS
- Câmera USB (opcional, para reconhecimento facial)
- Microfone USB ou conectado via GPIO
- Navegador web (Chromium recomendado)

## 🚀 Instalação

### 1. Instalar dependências

```bash
sudo apt-get update
sudo apt-get install -y python3 python3-pip python3-venv chromium-browser
```

### 2. Configurar ambiente Python

```bash
cd pi-client
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 3. Configurar variáveis de ambiente

```bash
cp .env.example .env
# Editar .env com a URL do servidor studio
```

### 4. Executar

```bash
python3 app.py
```

A interface estará disponível em `http://localhost:8080`

## 🏗️ Arquitetura

```
Raspberry Pi
├── app.py (Flask server)
├── static/
│   ├── css/ (estilos e temas)
│   ├── js/ (cliente CloudEvents)
│   └── images/ (assets)
├── templates/
│   └── index.html (interface principal)
└── camera.py (reconhecimento facial)
```

## 🔌 Conexão com Studio

O cliente se conecta ao mobile-api no cluster studio:

```
http://mobile-api.homelab-services.svc.cluster.local:8080/api/v1/cloudevents
```

Ou via Cloudflare Tunnel (se configurado):
```
https://speech-coach.your-domain.com/api/v1/cloudevents
```

## 📱 Autostart (Opcional)

Para iniciar automaticamente ao ligar o Raspberry Pi:

```bash
# Criar service systemd
sudo cp speech-coach.service /etc/systemd/system/
sudo systemctl enable speech-coach.service
sudo systemctl start speech-coach.service
```

## 🎨 Temas

Os temas podem ser customizados editando os arquivos CSS em `static/css/themes/`.

Temas disponíveis:
- `default.css` - Tema padrão
- `ocean.css` - Tema azul oceano
- `forest.css` - Tema verde floresta
- `sunset.css` - Tema laranja/rosa pôr do sol
- `space.css` - Tema espacial
