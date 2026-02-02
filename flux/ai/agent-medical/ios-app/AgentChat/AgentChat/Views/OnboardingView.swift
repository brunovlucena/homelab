import SwiftUI

// MARK: - Onboarding View

struct OnboardingView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    
    @State private var currentPage = 0
    @State private var name: String = ""
    @State private var email: String = ""
    @State private var role: UserRole = .patient
    @State private var agentURL: String = ""
    @State private var isTestingConnection = false
    @State private var connectionStatus: AgentStatus = .unknown
    
    var body: some View {
        TabView(selection: $currentPage) {
            // Page 1: Welcome
            welcomePage
                .tag(0)
            
            // Page 2: User Setup
            userSetupPage
                .tag(1)
            
            // Page 3: Agent Connection
            agentSetupPage
                .tag(2)
        }
        .tabViewStyle(.page(indexDisplayMode: .always))
        .indexViewStyle(.page(backgroundDisplayMode: .always))
        .animation(.easeInOut, value: currentPage)
    }
    
    // MARK: - Welcome Page
    
    private var welcomePage: some View {
        VStack(spacing: 32) {
            Spacer()
            
            // Logo/Icon
            ZStack {
                Circle()
                    .fill(
                        LinearGradient(
                            colors: [.blue, .purple],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: 120, height: 120)
                
                Image(systemName: "bubble.left.and.bubble.right.fill")
                    .font(.system(size: 50))
                    .foregroundColor(.white)
            }
            
            // Title
            VStack(spacing: 12) {
                Text("Agent Chat")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                
                Text("Connect to your homelab AI agents")
                    .font(.title3)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
            }
            
            // Features
            VStack(alignment: .leading, spacing: 16) {
                FeatureRow(
                    icon: "cpu",
                    title: "Multiple Agents",
                    description: "Connect to any CloudEvents-compatible agent"
                )
                
                FeatureRow(
                    icon: "lock.shield",
                    title: "Secure",
                    description: "HIPAA-compliant medical records support"
                )
                
                FeatureRow(
                    icon: "iphone",
                    title: "Native",
                    description: "Built for iOS with SwiftUI"
                )
            }
            .padding(.horizontal, 32)
            
            Spacer()
            
            // Continue button
            Button {
                withAnimation {
                    currentPage = 1
                }
            } label: {
                Text("Get Started")
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 16)
                    .background(
                        RoundedRectangle(cornerRadius: 14, style: .continuous)
                            .fill(Color.accentColor)
                    )
            }
            .padding(.horizontal, 32)
            .padding(.bottom, 32)
        }
    }
    
    // MARK: - User Setup Page
    
    private var userSetupPage: some View {
        VStack(spacing: 24) {
            Spacer()
            
            // Icon
            ZStack {
                Circle()
                    .fill(Color.accentColor.opacity(0.15))
                    .frame(width: 80, height: 80)
                
                Image(systemName: "person.fill")
                    .font(.system(size: 36))
                    .foregroundColor(.accentColor)
            }
            
            Text("Your Profile")
                .font(.title)
                .fontWeight(.bold)
            
            // Form
            VStack(spacing: 16) {
                TextField("Name", text: $name)
                    .textFieldStyle(.roundedBorder)
                
                TextField("Email", text: $email)
                    .textFieldStyle(.roundedBorder)
                    .textInputAutocapitalization(.never)
                    .keyboardType(.emailAddress)
                
                // Role picker
                VStack(alignment: .leading, spacing: 8) {
                    Text("Role")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    
                    Picker("Role", selection: $role) {
                        ForEach(UserRole.allCases, id: \.self) { role in
                            Label(role.displayName, systemImage: role.icon)
                                .tag(role)
                        }
                    }
                    .pickerStyle(.segmented)
                }
            }
            .padding(.horizontal, 32)
            
            Spacer()
            
            // Continue button
            Button {
                withAnimation {
                    currentPage = 2
                }
            } label: {
                Text("Continue")
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 16)
                    .background(
                        RoundedRectangle(cornerRadius: 14, style: .continuous)
                            .fill(canContinueUserSetup ? Color.accentColor : Color.gray)
                    )
            }
            .disabled(!canContinueUserSetup)
            .padding(.horizontal, 32)
            .padding(.bottom, 32)
        }
    }
    
    private var canContinueUserSetup: Bool {
        !name.isEmpty && !email.isEmpty
    }
    
    // MARK: - Agent Setup Page
    
    private var agentSetupPage: some View {
        VStack(spacing: 24) {
            Spacer()
            
            // Icon
            ZStack {
                Circle()
                    .fill(Color.green.opacity(0.15))
                    .frame(width: 80, height: 80)
                
                Image(systemName: "network")
                    .font(.system(size: 36))
                    .foregroundColor(.green)
            }
            
            Text("Connect Agent")
                .font(.title)
                .fontWeight(.bold)
            
            Text("Enter your agent's URL or use the default medical agent")
                .font(.body)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
            
            // URL input
            VStack(spacing: 12) {
                TextField("Agent URL (optional)", text: $agentURL)
                    .textFieldStyle(.roundedBorder)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .keyboardType(.URL)
                
                if !agentURL.isEmpty {
                    HStack {
                        Button {
                            testConnection()
                        } label: {
                            if isTestingConnection {
                                ProgressView()
                                    .scaleEffect(0.8)
                            } else {
                                Text("Test Connection")
                            }
                        }
                        .disabled(isTestingConnection)
                        
                        Spacer()
                        
                        AgentStatusBadge(status: connectionStatus)
                    }
                }
            }
            .padding(.horizontal, 32)
            
            // Quick options
            VStack(spacing: 12) {
                Text("Quick Setup")
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                HStack(spacing: 12) {
                    QuickAgentButton(
                        name: "Medical",
                        icon: "cross.case.fill",
                        color: .red,
                        action: { agentURL = Agent.medical.baseURL }
                    )
                    
                    QuickAgentButton(
                        name: "Assistant",
                        icon: "bubble.left.and.bubble.right",
                        color: .blue,
                        action: { agentURL = Agent.assistant.baseURL }
                    )
                    
                    QuickAgentButton(
                        name: "Code",
                        icon: "chevron.left.forwardslash.chevron.right",
                        color: .purple,
                        action: { agentURL = Agent.code.baseURL }
                    )
                }
            }
            .padding(.horizontal, 32)
            
            Spacer()
            
            // Finish button
            Button {
                completeOnboarding()
            } label: {
                Text("Start Chatting")
                    .font(.headline)
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 16)
                    .background(
                        RoundedRectangle(cornerRadius: 14, style: .continuous)
                            .fill(Color.accentColor)
                    )
            }
            .padding(.horizontal, 32)
            .padding(.bottom, 32)
        }
    }
    
    // MARK: - Actions
    
    private func testConnection() {
        isTestingConnection = true
        connectionStatus = .checking
        
        let testAgent = Agent(
            name: "Test",
            description: "Test",
            baseURL: agentURL
        )
        
        Task {
            do {
                let health = try await AgentService.shared.checkHealth(of: testAgent)
                await MainActor.run {
                    connectionStatus = health.isHealthy ? .online : .degraded
                    isTestingConnection = false
                }
            } catch {
                await MainActor.run {
                    connectionStatus = .offline
                    isTestingConnection = false
                }
            }
        }
    }
    
    private func completeOnboarding() {
        // Create user
        let user = User(
            name: name,
            email: email,
            role: role,
            token: "\(role.rawValue)-token" // Demo token
        )
        appViewModel.login(user: user)
        
        // Update agent URL if custom
        if !agentURL.isEmpty {
            var updatedAgent = Agent.medical
            updatedAgent = Agent(
                id: updatedAgent.id,
                name: updatedAgent.name,
                description: updatedAgent.description,
                baseURL: agentURL,
                icon: updatedAgent.icon,
                color: updatedAgent.color,
                eventTypes: updatedAgent.eventTypes,
                isDefault: true
            )
            appViewModel.updateAgent(updatedAgent)
        }
    }
}

// MARK: - Feature Row

struct FeatureRow: View {
    let icon: String
    let title: String
    let description: String
    
    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundColor(.accentColor)
                .frame(width: 40)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.headline)
                
                Text(description)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }
}

// MARK: - Quick Agent Button

struct QuickAgentButton: View {
    let name: String
    let icon: String
    let color: Color
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.title2)
                    .foregroundColor(color)
                
                Text(name)
                    .font(.caption)
                    .foregroundColor(.primary)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(Color(.systemGray6))
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Preview

#Preview {
    OnboardingView()
        .environmentObject(AppViewModel())
}
