import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    @StateObject private var authService = AuthService.shared
    
    var body: some View {
        NavigationView {
            Form {
                Section("Account") {
                    if let user = authService.currentUser {
                        LabeledContent("Email", value: user.email)
                        LabeledContent("Name", value: user.name)
                    }
                    
                    Button("Logout", role: .destructive) {
                        authService.logout()
                    }
                }
                
                Section("Appearance") {
                    Picker("Theme", selection: $appViewModel.settings.darkModeOverride) {
                        Text("System").tag(DarkModeOverride.system)
                        Text("Light").tag(DarkModeOverride.light)
                        Text("Dark").tag(DarkModeOverride.dark)
                    }
                }
                
                Section("Security") {
                    Toggle("Enable Biometric", isOn: $appViewModel.settings.enableBiometric)
                    Toggle("Notifications", isOn: $appViewModel.settings.enableNotifications)
                }
                
                Section("About") {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }
                    
                    Link("Documentation", destination: URL(string: "https://github.com/brunolucena/homelab")!)
                }
            }
            .navigationTitle("Settings")
        }
    }
}

#Preview {
    SettingsView()
        .environmentObject(AppViewModel())
}
