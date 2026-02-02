import React, { createContext, useContext, useState, ReactNode } from 'react';

interface ChatbotContextType {
  isOpen: boolean;
  openChatbot: () => void;
  closeChatbot: () => void;
  toggleChatbot: () => void;
}

const ChatbotContext = createContext<ChatbotContextType | undefined>(undefined);

export const useChatbot = () => {
  const context = useContext(ChatbotContext);
  if (context === undefined) {
    throw new Error('useChatbot must be used within a ChatbotProvider');
  }
  return context;
};

interface ChatbotProviderProps {
  children: ReactNode;
}

export const ChatbotProvider: React.FC<ChatbotProviderProps> = ({ children }) => {
  const [isOpen, setIsOpen] = useState(false);

  const openChatbot = () => setIsOpen(true);
  const closeChatbot = () => setIsOpen(false);
  const toggleChatbot = () => setIsOpen(!isOpen);

  const value = {
    isOpen,
    openChatbot,
    closeChatbot,
    toggleChatbot,
  };

  return (
    <ChatbotContext.Provider value={value}>
      {children}
    </ChatbotContext.Provider>
  );
};
