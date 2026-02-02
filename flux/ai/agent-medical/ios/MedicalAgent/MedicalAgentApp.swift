//
//  MedicalAgentApp.swift
//  MedicalAgent
//
//  iOS App for Medical Agent Communication
//

import SwiftUI

@main
struct MedicalAgentApp: App {
    @StateObject private var authManager = AuthenticationManager()
    @StateObject private var agentService = AgentService()
    
    var body: some Scene {
        WindowGroup {
            if authManager.isAuthenticated {
                ChatView()
                    .environmentObject(authManager)
                    .environmentObject(agentService)
            } else {
                LoginView()
                    .environmentObject(authManager)
            }
        }
    }
}
