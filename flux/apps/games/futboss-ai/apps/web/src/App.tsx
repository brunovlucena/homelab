// FutBoss AI - Main App Component
// Author: Bruno Lucena (bruno@lucena.cloud)

import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Team from './pages/Team';
import Market from './pages/Market';
import Match from './pages/Match';
import Wallet from './pages/Wallet';
import Login from './pages/Login';

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="team" element={<Team />} />
        <Route path="market" element={<Market />} />
        <Route path="match" element={<Match />} />
        <Route path="match/:matchId" element={<Match />} />
        <Route path="wallet" element={<Wallet />} />
      </Route>
    </Routes>
  );
}

export default App;

