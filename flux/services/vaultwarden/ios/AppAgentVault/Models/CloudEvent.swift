import Foundation

// MARK: - CloudEvent Models

struct CloudEvent: Codable {
    let specversion: String
    let type: String
    let source: String
    let id: String
    let time: String
    let datacontenttype: String?
    let data: CloudEventData?
}

struct CloudEventData: Codable {
    let query: String?
    let action: String?
    let cipher: CipherData?
    let cipherId: String?
    let token: String?
    let conversationId: String?
    
    enum CodingKeys: String, CodingKey {
        case query
        case action
        case cipher
        case cipherId
        case token
        case conversationId
    }
}

struct CloudEventResponse: Codable {
    let specversion: String
    let type: String
    let source: String
    let id: String
    let time: String
    let data: CloudEventResponseData?
}

struct CloudEventResponseData: Codable {
    let response: String?
    let ciphers: [Cipher]?
    let cipher: Cipher?
    let success: Bool?
    let error: String?
    let conversationId: String?
    let agentId: String?
    
    enum CodingKeys: String, CodingKey {
        case response
        case ciphers
        case cipher
        case success
        case error
        case conversationId
        case agentId
    }
}

// MARK: - Cipher Models

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

struct CipherData: Codable {
    var id: String?
    let type: Int
    var name: String?
    var notes: String?
    var login: LoginInfo?
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case type = "Type"
        case name = "Name"
        case notes = "Notes"
        case login = "Login"
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

// MARK: - Health & Info Responses

struct HealthResponse: Codable {
    let status: String
}

struct AgentInfo: Codable {
    let name: String
    let version: String
    let capabilities: [String]
}
