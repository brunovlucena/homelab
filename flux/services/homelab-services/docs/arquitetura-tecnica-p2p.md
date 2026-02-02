# Arquitetura Técnica: DJ Cloud P2P

## 🏗️ Visão Geral da Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    DJ Cloud P2P System                       │
└─────────────────────────────────────────────────────────────┘

┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   Device A   │◄───────►│   Device B   │◄───────►│   Device C   │
│  (Home PC)   │  P2P    │  (Mobile)    │  P2P    │  (Laptop)    │
└──────────────┘         └──────────────┘         └──────────────┘
       │                        │                        │
       │                        │                        │
       └────────────────────────┼────────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   Signaling Server      │
                    │   (STUN/TURN/WebSocket) │
                    └─────────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   DHT Bootstrap Nodes   │
                    │   (Peer Discovery)      │
                    └─────────────────────────┘
```

---

## 🔧 Componentes Principais

### 1. Cliente P2P (Peer Client)

**Responsabilidades:**
- Gerenciar conexões WebRTC com outros peers
- Participar da rede DHT para descoberta
- Gerenciar biblioteca local de músicas
- Stream de áudio para outros dispositivos
- Cache local inteligente

**Tecnologias:**
```typescript
// Core P2P
import SimplePeer from 'simple-peer';
import { DHT } from 'dht-rpc';
import WebTorrent from 'webtorrent';

// Audio
import { Howl } from 'howler';
import { extractMetadata } from 'music-metadata';

// Storage
import { IndexedDB } from 'idb';
```

### 2. Signaling Server

**Função:**
- Facilitar conexão inicial entre peers
- NAT traversal (STUN/TURN)
- Descoberta inicial de peers

**Implementação:**
```javascript
// Node.js + Socket.io
const io = require('socket.io')(server);

io.on('connection', (socket) => {
  // Peer A quer conectar com Peer B
  socket.on('offer', (data) => {
    // Encaminhar offer para Peer B
    io.to(data.targetPeerId).emit('offer', data);
  });
  
  socket.on('answer', (data) => {
    // Encaminhar answer para Peer A
    io.to(data.targetPeerId).emit('answer', data);
  });
  
  socket.on('ice-candidate', (data) => {
    // Encaminhar ICE candidates
    io.to(data.targetPeerId).emit('ice-candidate', data);
  });
});
```

**Custos:**
- 3-5 servidores globais
- ~$20-50/servidor/mês
- Total: $60-250/mês

### 3. DHT Network (Distributed Hash Table)

**Função:**
- Descoberta de peers sem servidor central
- Distribuição de metadata
- Content addressing

**Implementação:**
```javascript
import { DHT } from 'dht-rpc';

const dht = new DHT({
  bootstrap: [
    'bootstrap1.djcloudp2p.com:49737',
    'bootstrap2.djcloudp2p.com:49737',
    // ...
  ]
});

// Anunciar presença
dht.announce(Buffer.from(peerId), (err, hash) => {
  console.log('Announced on DHT:', hash.toString('hex'));
});

// Descobrir peers
dht.lookup(hash, (err, peers) => {
  console.log('Found peers:', peers);
});
```

**Bootstrap Nodes:**
- 5-10 nodes iniciais
- ~$10-20/node/mês
- Total: $50-200/mês

### 4. Biblioteca Local

**Estrutura de Dados:**
```typescript
interface Track {
  id: string;              // UUID único
  filePath: string;        // Caminho local
  fileName: string;
  title: string;
  artist: string;
  album: string;
  duration: number;        // segundos
  bpm?: number;
  key?: string;
  waveform?: number[];     // Dados do waveform
  cuePoints?: CuePoint[];
  metadata: {
    bitrate: number;
    sampleRate: number;
    format: string;
    size: number;
  };
  lastModified: Date;
  hash: string;            // SHA-256 para verificação
}

interface Playlist {
  id: string;
  name: string;
  tracks: string[];        // IDs dos tracks
  createdAt: Date;
  updatedAt: Date;
  shared: boolean;         // Se é compartilhada P2P
  collaborators?: string[]; // IDs de outros peers
}
```

**Indexação:**
```typescript
class LibraryManager {
  async scanDirectory(path: string): Promise<Track[]> {
    // Scan recursivo de pastas
    // Extrair metadata de cada arquivo
    // Indexar no IndexedDB
  }
  
