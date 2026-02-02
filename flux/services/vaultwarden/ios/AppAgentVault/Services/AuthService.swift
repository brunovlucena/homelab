import Foundation
import Combine

class AuthService: ObservableObject {
    static let shared = AuthService()
    
    @Published var isAuthenticated = false
    @Published var currentUser: User?
    
    private let keychainService = KeychainService.shared
    private let agentService = AgentService.shared
    
    private init() {
        // Check if we have a stored token
        if let email = UserDefaults.standard.string(forKey: "user_email"),
           let token = keychainService.getToken() {
            self.currentUser = User(
                email: email,
                name: UserDefaults.standard.string(forKey: "user_name") ?? email,
                token: token
            )
            self.isAuthenticated = true
        }
    }
    
    func login(email: String, password: String, agent: Agent) async throws {
        // Login via REST API (since CloudEvents login is not standard)
        guard let url = URL(string: "\(agent.baseURL)/api/identity/connect/token") else {
            throw AuthError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        
        let body = "grant_type=password&username=\(email.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? "")&password=\(password.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? "")"
        request.httpBody = body.data(using: .utf8)
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AuthError.invalidCredentials
        }
        
        let authResponse = try JSONDecoder().decode(AuthResponse.self, from: data)
        
        // Store token in keychain
        keychainService.saveToken(authResponse.access_token)
        
        // Store user info
        UserDefaults.standard.set(email, forKey: "user_email")
        UserDefaults.standard.set(email, forKey: "user_name") // Can be updated from profile
        
        DispatchQueue.main.async {
            self.currentUser = User(
                email: email,
                name: email,
                token: authResponse.access_token
            )
            self.isAuthenticated = true
        }
    }
    
    func logout() {
        keychainService.deleteToken()
        UserDefaults.standard.removeObject(forKey: "user_email")
        UserDefaults.standard.removeObject(forKey: "user_name")
        
        DispatchQueue.main.async {
            self.currentUser = nil
            self.isAuthenticated = false
        }
    }
}

struct AuthResponse: Codable {
    let access_token: String
    let token_type: String
    let expires_in: Int
}

enum AuthError: LocalizedError {
    case invalidURL
    case invalidCredentials
    case networkError
    
    var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid server URL"
        case .invalidCredentials: return "Invalid email or password"
        case .networkError: return "Network error occurred"
        }
    }
}
