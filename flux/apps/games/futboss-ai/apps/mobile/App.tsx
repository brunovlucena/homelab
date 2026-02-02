// FutBoss AI - Mobile App Entry
// Author: Bruno Lucena (bruno@lucena.cloud)

import { StatusBar } from 'expo-status-bar';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { Text } from 'react-native';

import DashboardScreen from './src/screens/DashboardScreen';
import TeamScreen from './src/screens/TeamScreen';
import MarketScreen from './src/screens/MarketScreen';
import MatchScreen from './src/screens/MatchScreen';
import WalletScreen from './src/screens/WalletScreen';

const Tab = createBottomTabNavigator();
const Stack = createNativeStackNavigator();

const theme = {
  dark: true,
  colors: {
    primary: '#00D4AA',
    background: '#1A1A2E',
    card: '#0F0F1A',
    text: '#FFFFFF',
    border: '#333',
    notification: '#FFD700',
  },
};

function TabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={{
        tabBarStyle: { backgroundColor: '#0F0F1A', borderTopColor: '#333' },
        tabBarActiveTintColor: '#00D4AA',
        tabBarInactiveTintColor: '#666',
        headerStyle: { backgroundColor: '#0F0F1A' },
        headerTintColor: '#00D4AA',
      }}
    >
      <Tab.Screen
        name="Dashboard"
        component={DashboardScreen}
        options={{
          tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 20 }}>🏠</Text>,
          title: 'Home',
        }}
      />
      <Tab.Screen
        name="Team"
        component={TeamScreen}
        options={{
          tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 20 }}>⚽</Text>,
          title: 'Team',
        }}
      />
      <Tab.Screen
        name="Market"
        component={MarketScreen}
        options={{
          tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 20 }}>🏪</Text>,
          title: 'Market',
        }}
      />
      <Tab.Screen
        name="Match"
        component={MatchScreen}
        options={{
          tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 20 }}>🎮</Text>,
          title: 'Match',
        }}
      />
      <Tab.Screen
        name="Wallet"
        component={WalletScreen}
        options={{
          tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 20 }}>💰</Text>,
          title: 'Wallet',
        }}
      />
    </Tab.Navigator>
  );
}

export default function App() {
  return (
    <NavigationContainer theme={theme}>
      <StatusBar style="light" />
      <TabNavigator />
    </NavigationContainer>
  );
}

