'use client';

import React, { useState, useEffect, useRef } from 'react';
import { WebSocketService } from '../services/websocket';
import { medicalService } from '../services/api';

interface Message {
  id: string;
  content: string;
  sender: 'doctor' | 'agent' | 'system';
  timestamp: Date;
  type?: 'text' | 'system';
}

interface ChatInterfaceProps {
  doctorId: string;
  conversationId: string;
  wsUrl: string;
  authToken: string;
}

export default function ChatInterface({
  doctorId,
  conversationId,
  wsUrl,
  authToken,
}: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const wsServiceRef = useRef<WebSocketService | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocketService(wsUrl, doctorId, authToken);
    wsServiceRef.current = ws;

    ws.on('auth_success', () => {
      setIsConnected(true);
      console.log('Connected to agent-medical');
    });

    ws.on('message', (payload: any) => {
      const message: Message = {
        id: `msg-${Date.now()}`,
        content: payload.content,
        sender: payload.sender_id === 'system' ? 'system' : 'agent',
        timestamp: new Date(payload.timestamp || Date.now()),
        type: payload.message_type || 'text',
      };
      setMessages((prev) => [...prev, message]);
    });

    ws.on('auth_error', (data: any) => {
      console.error('Authentication error:', data.error);
      setIsConnected(false);
    });

    ws.connect().catch((error) => {
      console.error('Failed to connect:', error);
      setIsConnected(false);
    });

    return () => {
      ws.disconnect();
    };
  }, [doctorId, wsUrl, authToken]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || !isConnected || !wsServiceRef.current) return;

    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      content: input,
      sender: 'doctor',
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      // Send via WebSocket
      await wsServiceRef.current.sendMessage(
        conversationId,
        'agent-medical',
        input
      );

      // Check if it's a special command
      if (input.toLowerCase().startsWith('/summarize')) {
        const patientId = input.split(' ')[1] || '';
        if (patientId) {
          const summary = await medicalService.summarizeCase(
            doctorId,
            patientId
          );
          const summaryMessage: Message = {
            id: `msg-${Date.now()}`,
            content: `📋 Case Summary:\n\n${summary.summary}`,
            sender: 'agent',
            timestamp: new Date(),
          };
          setMessages((prev) => [...prev, summaryMessage]);
        }
      } else if (input.toLowerCase().startsWith('/correlate')) {
        const parts = input.split(' ');
        const patientId = parts[1] || '';
        const query = parts.slice(2).join(' ') || 'all data';
        if (patientId) {
          const correlation = await medicalService.correlateData(
            doctorId,
            patientId,
            query
          );
          const correlationMessage: Message = {
            id: `msg-${Date.now()}`,
            content: `🔍 Correlation Analysis:\n\n${JSON.stringify(correlation.insights, null, 2)}`,
            sender: 'agent',
            timestamp: new Date(),
          };
          setMessages((prev) => [...prev, correlationMessage]);
        }
      }
    } catch (error) {
      console.error('Error sending message:', error);
      const errorMessage: Message = {
        id: `msg-${Date.now()}`,
        content: 'Error sending message. Please try again.',
        sender: 'system',
        timestamp: new Date(),
        type: 'system',
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col h-full bg-white rounded-lg shadow-lg">
      {/* Header */}
      <div className="p-4 border-b bg-teal-600 text-white rounded-t-lg">
        <h2 className="text-lg font-semibold">Agent-Medical Assistant</h2>
        <div className="flex items-center mt-1">
          <div
            className={`w-2 h-2 rounded-full mr-2 ${
              isConnected ? 'bg-green-400' : 'bg-red-400'
            }`}
          />
          <span className="text-sm">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 mt-8">
            <p>Start a conversation with your AI medical assistant</p>
            <p className="text-sm mt-2">
              Try: /summarize patient-123 or /correlate patient-123 lab results
            </p>
          </div>
        )}

        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${
              message.sender === 'doctor' ? 'justify-end' : 'justify-start'
            }`}
          >
            <div
              className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                message.sender === 'doctor'
                  ? 'bg-teal-600 text-white'
                  : message.sender === 'system'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-gray-100 text-gray-800'
              }`}
            >
              <p className="whitespace-pre-wrap">{message.content}</p>
              <p className="text-xs mt-1 opacity-70">
                {message.timestamp.toLocaleTimeString()}
              </p>
            </div>
          </div>
        ))}
        {isLoading && (
          <div className="flex justify-start">
            <div className="bg-gray-100 px-4 py-2 rounded-lg">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-75" />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-150" />
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t">
        <div className="flex space-x-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type your message or use /summarize or /correlate..."
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-teal-500"
            disabled={!isConnected || isLoading}
          />
          <button
            onClick={handleSend}
            disabled={!isConnected || isLoading || !input.trim()}
            className="px-6 py-2 bg-teal-600 text-white rounded-lg hover:bg-teal-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send
          </button>
        </div>
      </div>
    </div>
  );
}
