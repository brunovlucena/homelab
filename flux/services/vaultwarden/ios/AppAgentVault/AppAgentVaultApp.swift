import SwiftUI

@main
struct AppAgentVaultApp: App {
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

struct ContentView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    
    var body: some View {
        Group {
            if appViewModel.showOnboarding {
                OnboardingView()
            } else if !appViewModel.isAuthenticated {
                LoginView()
            } else {
                HomeView()
            }
        }
        .animation(.easeInOut, value: appViewModel.showOnboarding)
        .animation(.easeInOut, value: appViewModel.isAuthenticated)
    }
}

#Preview {
    ContentView()
        .environmentObject(AppViewModel())
}
