import { Track } from '../store';
import { v4 as uuidv4 } from 'uuid';
import { parseBlob } from 'music-metadata';

class LibraryService {
  async scanDirectory(dirPath?: string): Promise<Track[]> {
    // Por enquanto, vamos simular com dados de exemplo
    // Em produção, isso escanearia o diretório real
    
    // TODO: Implementar scan real de diretório usando Electron APIs
    // Por enquanto retornamos dados de exemplo para desenvolvimento
    return this.getMockTracks();
  }

  private getMockTracks(): Track[] {
    // Dados de exemplo para desenvolvimento
    return [
      {
        id: uuidv4(),
        filePath: '/mock/path/track1.mp3',
        fileName: 'track1.mp3',
        title: 'Example Track 1',
        artist: 'Example Artist',
        album: 'Example Album',
        duration: 180, // 3 minutos
        bpm: 128,
        key: 'C major',
        metadata: {
          bitrate: 320,
          sampleRate: 44100,
          format: 'mp3',
          size: 7200000, // ~7MB
        },
        lastModified: new Date(),
        hash: 'mock-hash-1',
      },
      {
        id: uuidv4(),
        filePath: '/mock/path/track2.mp3',
        fileName: 'track2.mp3',
        title: 'Example Track 2',
        artist: 'Another Artist',
        album: 'Another Album',
        duration: 240, // 4 minutos
        bpm: 130,
        key: 'A minor',
        metadata: {
          bitrate: 320,
          sampleRate: 44100,
          format: 'mp3',
          size: 9600000, // ~9.6MB
        },
        lastModified: new Date(),
        hash: 'mock-hash-2',
      },
      {
        id: uuidv4(),
        filePath: '/mock/path/track3.mp3',
        fileName: 'track3.mp3',
        title: 'Deep House Vibes',
        artist: 'DJ Producer',
        album: 'Summer Mix 2024',
        duration: 320, // 5:20
        bpm: 125,
        key: 'F# minor',
        metadata: {
          bitrate: 320,
          sampleRate: 44100,
          format: 'mp3',
          size: 12800000, // ~12.8MB
        },
        lastModified: new Date(),
        hash: 'mock-hash-3',
      },
    ];
  }

  // TODO: Implementar scan real quando Electron APIs estiverem disponíveis
  async scanRealDirectory(dirPath: string): Promise<Track[]> {
    // Esta função será implementada usando Electron APIs
    // Por enquanto retorna array vazio
    console.warn('Scan real de diretório ainda não implementado');
    return [];
  }
}

export const libraryService = new LibraryService();
