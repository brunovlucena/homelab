import SwiftUI

struct OnboardingView: View {
    @EnvironmentObject var appViewModel: AppViewModel
    
    var body: some View {
        VStack(spacing: 30) {
            Image(systemName: "lock.shield.fill")
                .font(.system(size: 80))
                .foregroundColor(.blue)
            
            Text("Welcome to AppAgentVault")
                .font(.largeTitle)
                .fontWeight(.bold)
                .multilineTextAlignment(.center)
            
            Text("Your self-hosted password manager powered by AI agents")
                .font(.body)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal)
            
            VStack(alignment: .leading, spacing: 16) {
                FeatureRow(icon: "lock.shield", title: "Secure", description: "Client-side encryption with HashiCorp Vault")
                FeatureRow(icon: "bubble.left.and.bubble.right", title: "AI-Powered", description: "Chat with your password manager")
                FeatureRow(icon: "icloud.fill", title: "Self-Hosted", description: "Your data, your control")
            }
            .padding()
            
            Spacer()
            
            Button(action: {
                appViewModel.completeOnboarding()
            }) {
                Text("Get Started")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.blue)
                    .foregroundColor(.white)
                    .cornerRadius(12)
            }
            .padding()
        }
        .padding()
    }
}

struct FeatureRow: View {
    let icon: String
    let title: String
    let description: String
    
    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: icon)
                .font(.system(size: 30))
                .foregroundColor(.blue)
                .frame(width: 50)
            
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.headline)
                Text(description)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }
}

#Preview {
    OnboardingView()
        .environmentObject(AppViewModel())
}
