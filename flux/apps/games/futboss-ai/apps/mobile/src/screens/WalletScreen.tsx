// FutBoss AI - Wallet Screen
// Author: Bruno Lucena (bruno@lucena.cloud)

import React, { useState } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView, Alert } from 'react-native';
import { useStore } from '../store';

const packages = [
  { id: 1, tokens: 1000, price: 10 },
  { id: 2, tokens: 5000, price: 45, popular: true },
  { id: 3, tokens: 10000, price: 80 },
  { id: 4, tokens: 25000, price: 175 },
];

export default function WalletScreen() {
  const { balance } = useStore();
  const [selectedPkg, setSelectedPkg] = useState<number | null>(null);
  const [paymentMethod, setPaymentMethod] = useState<'pix' | 'bitcoin'>('pix');

  const handlePurchase = () => {
    if (!selectedPkg) return;
    const pkg = packages.find(p => p.id === selectedPkg);
    Alert.alert(
      'Confirm Purchase',
      `Buy ${pkg?.tokens.toLocaleString()} FTC for R$ ${pkg?.price} via ${paymentMethod.toUpperCase()}?`,
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Buy', onPress: () => Alert.alert('Payment', 'Redirecting to payment...') },
      ]
    );
  };

  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>💰 Wallet</Text>

      {/* Balance */}
      <View style={styles.balanceCard}>
        <Text style={styles.balanceLabel}>Current Balance</Text>
        <Text style={styles.balanceValue}>{balance.toLocaleString()} FTC</Text>
      </View>

      {/* Packages */}
      <Text style={styles.sectionTitle}>Buy FutCoins</Text>
      <View style={styles.packagesGrid}>
        {packages.map((pkg) => (
          <TouchableOpacity
            key={pkg.id}
            style={[styles.packageCard, selectedPkg === pkg.id && styles.selectedPkg]}
            onPress={() => setSelectedPkg(pkg.id)}
          >
            {pkg.popular && (
              <View style={styles.popularBadge}>
                <Text style={styles.popularText}>POPULAR</Text>
              </View>
            )}
            <Text style={styles.pkgTokens}>{pkg.tokens.toLocaleString()}</Text>
            <Text style={styles.pkgLabel}>FutCoins</Text>
            <Text style={styles.pkgPrice}>R$ {pkg.price}</Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Payment Methods */}
      <Text style={styles.sectionTitle}>Payment Method</Text>
      <View style={styles.methodsRow}>
        <TouchableOpacity
          style={[styles.methodBtn, paymentMethod === 'pix' && styles.selectedMethod]}
          onPress={() => setPaymentMethod('pix')}
        >
          <Text style={styles.methodIcon}>📱</Text>
          <Text style={styles.methodText}>PIX</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.methodBtn, paymentMethod === 'bitcoin' && styles.selectedMethod]}
          onPress={() => setPaymentMethod('bitcoin')}
        >
          <Text style={styles.methodIcon}>₿</Text>
          <Text style={styles.methodText}>Bitcoin</Text>
        </TouchableOpacity>
      </View>

      {/* Buy Button */}
      <TouchableOpacity
        style={[styles.buyBtn, !selectedPkg && styles.disabledBtn]}
        onPress={handlePurchase}
        disabled={!selectedPkg}
      >
        <Text style={styles.buyBtnText}>
          {selectedPkg
            ? `Buy ${packages.find(p => p.id === selectedPkg)?.tokens.toLocaleString()} FTC`
            : 'Select a package'}
        </Text>
      </TouchableOpacity>

      {/* History */}
      <View style={styles.historyCard}>
        <Text style={styles.historyTitle}>Transaction History</Text>
        <Text style={styles.noHistory}>No transactions yet</Text>
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
    fontSize: 24,
    fontWeight: 'bold',
    color: '#00D4AA',
    marginBottom: 16,
  },
  balanceCard: {
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 20,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: '#333',
  },
  balanceLabel: {
    color: '#666',
    fontSize: 14,
  },
  balanceValue: {
    color: '#FFD700',
    fontSize: 36,
    fontWeight: 'bold',
  },
  sectionTitle: {
    color: '#FFF',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  packagesGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    marginBottom: 24,
  },
  packageCard: {
    width: '48%',
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  selectedPkg: {
    borderColor: '#00D4AA',
    borderWidth: 2,
  },
  popularBadge: {
    position: 'absolute',
    top: -8,
    backgroundColor: '#FFD700',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  popularText: {
    color: '#1A1A2E',
    fontSize: 10,
    fontWeight: 'bold',
  },
  pkgTokens: {
    color: '#00D4AA',
    fontSize: 24,
    fontWeight: 'bold',
  },
  pkgLabel: {
    color: '#666',
    fontSize: 12,
    marginBottom: 8,
  },
  pkgPrice: {
    color: '#FFF',
    fontSize: 18,
    fontWeight: 'bold',
  },
  methodsRow: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  methodBtn: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#0F0F1A',
    padding: 16,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#333',
    gap: 8,
  },
  selectedMethod: {
    borderColor: '#00D4AA',
  },
  methodIcon: {
    fontSize: 24,
  },
  methodText: {
    color: '#FFF',
    fontWeight: 'bold',
  },
  buyBtn: {
    backgroundColor: '#00D4AA',
    padding: 16,
    borderRadius: 12,
    marginBottom: 24,
  },
  disabledBtn: {
    backgroundColor: '#333',
  },
  buyBtnText: {
    color: '#1A1A2E',
    textAlign: 'center',
    fontSize: 18,
    fontWeight: 'bold',
  },
  historyCard: {
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  historyTitle: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  noHistory: {
    color: '#666',
    textAlign: 'center',
    paddingVertical: 16,
  },
});

