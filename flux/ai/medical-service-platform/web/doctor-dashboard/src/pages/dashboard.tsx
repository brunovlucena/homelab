'use client';

import React, { useState, useEffect } from 'react';
import ChatInterface from '../components/ChatInterface';
import { medicalService } from '../services/api';

export default function Dashboard() {
  const [doctorId, setDoctorId] = useState<string>('');
  const [conversationId, setConversationId] = useState<string>('');
  const [authToken, setAuthToken] = useState<string>('');
  const [wsUrl, setWsUrl] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'chat' | 'cases' | 'patients'>('chat');

  useEffect(() => {
    // Get from localStorage or API
    const storedDoctorId = localStorage.getItem('doctor_id');
    const storedToken = localStorage.getItem('auth_token');
    const storedConvId = localStorage.getItem('conversation_id');

    if (storedDoctorId) setDoctorId(storedDoctorId);
    if (storedToken) setAuthToken(storedToken);
    if (storedConvId) setConversationId(storedConvId);

    // WebSocket URL from env or default
    setWsUrl(
      process.env.NEXT_PUBLIC_WS_URL ||
        'ws://localhost:8080/ws'
    );
  }, []);

  if (!doctorId || !authToken) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Please log in</h1>
          <p className="text-gray-600">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-teal-600 text-white shadow-lg">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold">Medical Service Platform</h1>
            <div className="flex items-center space-x-4">
              <span className="text-sm">Doctor ID: {doctorId}</span>
              <button
                onClick={() => {
                  localStorage.removeItem('doctor_id');
                  localStorage.removeItem('auth_token');
                  window.location.href = '/login';
                }}
                className="px-4 py-2 bg-teal-700 rounded hover:bg-teal-800"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Tabs */}
        <div className="mb-6 border-b border-gray-200">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('chat')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'chat'
                  ? 'border-teal-500 text-teal-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              Chat with Agent
            </button>
            <button
              onClick={() => setActiveTab('cases')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'cases'
                  ? 'border-teal-500 text-teal-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              Case Summaries
            </button>
            <button
              onClick={() => setActiveTab('patients')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'patients'
                  ? 'border-teal-500 text-teal-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              Patient Records
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        <div className="h-[calc(100vh-250px)]">
          {activeTab === 'chat' && (
            <ChatInterface
              doctorId={doctorId}
              conversationId={conversationId || `doctor-${doctorId}-agent-medical`}
              wsUrl={wsUrl}
              authToken={authToken}
            />
          )}

          {activeTab === 'cases' && (
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-4">Case Summaries</h2>
              <p className="text-gray-600">
                Use the chat interface to request case summaries with:
                <code className="block mt-2 p-2 bg-gray-100 rounded">
                  /summarize patient-123
                </code>
              </p>
            </div>
          )}

          {activeTab === 'patients' && (
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-4">Patient Records</h2>
              <p className="text-gray-600">
                Ask your agent to show patient records:
                <code className="block mt-2 p-2 bg-gray-100 rounded">
                  Show me patient-123's lab results
                </code>
              </p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
