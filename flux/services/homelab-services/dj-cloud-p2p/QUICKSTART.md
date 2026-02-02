# 🚀 Quick Start - DJ Cloud P2P

## Pré-requisitos

- Node.js 18+ instalado
- npm ou yarn

## Instalação

```bash
# 1. Instalar dependências do app principal
npm install

# 2. Instalar dependências do servidor de signaling
cd signaling-server
npm install
cd ..
```

## Como Rodar

### 1. Iniciar o Servidor de Signaling (Terminal 1)

```bash
cd signaling-server
npm start
```

O servidor irá rodar em `http://localhost:3001`

### 2. Iniciar o App (Terminal 2)

```bash
# No diretório raiz do projeto
npm run dev
```

Isso irá:
- Iniciar o Vite dev server (React)
- Compilar o Electron
- Abrir a janela do app

## Estrutura do Projeto

```
dj-cloud-p2p/
├── electron/          # Código do Electron (main process)
├── src/              # Código React (renderer process)
│   ├── components/   # Componentes React
│   ├── services/     # Serviços (P2P, Library)
│   └── store.ts      # Estado global (Zustand)
├── signaling-server/ # Servidor de signaling P2P
└── package.json
```

## Funcionalidades Atuais (MVP)

✅ Interface básica
✅ Biblioteca de músicas (com dados mock)
✅ Player de áudio básico
✅ Conexão P2P (WebRTC)
✅ Servidor de signaling

## Próximos Passos

- [ ] Scan real de diretório de músicas
- [ ] Streaming de áudio via P2P
- [ ] Sincronização de biblioteca entre dispositivos
- [ ] Análise de música (BPM, key detection)

## Troubleshooting

### Erro ao conectar P2P
- Certifique-se de que o servidor de signaling está rodando
- Verifique se a porta 3001 está livre

### Erro ao tocar música
- Por enquanto, apenas dados mock estão disponíveis
- O scan real de diretório será implementado em breve

### Build do Electron não funciona
```bash
# Recompilar o Electron
npm run build:electron
```

## Desenvolvimento

O projeto usa:
- **React** para a UI
- **TypeScript** para type safety
- **TailwindCSS** para estilização
- **Zustand** para state management
- **Electron** para desktop app
- **WebRTC** (simple-peer) para P2P
- **Howler.js** para reprodução de áudio
