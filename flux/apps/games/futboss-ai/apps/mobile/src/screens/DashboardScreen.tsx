// FutBoss AI - Dashboard Screen
// Author: Bruno Lucena (bruno@lucena.cloud)

import React from 'react';
import { View, Text, StyleSheet, ScrollView, TouchableOpacity } from 'react-native';
import { useStore } from '../store';

export default function DashboardScreen({ navigation }: any) {
  const { balance, team } = useStore();

  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>⚽ FutBoss AI</Text>
      <Text style={styles.subtitle}>by Bruno Lucena</Text>

      {/* Stats Cards */}
      <View style={styles.statsRow}>
        <View style={styles.statCard}>
          <Text style={styles.statLabel}>Balance</Text>
          <Text style={styles.statValue}>{balance.toLocaleString()} FTC</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statLabel}>Team</Text>
          <Text style={styles.statValue}>{team?.name || 'Create Team'}</Text>
        </View>
      </View>

      {/* Quick Actions */}
      <Text style={styles.sectionTitle}>Quick Actions</Text>
      <View style={styles.actionsGrid}>
        <TouchableOpacity
          style={styles.actionBtn}
          onPress={() => navigation.navigate('Team')}
        >
          <Text style={styles.actionIcon}>👥</Text>
          <Text style={styles.actionText}>Manage Team</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.actionBtn}
          onPress={() => navigation.navigate('Market')}
        >
          <Text style={styles.actionIcon}>🏪</Text>
          <Text style={styles.actionText}>Market</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.actionBtn}
          onPress={() => navigation.navigate('Match')}
        >
          <Text style={styles.actionIcon}>🎮</Text>
          <Text style={styles.actionText}>New Match</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.actionBtn}
          onPress={() => navigation.navigate('Wallet')}
        >
          <Text style={styles.actionIcon}>💰</Text>
          <Text style={styles.actionText}>Buy Tokens</Text>
        </TouchableOpacity>
      </View>

      {/* Info */}
      <View style={styles.infoCard}>
        <Text style={styles.infoTitle}>How to Play</Text>
        <Text style={styles.infoText}>⚽ Build your dream team</Text>
        <Text style={styles.infoText}>🤖 AI agents make smart decisions</Text>
        <Text style={styles.infoText}>🎮 Challenge other players</Text>
        <Text style={styles.infoText}>💰 Win FutCoins and climb rankings</Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1A1A2E',
    padding: 16,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#00D4AA',
    textAlign: 'center',
    marginTop: 20,
  },
  subtitle: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
  },
  statsRow: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  statCard: {
    flex: 1,
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  statLabel: {
    color: '#666',
    fontSize: 12,
    marginBottom: 4,
  },
  statValue: {
    color: '#FFD700',
    fontSize: 20,
    fontWeight: 'bold',
  },
  sectionTitle: {
    color: '#FFF',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  actionsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginBottom: 24,
  },
  actionBtn: {
    width: '47%',
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#00D4AA33',
  },
  actionIcon: {
    fontSize: 32,
    marginBottom: 8,
  },
  actionText: {
    color: '#FFF',
    fontWeight: '600',
  },
  infoCard: {
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  infoTitle: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  infoText: {
    color: '#AAA',
    fontSize: 14,
    marginBottom: 8,
  },
});

