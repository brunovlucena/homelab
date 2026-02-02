import React, { useState, useEffect, useRef, useCallback } from 'react';
import ChatbotService from '../services/chatbot';
import { useChatbot } from '../contexts/ChatbotContext';

interface Message {
  id: string;
  text: string;
  isUser: boolean;
  timestamp: Date;
}

interface LLMStatus {
  status: string;
  model?: string;
  provider?: string;
  error?: string;
}

const Chatbot: React.FC = () => {
  const { isOpen, closeChatbot, toggleChatbot } = useChatbot();
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      text: 'How can I help you today?',
      isUser: false,
      timestamp: new Date()
    }
  ]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [llmStatus, setLlmStatus] = useState<LLMStatus | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll to bottom when new messages arrive
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  // Focus input when chat opens
  useEffect(() => {
    if (isOpen) {
      inputRef.current?.focus();
    }
  }, [isOpen]);

  useEffect(() => {
    // Initialize the chatbot service
    void ChatbotService.initialize();
    
    // Check LLM status
    const checkLLMStatus = async () => {
      try {
        const status = await ChatbotService.getLLMStatus();
        setLlmStatus(status);
      } catch (error) {
        console.error('Failed to check LLM status:', error);
        setLlmStatus({ status: 'error', error: 'Failed to check status' });
      }
    };
    
    void checkLLMStatus();
  }, []);

  const handleSendMessage = async () => {
    if (!inputValue.trim() || isLoading) {
      return;
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      text: inputValue,
      isUser: true,
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);

    try {
      const response = await ChatbotService.processMessage(inputValue);
      const botMessage: Message = {
        id: (Date.now() + 1).toString(),
        text: response.text,
        isUser: false,
        timestamp: new Date()
      };
      setMessages(prev => [...prev, botMessage]);
    } catch {
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        text: 'Sorry, I encountered an error. Please try again.',
        isUser: false,
        timestamp: new Date()
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="chatbot-container">
      {/* Chat Button */}
      <button
        className="chatbot-toggle"
        onClick={toggleChatbot}
        aria-label="Toggle chatbot"
      >
        {isOpen ? (
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        ) : (
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
          </svg>
        )}
      </button>

      {/* Chat Window */}
      {isOpen && (
        <div className="chatbot-window">
          <div className="chatbot-header">
            <button
              className="chatbot-close"
              onClick={closeChatbot}
              aria-label="Close chatbot"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <div 
            className="chatbot-messages" 
            role="log" 
            aria-live="polite" 
            aria-label="Chat messages"
          >
            {/* LLM Status indicator */}
            {llmStatus && llmStatus.status !== 'healthy' && (
              <div className="chatbot-status-warning" role="alert">
                <span>⚠️ AI service: {llmStatus.status}</span>
              </div>
            )}
            {messages.map((message) => (
              <div
                key={message.id}
                className={`chatbot-message ${message.isUser ? 'user' : 'bot'}`}
                role="article"
                aria-label={`${message.isUser ? 'You' : 'Assistant'} said`}
              >
                <div className="message-content">
                  {message.text}
                </div>
                <div className="message-time" aria-label={`Sent at ${message.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`}>
                  {message.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </div>
              </div>
            ))}
            {isLoading && (
              <div className="chatbot-message bot" role="status" aria-label="Assistant is typing">
                <div className="message-content">
                  <div className="typing-indicator" aria-hidden="true">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          <div className="chatbot-input">
            <textarea
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Type your message..."
              disabled={isLoading}
              rows={1}
              aria-label="Chat message input"
              aria-describedby="chat-input-hint"
            />
            <span id="chat-input-hint" className="sr-only">
              Press Enter to send, Shift+Enter for new line
            </span>
            <button
              onClick={handleSendMessage}
              disabled={!inputValue.trim() || isLoading}
              className="send-button"
              aria-label="Send message"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22,2 15,22 11,13 2,9"></polygon>
              </svg>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Chatbot;
