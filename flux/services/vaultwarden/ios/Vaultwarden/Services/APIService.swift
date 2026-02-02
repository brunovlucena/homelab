import Foundation
import Combine

class APIService {
    static let shared = APIService()
    
    private let baseURL = "https://vaultwarden.lucena.cloud"
    private let session = URLSession.shared
    
    private init() {}
    
    // MARK: - Authentication
    
    func login(email: String, password: String) async throws -> AuthResponse {
        let url = URL(string: "\(baseURL)/api/identity/connect/token")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        
        let body = "grant_type=password&username=\(email)&password=\(password)"
        request.httpBody = body.data(using: .utf8)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.invalidResponse
        }
        
        return try JSONDecoder().decode(AuthResponse.self, from: data)
    }
    
    // MARK: - Ciphers
    
    func listCiphers(token: String) async throws -> [Cipher] {
        let url = URL(string: "\(baseURL)/api/ciphers")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.invalidResponse
        }
        
        let responseData = try JSONDecoder().decode(CipherListResponse.self, from: data)
        return responseData.Data
    }
    
    func getCipher(id: String, token: String) async throws -> Cipher {
        let url = URL(string: "\(baseURL)/api/ciphers/\(id)")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.invalidResponse
        }
        
        return try JSONDecoder().decode(Cipher.self, from: data)
    }
    
    func createCipher(_ cipher: Cipher, token: String) async throws -> Cipher {
        let url = URL(string: "\(baseURL)/api/ciphers")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(cipher)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.invalidResponse
        }
        
        return try JSONDecoder().decode(Cipher.self, from: data)
    }
    
    func updateCipher(_ cipher: Cipher, token: String) async throws -> Cipher {
        guard let id = cipher.id else {
            throw APIError.invalidRequest
        }
        
        let url = URL(string: "\(baseURL)/api/ciphers/\(id)")!
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(cipher)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw APIError.invalidResponse
        }
        
        return try JSONDecoder().decode(Cipher.self, from: data)
    }
    
    func deleteCipher(id: String, token: String) async throws {
        let url = URL(string: "\(baseURL)/api/ciphers/\(id)")!
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 204 else {
            throw APIError.invalidResponse
        }
    }
}

// MARK: - Models

struct AuthResponse: Codable {
    let access_token: String
    let token_type: String
    let expires_in: Int
}

struct CipherListResponse: Codable {
    let Data: [Cipher]
    let Object: String
}

struct Cipher: Codable, Identifiable {
    var id: String?
    let type: Int
    var name: String?
    var notes: String?
    var login: LoginInfo?
    var organizationId: String?
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case type = "Type"
        case name = "Name"
        case notes = "Notes"
        case login = "Login"
        case organizationId = "OrganizationId"
    }
}

struct LoginInfo: Codable {
    var username: String?
    var password: String?
    var uris: [URIInfo]?
    
    enum CodingKeys: String, CodingKey {
        case username = "Username"
        case password = "Password"
        case uris = "Uris"
    }
}

struct URIInfo: Codable {
    let uri: String?
    let match: Int?
    
    enum CodingKeys: String, CodingKey {
        case uri = "Uri"
        case match = "Match"
    }
}

enum APIError: Error {
    case invalidResponse
    case invalidRequest
    case unauthorized
}
