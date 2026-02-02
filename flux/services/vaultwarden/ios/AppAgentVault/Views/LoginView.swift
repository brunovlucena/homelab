import SwiftUI

struct LoginView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    @StateObject private var authService = AuthService.shared
    @State private var email = ""
    @State private var password = ""
    @State private var errorMessage: String?
    @State private var isLoading = false
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                Image(systemName: "lock.shield.fill")
                    .font(.system(size: 60))
                    .foregroundColor(.blue)
                    .padding(.bottom, 20)
                
                Text("AppAgentVault")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                
                Text("Self-hosted password manager")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .padding(.bottom, 40)
                
                TextField("Email", text: $email)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .autocapitalization(.none)
                    .keyboardType(.emailAddress)
                
                SecureField("Password", text: $password)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                
                if let error = errorMessage {
                    Text(error)
                        .foregroundColor(.red)
                        .font(.caption)
                }
                
                Button(action: login) {
                    if isLoading {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                    } else {
                        Text("Login")
                            .frame(maxWidth: .infinity)
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(isLoading || email.isEmpty || password.isEmpty)
            }
            .padding()
            .navigationBarHidden(true)
        }
    }
    
    private func login() {
        isLoading = true
        errorMessage = nil
        
        Task {
            do {
                try await authService.login(
                    email: email,
                    password: password,
                    agent: appViewModel.selectedAgent
                )
            } catch {
                errorMessage = error.localizedDescription
            }
            isLoading = false
        }
    }
}

#Preview {
    LoginView()
        .environmentObject(AppViewModel())
}
