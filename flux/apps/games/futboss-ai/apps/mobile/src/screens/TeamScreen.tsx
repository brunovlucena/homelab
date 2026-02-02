// FutBoss AI - Team Screen
// Author: Bruno Lucena (bruno@lucena.cloud)

import React, { useState } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity } from 'react-native';
import type { Player } from '@futboss/shared';
import { calculateOverall } from '@futboss/shared';

const mockPlayers: Player[] = [
  { id: '1', name: 'Roberto Carlos', position: 'GK', nationality: 'Brasil', age: 28, attributes: { speed: 45, strength: 65, stamina: 70, finishing: 20, passing: 65, dribbling: 35, defense: 85, intelligence: 70, aggression: 40, leadership: 60, creativity: 30 }, personality: { temperament: 'calm', playStyle: 'defensive' }, price: 500, isListed: false, createdAt: new Date() },
  { id: '2', name: 'Marcelo Silva', position: 'CB', nationality: 'Brasil', age: 26, attributes: { speed: 60, strength: 80, stamina: 75, finishing: 30, passing: 55, dribbling: 40, defense: 82, intelligence: 68, aggression: 65, leadership: 70, creativity: 35 }, personality: { temperament: 'calculated', playStyle: 'defensive' }, price: 450, isListed: false, createdAt: new Date() },
  { id: '3', name: 'Bruno Fernandes', position: 'CM', nationality: 'Portugal', age: 27, attributes: { speed: 70, strength: 65, stamina: 80, finishing: 75, passing: 88, dribbling: 78, defense: 55, intelligence: 85, aggression: 55, leadership: 80, creativity: 88 }, personality: { temperament: 'calculated', playStyle: 'offensive' }, price: 800, isListed: false, createdAt: new Date() },
  { id: '4', name: 'Neymar Jr', position: 'LW', nationality: 'Brasil', age: 29, attributes: { speed: 90, strength: 55, stamina: 75, finishing: 85, passing: 80, dribbling: 95, defense: 30, intelligence: 82, aggression: 45, leadership: 60, creativity: 95 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1500, isListed: false, createdAt: new Date() },
  { id: '5', name: 'Pedro Striker', position: 'ST', nationality: 'Brasil', age: 24, attributes: { speed: 82, strength: 70, stamina: 78, finishing: 90, passing: 65, dribbling: 75, defense: 25, intelligence: 72, aggression: 70, leadership: 55, creativity: 70 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1200, isListed: false, createdAt: new Date() },
];

const getPositionColor = (pos: string) => {
  if (pos === 'GK') return '#FFD700';
  if (['CB', 'LB', 'RB'].includes(pos)) return '#4169E1';
  if (['CDM', 'CM', 'CAM'].includes(pos)) return '#32CD32';
  return '#FF4500';
};

export default function TeamScreen() {
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);

  const renderPlayer = ({ item }: { item: Player }) => (
    <TouchableOpacity
      style={[styles.playerRow, selectedPlayer?.id === item.id && styles.selectedRow]}
      onPress={() => setSelectedPlayer(item)}
    >
      <View style={styles.playerInfo}>
        <Text style={styles.playerName}>{item.name}</Text>
        <Text style={[styles.position, { color: getPositionColor(item.position) }]}>
          {item.position}
        </Text>
      </View>
      <View style={styles.playerStats}>
        <Text style={styles.overall}>{calculateOverall(item.attributes)}</Text>
        <Text style={styles.price}>{item.price} FTC</Text>
      </View>
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      <Text style={styles.title}>⚽ Squad</Text>
      
      <FlatList
        data={mockPlayers}
        renderItem={renderPlayer}
        keyExtractor={(item) => item.id}
        style={styles.list}
      />

      {selectedPlayer && (
        <View style={styles.detailCard}>
          <Text style={styles.detailName}>{selectedPlayer.name}</Text>
          <Text style={[styles.detailPos, { color: getPositionColor(selectedPlayer.position) }]}>
            {selectedPlayer.position} | {selectedPlayer.nationality}
          </Text>
          <View style={styles.attrRow}>
            <Text style={styles.attrLabel}>Speed: {selectedPlayer.attributes.speed}</Text>
            <Text style={styles.attrLabel}>Finishing: {selectedPlayer.attributes.finishing}</Text>
          </View>
          <View style={styles.attrRow}>
            <Text style={styles.attrLabel}>Passing: {selectedPlayer.attributes.passing}</Text>
            <Text style={styles.attrLabel}>Defense: {selectedPlayer.attributes.defense}</Text>
          </View>
          <TouchableOpacity style={styles.sellBtn}>
            <Text style={styles.sellBtnText}>List for Sale</Text>
          </TouchableOpacity>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1A1A2E',
    padding: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#00D4AA',
    marginBottom: 16,
  },
  list: {
    flex: 1,
  },
  playerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    backgroundColor: '#0F0F1A',
    padding: 16,
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  selectedRow: {
    borderColor: '#00D4AA',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '600',
  },
  position: {
    fontSize: 14,
    fontWeight: 'bold',
  },
  playerStats: {
    alignItems: 'flex-end',
  },
  overall: {
    color: '#00D4AA',
    fontSize: 20,
    fontWeight: 'bold',
  },
  price: {
    color: '#FFD700',
    fontSize: 12,
  },
  detailCard: {
    backgroundColor: '#0F0F1A',
    padding: 16,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#00D4AA',
    marginTop: 16,
  },
  detailName: {
    color: '#FFF',
    fontSize: 20,
    fontWeight: 'bold',
  },
  detailPos: {
    fontSize: 14,
    marginBottom: 12,
  },
  attrRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  attrLabel: {
    color: '#AAA',
    fontSize: 14,
  },
  sellBtn: {
    backgroundColor: '#7B61FF',
    padding: 12,
    borderRadius: 8,
    marginTop: 12,
  },
  sellBtnText: {
    color: '#FFF',
    textAlign: 'center',
    fontWeight: 'bold',
  },
});

