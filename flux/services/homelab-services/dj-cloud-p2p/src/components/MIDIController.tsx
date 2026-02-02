import { useState, useEffect } from 'react';
import { useStore } from '../store';
import { midiService } from '../services/midiService';
import { isDDJRev5 } from '../services/ddj-rev5-mapping';

export default function MIDIController() {
  const {
    midiConnected,
    midiDeviceName,
    midiDevices,
    setMIDIConnected,
    setMIDIDeviceName,
    setMIDIDevices,
  } = useStore();

  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isScanning, setIsScanning] = useState(false);

  useEffect(() => {
    // Setup event listeners
    midiService.on('connected', (device) => {
      setMIDIConnected(true);
      setMIDIDeviceName(device?.name || null);
      setIsConnecting(false);
      setError(null);
    });

    midiService.on('disconnected', () => {
      setMIDIConnected(false);
      setMIDIDeviceName(null);
      setIsConnecting(false);
    });

    midiService.on('error', (err: Error) => {
      setError(err.message);
      setIsConnecting(false);
    });

    // Scan for devices on mount
    scanDevices();

    return () => {
      midiService.disconnect();
    };
  }, []);

  const scanDevices = async () => {
    setIsScanning(true);
    setError(null);
    try {
      const devices = await midiService.listDevices();
      setMIDIDevices(devices);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao escanear dispositivos');
    } finally {
      setIsScanning(false);
    }
  };

  const handleConnect = async (deviceName?: string) => {
    setIsConnecting(true);
    setError(null);
    try {
      const success = await midiService.connect(deviceName);
      if (!success) {
        setError('Falha ao conectar à controladora');
        setIsConnecting(false);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao conectar');
      setIsConnecting(false);
    }
  };

  const handleDisconnect = async () => {
    try {
      await midiService.disconnect();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao desconectar');
    }
  };

  const ddjRev5Devices = midiDevices.filter((d) => isDDJRev5(d.name));
  const otherDevices = midiDevices.filter((d) => !isDDJRev5(d.name));

  return (
    <div className="space-y-4 max-w-2xl mx-auto">
      <div>
        <h2 className="text-2xl font-bold text-white mb-2">Controladora MIDI</h2>
        <p className="text-slate-400">
          Conecte sua controladora DJ (DDJ-REV5, etc.) para controlar o player
        </p>
      </div>

      {/* Status Card */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-white">Status da Controladora</h3>
            <p className="text-slate-400 text-sm mt-1">
              {midiConnected && midiDeviceName && (
                <>
                  ✅ Conectado: <span className="text-primary-400">{midiDeviceName}</span>
                </>
              )}
              {isConnecting && '🔄 Conectando...'}
              {!midiConnected && !isConnecting && '❌ Desconectado'}
            </p>
          </div>
          <button
            onClick={scanDevices}
            disabled={isScanning}
            className="btn-secondary text-sm disabled:opacity-50"
          >
            {isScanning ? 'Escaneando...' : '🔄 Atualizar'}
          </button>
        </div>

        {error && (
          <div className="bg-red-900/50 border border-red-700 rounded-lg p-3 mb-4">
            <p className="text-red-200 text-sm">{error}</p>
          </div>
        )}

        {!midiConnected ? (
          <div className="space-y-4">
            {/* DDJ-REV5 Devices */}
            {ddjRev5Devices.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold text-white mb-2">
                  🎛️ Controladoras DDJ-REV5
                </h4>
                <div className="space-y-2">
                  {ddjRev5Devices.map((device) => (
                    <button
                      key={device.name}
                      onClick={() => handleConnect(device.name)}
                      disabled={isConnecting}
                      className="w-full text-left bg-slate-700 hover:bg-slate-600 px-4 py-3 rounded-lg transition-colors disabled:opacity-50"
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-white font-medium">{device.name}</p>
                          {device.manufacturer && (
                            <p className="text-slate-400 text-xs">{device.manufacturer}</p>
                          )}
                        </div>
                        <span className="text-primary-400">▶ Conectar</span>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Other MIDI Devices */}
            {otherDevices.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold text-white mb-2">
                  🎹 Outros Dispositivos MIDI
                </h4>
                <div className="space-y-2">
                  {otherDevices.map((device) => (
                    <button
                      key={device.name}
                      onClick={() => handleConnect(device.name)}
                      disabled={isConnecting}
                      className="w-full text-left bg-slate-700 hover:bg-slate-600 px-4 py-3 rounded-lg transition-colors disabled:opacity-50"
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-white font-medium">{device.name}</p>
                          {device.manufacturer && (
                            <p className="text-slate-400 text-xs">{device.manufacturer}</p>
                          )}
                        </div>
                        <span className="text-primary-400">▶ Conectar</span>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {midiDevices.length === 0 && !isScanning && (
              <div className="text-center py-8">
                <p className="text-slate-400 mb-4">Nenhum dispositivo MIDI encontrado</p>
                <p className="text-slate-500 text-sm">
                  Certifique-se de que sua controladora está conectada via USB
                </p>
              </div>
            )}

            {/* Auto-connect DDJ-REV5 button */}
            {ddjRev5Devices.length > 0 && (
              <button
                onClick={() => handleConnect()}
                disabled={isConnecting}
                className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isConnecting ? 'Conectando...' : '🔌 Conectar DDJ-REV5 Automaticamente'}
              </button>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="bg-green-900/20 border border-green-700 rounded-lg p-4">
              <p className="text-green-200 text-sm">
                ✅ Controladora conectada! Use os controles físicos para controlar o player.
              </p>
            </div>
            <button onClick={handleDisconnect} className="btn-secondary w-full">
              Desconectar
            </button>
          </div>
        )}
      </div>

      {/* Instructions */}
      {midiConnected && (
        <div className="card bg-slate-800/50">
          <h3 className="text-lg font-semibold text-white mb-3">Controles Disponíveis:</h3>
          <ul className="list-disc list-inside space-y-1 text-slate-300 text-sm">
            <li><strong>Play/Pause:</strong> Botão Play do Deck 1</li>
            <li><strong>Cue:</strong> Botão Cue para voltar ao início</li>
            <li><strong>Pitch:</strong> Controle de pitch do Deck 1</li>
            <li><strong>Jog Wheel:</strong> Rotacione para navegar na música</li>
            <li><strong>EQ:</strong> Controles de High, Mid, Low</li>
            <li><strong>Load:</strong> Botão Load para carregar próxima música</li>
          </ul>
        </div>
      )}
    </div>
  );
}
