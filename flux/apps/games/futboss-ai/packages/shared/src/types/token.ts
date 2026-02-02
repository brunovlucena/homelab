// FutBoss AI - Token Types
// Author: Bruno Lucena (bruno@lucena.cloud)

export type TransactionType = 'purchase' | 'sale' | 'buy_player' | 'match_reward' | 'bonus';

export interface TokenTransaction {
  id: string;
  userId: string;
  amount: number;
  transactionType: TransactionType;
  description: string;
  referenceId?: string;
  createdAt: Date;
}

export interface TokenWallet {
  userId: string;
  balance: number;
  totalEarned: number;
  totalSpent: number;
}

export function canAfford(wallet: TokenWallet, amount: number): boolean {
  return wallet.balance >= amount;
}

export function calculateTokensFromBRL(amountBRL: number, rate: number = 0.01): number {
  return Math.floor(amountBRL / rate);
}

