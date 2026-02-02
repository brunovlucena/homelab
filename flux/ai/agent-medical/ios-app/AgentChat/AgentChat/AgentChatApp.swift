import SwiftUI

// MARK: - Main App Entry Point

@main
struct AgentChatApp: App {
    @StateObject private var appViewModel = AppViewModel()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(appViewModel)
                .preferredColorScheme(colorScheme)
        }
    }
    
    private var colorScheme: ColorScheme? {
        switch appViewModel.settings.darkModeOverride {
        case .system: return nil
        case .light: return .light
        case .dark: return .dark
        }
    }
}

// MARK: - Content View (Root Router)

struct ContentView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    
    var body: some View {
        Group {
            if appViewModel.showOnboarding {
                OnboardingView()
            } else {
                HomeView()
            }
        }
        .animation(.easeInOut, value: appViewModel.showOnboarding)
    }
}

// MARK: - Preview

#Preview {
    ContentView()
        .environmentObject(AppViewModel())
}