  async getTrack(id: string): Promise<Track> {
    // Buscar do IndexedDB
  }
  
  async search(query: string): Promise<Track[]> {
    // Busca full-text local
  }
}
```

---

## 🌐 Protocolo de Comunicação P2P

### Handshake Inicial

```
1. Peer A conecta ao Signaling Server
2. Peer A envia "offer" para Peer B (via Signaling)
3. Peer B recebe "offer" e envia "answer"
4. Ambos trocam ICE candidates
5. Conexão WebRTC estabelecida
6. Comunicação direta (sem servidor)
```

### Streaming de Áudio

**Opção 1: WebRTC DataChannel (Recomendado)**
```typescript
// Enviar chunk de áudio
const dataChannel = peer.createDataChannel('audio');
const audioChunk = await readAudioFile(trackId, offset, length);
dataChannel.send(audioChunk);

// Receber e tocar
dataChannel.onmessage = (event) => {
  const audioChunk = event.data;
  audioBuffer.append(audioChunk);
  playAudio(audioBuffer);
};
```

**Opção 2: WebRTC MediaStream (Alternativa)**
```typescript
// Criar MediaStream do arquivo local
const audioElement = new Audio(trackPath);
const mediaStream = audioElement.captureStream();
peer.addStream(mediaStream);

// Receber e tocar
peer.on('stream', (stream) => {
  const audio = new Audio();
  audio.srcObject = stream;
  audio.play();
});
```

### Sincronização de Biblioteca

**Protocolo Customizado:**
```typescript
interface SyncMessage {
  type: 'sync-request' | 'sync-response' | 'track-update';
  peerId: string;
  tracks?: Track[];
  playlists?: Playlist[];
  timestamp: number;
}

// Peer A solicita sincronização
sendMessage({
  type: 'sync-request',
  peerId: 'peer-a',
  timestamp: Date.now()
});

// Peer B responde com diferenças
sendMessage({
  type: 'sync-response',
  peerId: 'peer-b',
  tracks: newTracks,
  playlists: updatedPlaylists,
  timestamp: Date.now()
});
```

---

## 💾 Armazenamento

### Local (IndexedDB)

**Estrutura:**
```typescript
// Database: djcloudp2p
// Stores:
//   - tracks: Track[]
//   - playlists: Playlist[]
//   - cache: { key: string, data: Blob, expires: Date }
//   - peers: { id: string, lastSeen: Date, metadata: any }
```

**Cache Inteligente:**
```typescript
class CacheManager {
  async cacheTrack(trackId: string, data: Blob) {
    // Armazenar chunk de áudio
    // LRU eviction policy
    // Limite de espaço (ex: 5GB)
  }
  
  async getCachedTrack(trackId: string): Promise<Blob | null> {
    // Verificar se está em cache
    // Retornar se disponível
  }
}
```

### Distribuído (P2P)

**Backup entre Peers:**
```typescript
// Compartilhar metadata (não os arquivos de áudio)
// Usuário escolhe quais peers são "trusted"
// Backup automático de playlists e metadata
```

---

## 🎵 Análise de Música

### Análise Local

**BPM Detection:**
```typescript
import { Essentia } from 'essentia.js';

async function detectBPM(audioBuffer: AudioBuffer): Promise<number> {
  const essentia = new Essentia();
  const bpm = essentia.RhythmExtractor2013(audioBuffer);
  return bpm.bpm;
}
```

**Key Detection:**
```typescript
async function detectKey(audioBuffer: AudioBuffer): Promise<string> {
  const essentia = new Essentia();
  const key = essentia.KeyExtractor(audioBuffer);
  return key.key; // Ex: "C major"
}
```

**Waveform:**
```typescript
async function generateWaveform(
  audioBuffer: AudioBuffer,
  width: number = 200
): Promise<number[]> {
  const samples = audioBuffer.getChannelData(0);
  const blockSize = Math.floor(samples.length / width);
  const waveform: number[] = [];
  
  for (let i = 0; i < width; i++) {
    const start = i * blockSize;
    const end = start + blockSize;
    const chunk = samples.slice(start, end);
    const max = Math.max(...chunk.map(Math.abs));
    waveform.push(max);
  }
  
  return waveform;
}
```

### Compartilhamento de Análises

**Opcional - via P2P:**
```typescript
// Análises podem ser compartilhadas entre peers
// Reduz necessidade de re-análise
// Cache distribuído de análises
```

---

## 🔐 Segurança e Privacidade

### Autenticação

**Modelo Simplificado:**
```typescript
// Cada dispositivo gera um par de chaves
import { generateKeyPair } from 'crypto';

