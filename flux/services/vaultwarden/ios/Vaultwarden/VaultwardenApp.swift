import SwiftUI

@main
struct VaultwardenApp: App {
    @StateObject private var authService = AuthService.shared
    
    var body: some Scene {
        WindowGroup {
            if authService.isAuthenticated {
                VaultView()
                    .environmentObject(authService)
            } else {
                LoginView()
                    .environmentObject(authService)
            }
        }
    }
}
