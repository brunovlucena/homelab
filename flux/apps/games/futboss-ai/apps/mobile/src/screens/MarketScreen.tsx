// FutBoss AI - Market Screen
// Author: Bruno Lucena (bruno@lucena.cloud)

import React, { useState } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, Alert } from 'react-native';
import type { Player } from '@futboss/shared';
import { calculateOverall } from '@futboss/shared';
import { useStore } from '../store';

const marketPlayers: Player[] = [
  { id: 'm1', name: 'Ronaldo Silva', position: 'ST', nationality: 'Brasil', age: 23, attributes: { speed: 88, strength: 75, stamina: 82, finishing: 92, passing: 70, dribbling: 85, defense: 30, intelligence: 78, aggression: 65, leadership: 55, creativity: 80 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1500, isListed: true, createdAt: new Date() },
  { id: 'm2', name: 'Casemiro Jr', position: 'CDM', nationality: 'Brasil', age: 25, attributes: { speed: 65, strength: 85, stamina: 88, finishing: 55, passing: 75, dribbling: 60, defense: 88, intelligence: 82, aggression: 75, leadership: 80, creativity: 55 }, personality: { temperament: 'calculated', playStyle: 'defensive' }, price: 1200, isListed: true, createdAt: new Date() },
  { id: 'm3', name: 'Raphinha Costa', position: 'RW', nationality: 'Brasil', age: 26, attributes: { speed: 88, strength: 60, stamina: 78, finishing: 78, passing: 75, dribbling: 85, defense: 40, intelligence: 75, aggression: 55, leadership: 50, creativity: 82 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 950, isListed: true, createdAt: new Date() },
  { id: 'm4', name: 'Endrick Felipe', position: 'ST', nationality: 'Brasil', age: 18, attributes: { speed: 85, strength: 65, stamina: 78, finishing: 82, passing: 55, dribbling: 78, defense: 25, intelligence: 68, aggression: 70, leadership: 45, creativity: 75 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 800, isListed: true, createdAt: new Date() },
];

export default function MarketScreen() {
  const { balance } = useStore();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const handleBuy = (player: Player) => {
    if (balance < player.price) {
      Alert.alert('Insufficient Balance', 'You need more FutCoins to buy this player.');
      return;
    }
    Alert.alert('Confirm Purchase', `Buy ${player.name} for ${player.price} FTC?`, [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Buy', onPress: () => Alert.alert('Success!', `${player.name} joined your team!`) },
    ]);
  };

  const renderPlayer = ({ item }: { item: Player }) => (
    <TouchableOpacity
      style={[styles.playerCard, selectedId === item.id && styles.selectedCard]}
      onPress={() => setSelectedId(item.id)}
    >
      <View style={styles.cardHeader}>
        <Text style={styles.playerName}>{item.name}</Text>
        <View style={styles.overallBadge}>
          <Text style={styles.overallText}>{calculateOverall(item.attributes)}</Text>
        </View>
      </View>
      <Text style={styles.position}>{item.position} | {item.nationality}</Text>
      <Text style={styles.price}>{item.price} FTC</Text>
      
      {selectedId === item.id && (
        <TouchableOpacity
          style={[styles.buyBtn, balance < item.price && styles.disabledBtn]}
          onPress={() => handleBuy(item)}
        >
          <Text style={styles.buyBtnText}>
            {balance >= item.price ? 'Buy Player' : 'Not Enough Funds'}
          </Text>
        </TouchableOpacity>
      )}
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>🏪 Transfer Market</Text>
        <Text style={styles.balance}>{balance.toLocaleString()} FTC</Text>
      </View>

      <FlatList
        data={marketPlayers}
        renderItem={renderPlayer}
        keyExtractor={(item) => item.id}
        numColumns={2}
        columnWrapperStyle={styles.row}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1A1A2E',
    padding: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#00D4AA',
  },
  balance: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#FFD700',
  },
  row: {
    justifyContent: 'space-between',
  },
  playerCard: {
    width: '48%',
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 12,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#333',
  },
  selectedCard: {
    borderColor: '#00D4AA',
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  playerName: {
    color: '#FFF',
    fontSize: 14,
    fontWeight: 'bold',
    flex: 1,
  },
  overallBadge: {
    backgroundColor: '#7B61FF33',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  overallText: {
    color: '#7B61FF',
    fontWeight: 'bold',
  },
  position: {
    color: '#00D4AA',
    fontSize: 12,
    marginBottom: 8,
  },
  price: {
    color: '#FFD700',
    fontSize: 16,
    fontWeight: 'bold',
  },
  buyBtn: {
    backgroundColor: '#00D4AA',
    padding: 10,
    borderRadius: 8,
    marginTop: 12,
  },
  disabledBtn: {
    backgroundColor: '#666',
  },
  buyBtnText: {
    color: '#1A1A2E',
    textAlign: 'center',
    fontWeight: 'bold',
  },
});

