export interface WebSocketMessage {
  type: string;
  client_message_id?: string;
  idempotency_key?: string;
  payload?: any;
  error?: string;
  message_id?: string;
  timestamp?: number;
}

export interface MessagePayload {
  conversation_id: string;
  receiver_id: string;
  content: string;
  type: 'text' | 'image' | 'video' | 'audio' | 'document' | 'system';
  media_url?: string;
  reply_to_message_id?: string;
  timestamp: number;
}

export interface MessageAckPayload {
  message_id: string;
  sequence_number: number;
  status: 'sent' | 'delivered' | 'read';
  timestamp: number;
}
