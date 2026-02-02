import Foundation
import Combine

class AuthService: ObservableObject {
    static let shared = AuthService()
    
    @Published var isAuthenticated = false
    @Published var accessToken: String?
    
    private let keychainService = KeychainService.shared
    private let apiService = APIService.shared
    
    private init() {
        // Check if we have a stored token
        if let token = keychainService.getToken() {
            self.accessToken = token
            self.isAuthenticated = true
        }
    }
    
    func login(email: String, password: String) async throws {
        let response = try await apiService.login(email: email, password: password)
        
        // Store token in keychain
        keychainService.saveToken(response.access_token)
        
        DispatchQueue.main.async {
            self.accessToken = response.access_token
            self.isAuthenticated = true
        }
    }
    
    func logout() {
        keychainService.deleteToken()
        
        DispatchQueue.main.async {
            self.accessToken = nil
            self.isAuthenticated = false
        }
    }
}
