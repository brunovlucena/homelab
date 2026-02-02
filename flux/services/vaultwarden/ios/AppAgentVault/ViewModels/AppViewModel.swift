import Foundation
import SwiftUI
import Combine

class AppViewModel: ObservableObject {
    @Published var showOnboarding = false
    @Published var isAuthenticated = false
    @Published var selectedAgent: Agent = Agent.vault
    @Published var settings = AppSettings()
    
    private let authService = AuthService.shared
    
    init() {
        // Check if first launch
        if UserDefaults.standard.bool(forKey: "has_completed_onboarding") == false {
            showOnboarding = true
        }
        
        // Observe authentication state
        authService.$isAuthenticated
            .assign(to: &$isAuthenticated)
    }
    
    func completeOnboarding() {
        UserDefaults.standard.set(true, forKey: "has_completed_onboarding")
        showOnboarding = false
    }
}

struct AppSettings: Codable {
    var darkModeOverride: DarkModeOverride = .system
    var enableNotifications: Bool = true
    var enableBiometric: Bool = true
}

enum DarkModeOverride: String, Codable, CaseIterable {
    case system
    case light
    case dark
}
