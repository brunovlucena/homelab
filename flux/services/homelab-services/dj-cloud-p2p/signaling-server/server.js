const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const cors = require('cors');

const app = express();
const server = http.createServer(app);
const io = new Server(server, {
  cors: {
    origin: '*',
    methods: ['GET', 'POST'],
  },
});

app.use(cors());
app.use(express.json());

// Armazenar peers conectados
const peers = new Map();

io.on('connection', (socket) => {
  console.log('Cliente conectado:', socket.id);

  // Registrar peer
  socket.on('register', (peerId) => {
    peers.set(peerId, socket.id);
    socket.peerId = peerId;
    console.log(`Peer registrado: ${peerId} -> ${socket.id}`);
  });

  // Enviar offer
  socket.on('offer', (data) => {
    const { from, to, offer } = data;
    console.log(`Offer de ${from} para ${to}`);
    
    const targetSocketId = peers.get(to);
    if (targetSocketId) {
      io.to(targetSocketId).emit('offer', { from, offer });
    } else {
      socket.emit('error', { message: `Peer ${to} não encontrado` });
    }
  });

  // Enviar answer
  socket.on('answer', (data) => {
    const { from, to, answer } = data;
    console.log(`Answer de ${from} para ${to}`);
    
    const targetSocketId = peers.get(to);
    if (targetSocketId) {
      io.to(targetSocketId).emit('answer', { from, answer });
    } else {
      socket.emit('error', { message: `Peer ${to} não encontrado` });
    }
  });

  // Enviar ICE candidate
  socket.on('ice-candidate', (data) => {
    const { from, to, candidate } = data;
    console.log(`ICE candidate de ${from} para ${to}`);
    
    const targetSocketId = peers.get(to);
    if (targetSocketId) {
      io.to(targetSocketId).emit('ice-candidate', { from, candidate });
    }
  });

  // Desconexão
  socket.on('disconnect', () => {
    if (socket.peerId) {
      peers.delete(socket.peerId);
      console.log(`Peer desconectado: ${socket.peerId}`);
    }
  });
});

const PORT = process.env.PORT || 3001;
server.listen(PORT, () => {
  console.log(`🚀 Servidor de signaling rodando na porta ${PORT}`);
  console.log(`📡 Aguardando conexões P2P...`);
});