const { publicKey, privateKey } = generateKeyPairSync('rsa', {
  modulusLength: 2048,
});

// Peer ID = hash da public key
const peerId = sha256(publicKey);
```

### Criptografia

**Comunicação:**
- WebRTC já criptografa automaticamente (DTLS)
- Dados sensíveis podem ser criptografados adicionalmente

**Armazenamento:**
- Metadata pode ser criptografada localmente
- Arquivos de áudio não são compartilhados (apenas streaming)

### Privacidade

- **Sem tracking**: Não coletamos dados de uso
- **Dados locais**: Tudo fica no dispositivo do usuário
- **P2P direto**: Sem servidores intermediários após conexão
- **Opt-in**: Usuário escolhe o que compartilhar

---

## 📱 Aplicativo Mobile

### React Native

**Estrutura:**
```typescript
// Componentes principais
- LibraryScreen: Lista de músicas
- PlayerScreen: Player de áudio
- SettingsScreen: Configurações
- P2PConnectionScreen: Status de conexão
```

**P2P no Mobile:**
```typescript
import { RTCPeerConnection } from 'react-native-webrtc';

// Similar ao desktop, mas com adaptações mobile
// - Gerenciamento de bateria
// - Otimização de rede (WiFi vs. dados)
// - Cache mais agressivo
```

---

## 🚀 Otimizações

### Performance

1. **Lazy Loading**
   - Carregar tracks sob demanda
   - Não carregar toda biblioteca na memória

2. **Compressão**
   - Comprimir metadata antes de enviar
   - Codec adaptativo para streaming

3. **Prefetching**
   - Pre-carregar próximas músicas
   - Cache inteligente baseado em uso

### Escalabilidade

1. **DHT Distribuído**
   - Quanto mais peers, melhor funciona
   - Sem gargalos centralizados

2. **Caching Hierárquico**
   - Cache local → Cache de peer próximo → Stream direto

3. **Load Balancing**
   - Múltiplos peers podem servir o mesmo conteúdo
   - Redundância automática

---

## 🧪 Testes

### Testes Unitários

```typescript
// Exemplo: Teste de sincronização
describe('LibrarySync', () => {
  it('should sync tracks between peers', async () => {
    const peerA = new PeerClient('peer-a');
    const peerB = new PeerClient('peer-b');
    
    await peerA.connect(peerB.id);
    await peerA.syncLibrary();
    
    expect(peerB.getTracks()).toEqual(peerA.getTracks());
  });
});
```

### Testes de Integração

- Testar conexão P2P em diferentes redes
- Testar NAT traversal
- Testar streaming com latência variável
- Testar sincronização com múltiplos dispositivos

### Testes de Performance

- Latência de streaming < 200ms
- Uso de memória < 500MB
- CPU usage < 20% durante streaming
- Bateria mobile: < 10%/hora

---

## 📊 Monitoramento

### Métricas a Coletar (Opcional - Privacy First)

**Localmente (não enviado):**
- Número de tracks na biblioteca
- Tempo de uso
- Dispositivos conectados
- Erros e crashes

**Não Coletamos:**
- Quais músicas você tem
- Onde você está
- Informações pessoais
- Dados de uso detalhados

---

## 🔄 Roadmap Técnico

### MVP (Mês 1-4)
- [x] Conexão WebRTC básica
- [ ] Streaming de áudio
- [ ] Biblioteca local
- [ ] Interface básica

### Core (Mês 5-8)
- [ ] Sincronização multi-dispositivo
- [ ] Análise de música
- [ ] Playlists
- [ ] Cache inteligente

### Avançado (Mês 9+)
- [ ] Backup distribuído
- [ ] Playlists colaborativas
- [ ] App mobile
- [ ] Integrações

---

**Documento Técnico v1.0**
**Última atualização:** 2025-01-27
