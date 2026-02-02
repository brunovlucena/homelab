import SimplePeer from 'simple-peer';
import { io, Socket } from 'socket.io-client';
import { v4 as uuidv4 } from 'uuid';
import { EventEmitter } from 'events';

class P2PService extends EventEmitter {
  private peer: SimplePeer.Instance | null = null;
  private socket: Socket | null = null;
  private peerId: string;
  private remotePeerId: string | null = null;
  private signalingUrl: string;

  constructor() {
    super();
    this.peerId = this.generatePeerId();
    // Por enquanto, vamos usar um servidor de signaling local
    // Em produção, isso seria configurável
    this.signalingUrl = process.env.VITE_SIGNALING_URL || 'http://localhost:3001';
  }

  generatePeerId(): string {
    // Gerar um ID único para este peer
    return uuidv4();
  }

  getPeerId(): string {
    return this.peerId;
  }

  async connect(remotePeerId: string): Promise<void> {
    this.remotePeerId = remotePeerId;

    // Conectar ao servidor de signaling
    this.socket = io(this.signalingUrl);

    this.socket.on('connect', () => {
      console.log('Conectado ao servidor de signaling');
      this.socket?.emit('register', this.peerId);
    });

    this.socket.on('disconnect', () => {
      console.log('Desconectado do servidor de signaling');
      this.emit('disconnected');
    });

    // Criar peer como initiator
    this.peer = new SimplePeer({
      initiator: true,
      trickle: false,
    });

    this.setupPeerEvents();

    // Enviar offer quando criado
    this.peer.on('signal', (data) => {
      this.socket?.emit('offer', {
        from: this.peerId,
        to: remotePeerId,
        offer: data,
      });
    });

    // Aguardar answer
    this.socket.on('answer', (data: { from: string; answer: any }) => {
      if (data.from === remotePeerId && this.peer) {
        this.peer.signal(data.answer);
      }
    });

    // Aguardar ICE candidates
    this.socket.on('ice-candidate', (data: { from: string; candidate: any }) => {
      if (data.from === remotePeerId && this.peer) {
        this.peer.signal(data.candidate);
      }
    });
  }

  async acceptConnection(remotePeerId: string): Promise<void> {
    this.remotePeerId = remotePeerId;

    // Conectar ao servidor de signaling
    this.socket = io(this.signalingUrl);

    this.socket.on('connect', () => {
      console.log('Conectado ao servidor de signaling');
      this.socket?.emit('register', this.peerId);
    });

    // Criar peer como receiver
    this.peer = new SimplePeer({
      initiator: false,
      trickle: false,
    });

    this.setupPeerEvents();

    // Enviar answer quando receber offer
    this.socket.on('offer', (data: { from: string; offer: any }) => {
      if (data.from === remotePeerId && this.peer) {
        this.peer.signal(data.offer);

        this.peer.on('signal', (answerData) => {
          this.socket?.emit('answer', {
            from: this.peerId,
            to: remotePeerId,
            answer: answerData,
          });
        });
      }
    });

    // Enviar ICE candidates
    this.peer.on('signal', (data) => {
      this.socket?.emit('ice-candidate', {
        from: this.peerId,
        to: remotePeerId,
        candidate: data,
      });
    });
  }

  private setupPeerEvents(): void {
    if (!this.peer) return;

    this.peer.on('connect', () => {
      console.log('Conexão P2P estabelecida!');
      this.emit('connected');
    });

    this.peer.on('close', () => {
      console.log('Conexão P2P fechada');
      this.emit('disconnected');
    });

    this.peer.on('error', (err) => {
      console.error('Erro P2P:', err);
      this.emit('error', err);
    });

    this.peer.on('data', (data) => {
      try {
        const message = JSON.parse(data.toString());
        this.emit('message', message);
      } catch (error) {
        console.error('Erro ao processar mensagem:', error);
      }
    });
  }

  sendMessage(type: string, data: any): void {
    if (!this.peer || !this.peer.connected) {
      console.error('Peer não está conectado');
      return;
    }

    this.peer.send(
      JSON.stringify({
        type,
        data,
        timestamp: Date.now(),
      })
    );
  }

  disconnect(): void {
    if (this.peer) {
      this.peer.destroy();
      this.peer = null;
    }

    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }

    this.remotePeerId = null;
    this.emit('disconnected');
  }

  isConnected(): boolean {
    return this.peer?.connected || false;
  }
}

export const p2pService = new P2PService();
