import { WebSocketMessage, MessagePayload } from '../types/websocket';

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private messageHandlers: Map<string, (data: any) => void> = new Map();
  private isAuthenticated = false;

  constructor(
    private wsUrl: string,
    private userId: string,
    private authToken: string
  ) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.wsUrl);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.authenticate();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('WebSocket closed');
          this.isAuthenticated = false;
          this.attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  private authenticate() {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

    const authMessage: WebSocketMessage = {
      type: 'auth',
      payload: {
        user_id: this.userId,
        auth_token: this.authToken,
        device_id: this.getDeviceId(),
        platform: 'web',
        app_version: '1.0.0',
      },
    };

    this.ws.send(JSON.stringify(authMessage));
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case 'auth_success':
        this.isAuthenticated = true;
        console.log('Authentication successful');
        this.emit('auth_success', message.payload);
        break;

      case 'auth_error':
        console.error('Authentication error:', message.error);
        this.emit('auth_error', { error: message.error });
        break;

      case 'message':
        this.emit('message', message.payload);
        break;

      case 'message_ack':
        this.emit('message_ack', message.payload);
        break;

      case 'delivery_ack':
        this.emit('delivery_ack', { message_id: message.message_id });
        break;

      default:
        console.log('Unhandled message type:', message.type);
    }
  }

  sendMessage(
    conversationId: string,
    receiverId: string,
    content: string,
    messageType: string = 'text'
  ): Promise<string> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket not connected'));
        return;
      }

      if (!this.isAuthenticated) {
        reject(new Error('Not authenticated'));
        return;
      }

      const clientMessageId = `msg-${Date.now()}-${Math.random()}`;
      const idempotencyKey = `key-${Date.now()}-${Math.random()}`;

      const message: WebSocketMessage = {
        type: 'message',
        client_message_id: clientMessageId,
        idempotency_key: idempotencyKey,
        payload: {
          conversation_id: conversationId,
          receiver_id: receiverId,
          content,
          type: messageType as any,
          timestamp: Date.now(),
        },
      };

      this.ws.send(JSON.stringify(message));
      resolve(clientMessageId);
    });
  }

  on(event: string, handler: (data: any) => void) {
    this.messageHandlers.set(event, handler);
  }

  off(event: string) {
    this.messageHandlers.delete(event);
  }

  private emit(event: string, data: any) {
    const handler = this.messageHandlers.get(event);
    if (handler) {
      handler(data);
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    setTimeout(() => {
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      this.connect().catch((error) => {
        console.error('Reconnection failed:', error);
      });
    }, delay);
  }

  private getDeviceId(): string {
    let deviceId = localStorage.getItem('device_id');
    if (!deviceId) {
      deviceId = `web-${Date.now()}-${Math.random()}`;
      localStorage.setItem('device_id', deviceId);
    }
    return deviceId;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isAuthenticated = false;
    }
  }
}
