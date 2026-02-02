// FutBoss AI - Match Screen
// Author: Bruno Lucena (bruno@lucena.cloud)

import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView } from 'react-native';

type MatchState = 'lobby' | 'playing' | 'finished';

export default function MatchScreen() {
  const [matchState, setMatchState] = useState<MatchState>('lobby');
  const [minute, setMinute] = useState(0);
  const [homeScore, setHomeScore] = useState(0);
  const [awayScore, setAwayScore] = useState(0);
  const [events, setEvents] = useState<string[]>([]);

  useEffect(() => {
    if (matchState !== 'playing') return;

    const interval = setInterval(() => {
      setMinute((m) => {
        if (m >= 90) {
          setMatchState('finished');
          return 90;
        }
        
        // Random events
        if (Math.random() < 0.08) {
          const isHome = Math.random() < 0.5;
          if (isHome) setHomeScore(s => s + 1);
          else setAwayScore(s => s + 1);
          setEvents(e => [...e, `${m}' ⚽ GOAL! ${isHome ? 'Home' : 'Away'} team scores!`]);
        }
        
        return m + 1;
      });
    }, 400);

    return () => clearInterval(interval);
  }, [matchState]);

  const startMatch = () => {
    setMatchState('playing');
    setMinute(0);
    setHomeScore(0);
    setAwayScore(0);
    setEvents([]);
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>🎮 Match Center</Text>

      {matchState === 'lobby' && (
        <View style={styles.lobbyCard}>
          <Text style={styles.lobbyTitle}>Ready to Play?</Text>
          <Text style={styles.lobbySubtitle}>Start a match against AI</Text>
          <TouchableOpacity style={styles.startBtn} onPress={startMatch}>
            <Text style={styles.startBtnText}>Start Match</Text>
          </TouchableOpacity>
        </View>
      )}

      {(matchState === 'playing' || matchState === 'finished') && (
        <>
          {/* Scoreboard */}
          <View style={styles.scoreboard}>
            <View style={styles.teamScore}>
              <Text style={styles.teamName}>My Team</Text>
              <Text style={styles.score}>{homeScore}</Text>
            </View>
            <View style={styles.matchInfo}>
              <Text style={styles.status}>
                {matchState === 'finished' ? 'FULL TIME' : 'LIVE'}
              </Text>
              <Text style={styles.minute}>{minute}'</Text>
            </View>
            <View style={styles.teamScore}>
              <Text style={styles.teamName}>AI United</Text>
              <Text style={[styles.score, { color: '#7B61FF' }]}>{awayScore}</Text>
            </View>
          </View>

          {/* Events */}
          <View style={styles.eventsCard}>
            <Text style={styles.eventsTitle}>Match Events</Text>
            <ScrollView style={styles.eventsList}>
              {events.length === 0 ? (
                <Text style={styles.noEvents}>Waiting for action...</Text>
              ) : (
                events.slice().reverse().map((event, i) => (
                  <Text key={i} style={styles.eventItem}>{event}</Text>
                ))
              )}
            </ScrollView>
          </View>

          {matchState === 'finished' && (
            <View style={styles.resultCard}>
              <Text style={styles.resultText}>
                {homeScore > awayScore ? '🎉 Victory!' : homeScore < awayScore ? '😢 Defeat' : '🤝 Draw'}
              </Text>
              <TouchableOpacity style={styles.startBtn} onPress={startMatch}>
                <Text style={styles.startBtnText}>Play Again</Text>
              </TouchableOpacity>
            </View>
          )}
        </>
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
  lobbyCard: {
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 32,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#333',
  },
  lobbyTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFF',
    marginBottom: 8,
  },
  lobbySubtitle: {
    color: '#666',
    marginBottom: 24,
  },
  startBtn: {
    backgroundColor: '#00D4AA',
    paddingVertical: 16,
    paddingHorizontal: 48,
    borderRadius: 12,
  },
  startBtnText: {
    color: '#1A1A2E',
    fontSize: 18,
    fontWeight: 'bold',
  },
  scoreboard: {
    flexDirection: 'row',
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  teamScore: {
    flex: 1,
    alignItems: 'center',
  },
  teamName: {
    color: '#FFF',
    fontSize: 14,
    marginBottom: 8,
  },
  score: {
    color: '#00D4AA',
    fontSize: 48,
    fontWeight: 'bold',
  },
  matchInfo: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 16,
  },
  status: {
    color: '#666',
    fontSize: 12,
  },
  minute: {
    color: '#FFD700',
    fontSize: 24,
    fontWeight: 'bold',
  },
  eventsCard: {
    flex: 1,
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  eventsTitle: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  eventsList: {
    flex: 1,
  },
  noEvents: {
    color: '#666',
    textAlign: 'center',
  },
  eventItem: {
    color: '#AAA',
    fontSize: 14,
    marginBottom: 8,
    paddingLeft: 8,
    borderLeftWidth: 2,
    borderLeftColor: '#00D4AA',
  },
  resultCard: {
    backgroundColor: '#0F0F1A',
    borderRadius: 12,
    padding: 24,
    alignItems: 'center',
    marginTop: 16,
    borderWidth: 1,
    borderColor: '#00D4AA',
  },
  resultText: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFF',
    marginBottom: 16,
  },
});

