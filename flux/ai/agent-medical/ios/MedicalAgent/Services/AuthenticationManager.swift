//
//  AuthenticationManager.swift
//  MedicalAgent
//
//  Manages user authentication and token storage
//

import Foundation

@MainActor
class AuthenticationManager: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: User?
    @Published var errorMessage: String?
    
    private let tokenKey = "auth_token"
    private let userKey = "current_user"
    
    init() {
        // Check if user is already authenticated
        if let token = UserDefaults.standard.string(forKey: tokenKey),
           let userData = UserDefaults.standard.data(forKey: userKey),
           let user = try? JSONDecoder().decode(User.self, from: userData) {
            self.currentUser = user
            self.isAuthenticated = true
        }
    }
    
    func login(token: String, user: User) {
        UserDefaults.standard.set(token, forKey: tokenKey)
        if let userData = try? JSONEncoder().encode(user) {
            UserDefaults.standard.set(userData, forKey: userKey)
        }
        self.currentUser = user
        self.isAuthenticated = true
        errorMessage = nil
    }
    
    func logout() {
        UserDefaults.standard.removeObject(forKey: tokenKey)
        UserDefaults.standard.removeObject(forKey: userKey)
        self.currentUser = nil
        self.isAuthenticated = false
    }
    
    func getToken() -> String? {
        return UserDefaults.standard.string(forKey: tokenKey)
    }
}

// MARK: - User Model

struct User: Codable, Identifiable {
    let id: String
    let email: String
    let role: UserRole
    let name: String
    let patientAccess: [String]?
    
    enum CodingKeys: String, CodingKey {
        case id
        case email
        case role
        case name
        case patientAccess = "patient_access"
    }
}

enum UserRole: String, Codable {
    case doctor = "doctor"
    case nurse = "nurse"
    case patient = "patient"
    case admin = "admin"
    
    var displayName: String {
        switch self {
        case .doctor: return "Médico"
        case .nurse: return "Enfermeiro"
        case .patient: return "Paciente"
        case .admin: return "Administrador"
        }
    }
}
