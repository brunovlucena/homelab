const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const fs = require('fs');
const { midiHandler } = require('./midi-handler');

let mainWindow: typeof BrowserWindow.prototype | null = null;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
    },
    titleBarStyle: 'hiddenInset',
    backgroundColor: '#0f172a',
  });

  // Carregar app
  if (process.env.NODE_ENV === 'development') {
    mainWindow.loadURL('http://localhost:5173');
    mainWindow.webContents.openDevTools();
  } else {
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// IPC Handlers
ipcMain.handle('get-app-version', () => {
  return app.getVersion();
});

ipcMain.handle('select-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow!, {
    properties: ['openDirectory'],
  });
  
  if (result.canceled) {
    return null;
  }
  
  return result.filePaths[0];
});

ipcMain.handle('read-dir', async (_: any, dirPath: string) => {
  try {
    return fs.readdirSync(dirPath);
  } catch (error) {
    console.error('Erro ao ler diretório:', error);
    return [];
  }
});

ipcMain.handle('get-file-stats', async (_: any, filePath: string) => {
  try {
    const stats = fs.statSync(filePath);
    return {
      size: stats.size,
      mtime: stats.mtime,
      isDirectory: stats.isDirectory(),
    };
  } catch (error) {
    console.error('Erro ao obter stats:', error);
    return null;
  }
});

// MIDI IPC Handlers
ipcMain.handle('midi-list-devices', async () => {
  try {
    return midiHandler.listDevices();
  } catch (error) {
    console.error('Error listing MIDI devices:', error);
    return [];
  }
});

ipcMain.handle('midi-connect', async (_: any, deviceName: string) => {
  try {
    const success = midiHandler.connect(deviceName);
    if (success && mainWindow) {
      // Setup message forwarding
      midiHandler.onMessage((message) => {
        mainWindow?.webContents.send('midi-message', message);
      });
    }
    return success;
  } catch (error) {
    console.error('Error connecting to MIDI device:', error);
    return false;
  }
});

ipcMain.handle('midi-disconnect', async () => {
  try {
    midiHandler.disconnect();
    return true;
  } catch (error) {
    console.error('Error disconnecting from MIDI device:', error);
    return false;
  }
});

ipcMain.handle('midi-send-message', async (_: any, message: any) => {
  try {
    midiHandler.sendMessage(message);
    return true;
  } catch (error) {
    console.error('Error sending MIDI message:', error);
    return false;
  }
});
