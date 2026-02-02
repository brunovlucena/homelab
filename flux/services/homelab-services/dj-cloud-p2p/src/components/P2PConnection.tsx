import { useState, useEffect } from 'react';
import { useStore } from '../store';
import { p2pService } from '../services/p2pService';

export default function P2PConnection() {
  const { p2pConnected, p2pPeerId, setP2PConnected, setP2PPeerId } = useStore();
  const [remotePeerId, setRemotePeerId] = useState('');
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connecting' | 'connected'>('disconnected');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Gerar Peer ID ao montar
    const peerId = p2pService.generatePeerId();
    setP2PPeerId(peerId);

    // Setup event listeners
    p2pService.on('connected', () => {
      setP2PConnected(true);
      setConnectionStatus('connected');
      setError(null);
    });

    p2pService.on('disconnected', () => {
      setP2PConnected(false);
      setConnectionStatus('disconnected');
    });

    p2pService.on('error', (err: Error) => {
      setError(err.message);
      setConnectionStatus('disconnected');
    });

    return () => {
      p2pService.disconnect();
    };
  }, []);

  const handleConnect = async () => {
    if (!remotePeerId.trim()) {
      setError('Por favor, insira um Peer ID');
      return;
    }

    setConnectionStatus('connecting');
    setError(null);

    try {
      await p2pService.connect(remotePeerId);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao conectar');
      setConnectionStatus('disconnected');
    }
  };

  const handleDisconnect = () => {
    p2pService.disconnect();
    setP2PConnected(false);
    setConnectionStatus('disconnected');
    setRemotePeerId('');
  };

  return (
    <div className="space-y-4 max-w-2xl mx-auto">
      <div>
        <h2 className="text-2xl font-bold text-white mb-2">Conexão P2P</h2>
        <p className="text-slate-400">
          Conecte-se a outro dispositivo para fazer streaming de músicas
        </p>
      </div>

      {/* Status Card */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-white">Status da Conexão</h3>
            <p className="text-slate-400 text-sm mt-1">
              {connectionStatus === 'connected' && '✅ Conectado'}
              {connectionStatus === 'connecting' && '🔄 Conectando...'}
              {connectionStatus === 'disconnected' && '❌ Desconectado'}
            </p>
          </div>
          <div className="text-right">
            <p className="text-xs text-slate-400">Seu Peer ID</p>
            <p className="text-sm font-mono text-primary-400 break-all max-w-xs">
              {p2pPeerId || 'Gerando...'}
            </p>
          </div>
        </div>

        {error && (
          <div className="bg-red-900/50 border border-red-700 rounded-lg p-3 mb-4">
            <p className="text-red-200 text-sm">{error}</p>
          </div>
        )}

        {!p2pConnected ? (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Peer ID do Dispositivo Remoto
              </label>
              <input
                type="text"
                value={remotePeerId}
                onChange={(e) => setRemotePeerId(e.target.value)}
                placeholder="Cole o Peer ID aqui"
                className="w-full bg-slate-700 text-white px-4 py-2 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
            <button
              onClick={handleConnect}
              disabled={connectionStatus === 'connecting'}
              className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {connectionStatus === 'connecting' ? 'Conectando...' : 'Conectar'}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="bg-green-900/20 border border-green-700 rounded-lg p-4">
              <p className="text-green-200 text-sm">
                ✅ Conectado com sucesso! Você pode agora fazer streaming de músicas.
              </p>
            </div>
            <button onClick={handleDisconnect} className="btn-secondary w-full">
              Desconectar
            </button>
          </div>
        )}
      </div>

      {/* Instructions */}
      <div className="card bg-slate-800/50">
        <h3 className="text-lg font-semibold text-white mb-3">Como conectar:</h3>
        <ol className="list-decimal list-inside space-y-2 text-slate-300 text-sm">
          <li>Abra o app em outro dispositivo</li>
          <li>Copie o Peer ID do dispositivo remoto</li>
          <li>Cole o Peer ID no campo acima</li>
          <li>Clique em "Conectar"</li>
          <li>Comece a fazer streaming de músicas!</li>
        </ol>
      </div>
    </div>
  );
}
