// FutBoss AI - Wallet Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useState } from 'react';
import { useStore } from '../store';

type PaymentMethod = 'pix' | 'bitcoin';

const tokenPackages = [
  { id: 1, tokens: 1000, price: 10, popular: false },
  { id: 2, tokens: 5000, price: 45, popular: true },
  { id: 3, tokens: 10000, price: 80, popular: false },
  { id: 4, tokens: 25000, price: 175, popular: false },
];

export default function Wallet() {
  const { balance } = useStore();
  const [selectedPackage, setSelectedPackage] = useState<number | null>(null);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('pix');
  const [showPayment, setShowPayment] = useState(false);

  const handlePurchase = () => {
    if (!selectedPackage) return;
    setShowPayment(true);
  };

  const pkg = tokenPackages.find(p => p.id === selectedPackage);

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">💰 Wallet</h1>

      {/* Balance */}
      <div className="card mb-8">
        <p className="text-gray-400">Current Balance</p>
        <p className="text-4xl font-bold text-accent">{balance.toLocaleString()} FTC</p>
        <p className="text-sm text-gray-400 mt-2">
          FutCoins can be used to buy players and items
        </p>
      </div>

      {!showPayment ? (
        <>
          {/* Token Packages */}
          <h2 className="text-2xl font-bold mb-4">Buy FutCoins</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            {tokenPackages.map((pkg) => (
              <div
                key={pkg.id}
                onClick={() => setSelectedPackage(pkg.id)}
                className={`card cursor-pointer transition-all relative ${
                  selectedPackage === pkg.id
                    ? 'border-2 border-primary'
                    : 'hover:border-primary/50'
                }`}
              >
                {pkg.popular && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 bg-accent text-dark text-xs font-bold px-2 py-1 rounded">
                    POPULAR
                  </span>
                )}
                <p className="text-3xl font-bold text-primary mb-2">
                  {pkg.tokens.toLocaleString()}
                </p>
                <p className="text-gray-400">FutCoins</p>
                <p className="text-2xl font-bold mt-4">R$ {pkg.price}</p>
              </div>
            ))}
          </div>

          {/* Payment Method */}
          <h2 className="text-2xl font-bold mb-4">Payment Method</h2>
          <div className="grid grid-cols-2 gap-4 mb-8">
            <div
              onClick={() => setPaymentMethod('pix')}
              className={`card cursor-pointer flex items-center gap-4 ${
                paymentMethod === 'pix' ? 'border-2 border-primary' : ''
              }`}
            >
              <span className="text-4xl">📱</span>
              <div>
                <p className="font-bold">PIX</p>
                <p className="text-sm text-gray-400">Instant transfer</p>
              </div>
            </div>
            <div
              onClick={() => setPaymentMethod('bitcoin')}
              className={`card cursor-pointer flex items-center gap-4 ${
                paymentMethod === 'bitcoin' ? 'border-2 border-primary' : ''
              }`}
            >
              <span className="text-4xl">₿</span>
              <div>
                <p className="font-bold">Bitcoin</p>
                <p className="text-sm text-gray-400">Crypto payment</p>
              </div>
            </div>
          </div>

          <button
            onClick={handlePurchase}
            disabled={!selectedPackage}
            className="btn-primary w-full text-xl py-4 disabled:opacity-50"
          >
            {selectedPackage
              ? `Buy ${tokenPackages.find(p => p.id === selectedPackage)?.tokens.toLocaleString()} FTC`
              : 'Select a package'}
          </button>
        </>
      ) : (
        /* Payment Screen */
        <div className="card">
          <h2 className="text-2xl font-bold mb-6">Complete Purchase</h2>
          
          <div className="bg-dark p-4 rounded-lg mb-6">
            <div className="flex justify-between mb-2">
              <span>Package:</span>
              <span className="font-bold">{pkg?.tokens.toLocaleString()} FTC</span>
            </div>
            <div className="flex justify-between mb-2">
              <span>Method:</span>
              <span className="font-bold uppercase">{paymentMethod}</span>
            </div>
            <div className="flex justify-between text-xl">
              <span>Total:</span>
              <span className="font-bold text-accent">R$ {pkg?.price}</span>
            </div>
          </div>

          {paymentMethod === 'pix' ? (
            <div className="text-center">
              <p className="mb-4">Scan the QR Code or copy the PIX code:</p>
              <div className="bg-white p-4 rounded-lg inline-block mb-4">
                <div className="w-48 h-48 bg-gray-200 flex items-center justify-center">
                  [QR Code]
                </div>
              </div>
              <div className="bg-dark p-3 rounded-lg">
                <code className="text-sm break-all">
                  00020126580014br.gov.bcb.pix...
                </code>
              </div>
              <button className="btn-secondary mt-4">Copy PIX Code</button>
            </div>
          ) : (
            <div className="text-center">
              <p className="mb-4">Send Bitcoin to this address:</p>
              <div className="bg-dark p-3 rounded-lg mb-4">
                <code className="text-sm break-all">
                  bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh
                </code>
              </div>
              <p className="text-gray-400 mb-2">Amount: 0.00000285 BTC</p>
              <button className="btn-secondary">Copy Address</button>
            </div>
          )}

          <div className="flex gap-4 mt-8">
            <button
              onClick={() => setShowPayment(false)}
              className="btn-secondary flex-1"
            >
              Cancel
            </button>
            <button className="btn-primary flex-1">
              I've Paid
            </button>
          </div>
        </div>
      )}

      {/* Transaction History */}
      <div className="card mt-8">
        <h2 className="text-xl font-bold mb-4">Recent Transactions</h2>
        <div className="text-gray-400 text-center py-4">
          No transactions yet
        </div>
      </div>
    </div>
  );
}

