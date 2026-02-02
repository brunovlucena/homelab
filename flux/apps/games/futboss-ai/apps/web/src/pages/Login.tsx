// FutBoss AI - Login Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';

export default function Login() {
  const navigate = useNavigate();
  const { setUser } = useStore();
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Mock login - in production this would call the API
    setUser({
      id: '1',
      username: username || 'Player',
      email,
      tokens: 1000,
      createdAt: new Date(),
      isActive: true,
    }, 'mock-jwt-token');
    
    navigate('/');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-dark">
      <div className="card w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-primary mb-2">⚽ FutBoss AI</h1>
          <p className="text-gray-400">Football management with AI agents</p>
        </div>

        <div className="flex mb-6">
          <button
            onClick={() => setIsLogin(true)}
            className={`flex-1 py-2 ${isLogin ? 'border-b-2 border-primary text-primary' : 'text-gray-400'}`}
          >
            Login
          </button>
          <button
            onClick={() => setIsLogin(false)}
            className={`flex-1 py-2 ${!isLogin ? 'border-b-2 border-primary text-primary' : 'text-gray-400'}`}
          >
            Register
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          {!isLogin && (
            <div className="mb-4">
              <label className="block text-sm text-gray-400 mb-2">Username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-dark border border-gray-700 rounded-lg px-4 py-3 focus:border-primary focus:outline-none"
                placeholder="Choose a username"
                required={!isLogin}
              />
            </div>
          )}

          <div className="mb-4">
            <label className="block text-sm text-gray-400 mb-2">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full bg-dark border border-gray-700 rounded-lg px-4 py-3 focus:border-primary focus:outline-none"
              placeholder="your@email.com"
              required
            />
          </div>

          <div className="mb-6">
            <label className="block text-sm text-gray-400 mb-2">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-dark border border-gray-700 rounded-lg px-4 py-3 focus:border-primary focus:outline-none"
              placeholder="••••••••"
              required
            />
          </div>

          <button type="submit" className="btn-primary w-full text-lg py-3">
            {isLogin ? 'Login' : 'Create Account'}
          </button>
        </form>

        <p className="text-center text-gray-400 text-sm mt-6">
          By {isLogin ? 'logging in' : 'registering'}, you agree to our Terms of Service
        </p>

        <div className="text-center text-gray-500 text-xs mt-8">
          Developed by Bruno Lucena (bruno@lucena.cloud)
        </div>
      </div>
    </div>
  );
}

