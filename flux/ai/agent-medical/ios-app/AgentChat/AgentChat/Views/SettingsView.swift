import SwiftUI

// MARK: - Settings View

struct SettingsView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss
    
    @State private var settings: AppSettings = AppSettings()
    @State private var showLogoutConfirmation = false
    
    var body: some View {
        NavigationStack {
            Form {
                // User Section
                if let user = appViewModel.currentUser {
                    Section("User") {
                        HStack(spacing: 16) {
                            ZStack {
                                Circle()
                                    .fill(Color.accentColor.opacity(0.2))
                                    .frame(width: 50, height: 50)
                                
                                Image(systemName: user.role.icon)
                                    .font(.title2)
                                    .foregroundColor(.accentColor)
                            }
                            
                            VStack(alignment: .leading, spacing: 4) {
                                Text(user.name)
                                    .font(.headline)
                                
                                Text(user.email)
                                    .font(.caption)
                                    .foregroundColor(.secondary)
                                
                                HStack(spacing: 4) {
                                    Image(systemName: user.role.icon)
                                        .font(.caption2)
                                    Text(user.role.displayName)
                                        .font(.caption)
                                }
                                .foregroundColor(.secondary)
                            }
                        }
                        .padding(.vertical, 4)
                        
                        NavigationLink {
                            UserEditorView()
                        } label: {
                            Label("Edit Profile", systemImage: "pencil")
                        }
                    }
                }
                
                // Appearance
                Section("Appearance") {
                    Picker("Theme", selection: $settings.darkModeOverride) {
                        ForEach(AppSettings.DarkModeOption.allCases, id: \.self) { option in
                            Text(option.displayName).tag(option)
                        }
                    }
                    
                    Picker("Font Size", selection: $settings.fontSize) {
                        ForEach(AppSettings.FontSizeOption.allCases, id: \.self) { option in
                            Text(option.displayName).tag(option)
                        }
                    }
                }
                
                // Chat Settings
                Section("Chat") {
                    Toggle("Show Response Metadata", isOn: $settings.showMetadata)
                    Toggle("Auto-scroll to Bottom", isOn: $settings.autoScrollToBottom)
                    Toggle("Haptic Feedback", isOn: $settings.hapticFeedback)
                }
                
                // About
                Section("About") {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }
                    
                    Link(destination: URL(string: "https://github.com/brunolucena/homelab")!) {
                        HStack {
                            Text("GitHub")
                            Spacer()
                            Image(systemName: "arrow.up.right")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }
                    }
                }
                
                // Danger Zone
                Section {
                    Button(role: .destructive) {
                        showLogoutConfirmation = true
                    } label: {
                        Label("Sign Out", systemImage: "rectangle.portrait.and.arrow.right")
                    }
                }
            }
            .navigationTitle("Settings")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        appViewModel.updateSettings(settings)
                        dismiss()
                    }
                }
            }
            .onAppear {
                settings = appViewModel.settings
            }
            .confirmationDialog(
                "Sign Out",
                isPresented: $showLogoutConfirmation,
                titleVisibility: .visible
            ) {
                Button("Sign Out", role: .destructive) {
                    appViewModel.logout()
                    dismiss()
                }
                Button("Cancel", role: .cancel) {}
            } message: {
                Text("Are you sure you want to sign out? Your conversations will be deleted.")
            }
        }
    }
}

// MARK: - User Editor View

struct UserEditorView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss
    
    @State private var name: String = ""
    @State private var email: String = ""
    @State private var role: UserRole = .user
    @State private var token: String = ""
    
    var body: some View {
        Form {
            Section("Profile") {
                TextField("Name", text: $name)
                TextField("Email", text: $email)
                    .textInputAutocapitalization(.never)
                    .keyboardType(.emailAddress)
            }
            
            Section("Role") {
                Picker("Role", selection: $role) {
                    ForEach(UserRole.allCases, id: \.self) { role in
                        Label(role.displayName, systemImage: role.icon)
                            .tag(role)
                    }
                }
                .pickerStyle(.menu)
            }
            
            Section("Authentication") {
                SecureField("Token", text: $token)
                    .textInputAutocapitalization(.never)
                
                Text("Enter your authentication token to access protected agents.")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
        .navigationTitle("Edit Profile")
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button("Save") {
                    saveUser()
                }
                .disabled(name.isEmpty || email.isEmpty)
            }
        }
        .onAppear {
            if let user = appViewModel.currentUser {
                name = user.name
                email = user.email
                role = user.role
                token = user.token ?? ""
            }
        }
    }
    
    private func saveUser() {
        let user = User(
            id: appViewModel.currentUser?.id ?? UUID().uuidString,
            name: name,
            email: email,
            role: role,
            token: token.isEmpty ? nil : token
        )
        appViewModel.updateUser(user)
        dismiss()
    }
}

// MARK: - Preview

#Preview {
    SettingsView()
        .environmentObject(AppViewModel())
}
