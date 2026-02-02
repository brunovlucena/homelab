import SwiftUI

// MARK: - Agent Picker View

struct AgentPickerView: View {
    @EnvironmentObject private var appViewModel: AppViewModel
    @Environment(\.dismiss) private var dismiss
    
    @Binding var agentStatuses: [UUID: AgentStatus]
    @State private var showAddAgent = false
    @State private var editingAgent: Agent?
    
    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    // Instructions
                    Text("Select an agent to chat with")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                        .padding(.top, 8)
                    
                    // Agent cards
                    ForEach(appViewModel.agents) { agent in
                        AgentCard(
                            agent: agent,
                            isSelected: appViewModel.currentAgent?.id == agent.id,
                            status: agentStatuses[agent.id] ?? .unknown,
                            onTap: {
                                appViewModel.selectAgent(agent)
                                dismiss()
                            }
                        )
                        .contextMenu {
                            Button {
                                editingAgent = agent
                            } label: {
                                Label("Edit", systemImage: "pencil")
                            }
                            
                            if !agent.isDefault {
                                Button(role: .destructive) {
                                    appViewModel.deleteAgent(agent)
                                } label: {
                                    Label("Delete", systemImage: "trash")
                                }
                            }
                        }
                    }
                    
                    // Add agent button
                    Button {
                        showAddAgent = true
                    } label: {
                        HStack {
                            Image(systemName: "plus.circle.fill")
                            Text("Add Custom Agent")
                        }
                        .font(.headline)
                        .foregroundColor(.accentColor)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 16, style: .continuous)
                                .stroke(Color.accentColor, style: StrokeStyle(lineWidth: 2, dash: [8]))
                        )
                    }
                }
                .padding(16)
            }
            .navigationTitle("Agents")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
            .sheet(isPresented: $showAddAgent) {
                AgentEditorView(agent: nil, onSave: { newAgent in
                    appViewModel.addAgent(newAgent)
                })
            }
            .sheet(item: $editingAgent) { agent in
                AgentEditorView(agent: agent, onSave: { updatedAgent in
                    appViewModel.updateAgent(updatedAgent)
                })
            }
        }
        .onAppear {
            refreshStatuses()
        }
    }
    
    private func refreshStatuses() {
        Task {
            for agent in appViewModel.agents {
                agentStatuses[agent.id] = .checking
                do {
                    let health = try await AgentService.shared.checkHealth(of: agent)
                    await MainActor.run {
                        agentStatuses[agent.id] = health.isHealthy ? .online : .degraded
                    }
                } catch {
                    await MainActor.run {
                        agentStatuses[agent.id] = .offline
                    }
                }
            }
        }
    }
}

// MARK: - Agent Editor View

struct AgentEditorView: View {
    let agent: Agent?
    let onSave: (Agent) -> Void
    
    @Environment(\.dismiss) private var dismiss
    
    @State private var name: String = ""
    @State private var description: String = ""
    @State private var baseURL: String = ""
    @State private var icon: String = "cpu"
    @State private var color: String = "#007AFF"
    @State private var eventType: String = ""
    @State private var isDefault: Bool = false
    @State private var testStatus: AgentStatus = .unknown
    @State private var isTesting: Bool = false
    
    private let icons = [
        "cpu", "bubble.left.and.bubble.right", "brain", "waveform",
        "cross.case.fill", "stethoscope", "heart.text.square",
        "doc.text", "folder", "chevron.left.forwardslash.chevron.right",
        "person.fill", "gear", "hammer", "wrench.and.screwdriver"
    ]
    
    private let colors = [
        "#007AFF", "#FF3B30", "#34C759", "#FF9500",
        "#5856D6", "#AF52DE", "#FF2D55", "#00C7BE"
    ]
    
    init(agent: Agent?, onSave: @escaping (Agent) -> Void) {
        self.agent = agent
        self.onSave = onSave
    }
    
    var body: some View {
        NavigationStack {
            Form {
                Section("Basic Info") {
                    TextField("Name", text: $name)
                    TextField("Description", text: $description)
                }
                
                Section("Connection") {
                    TextField("Base URL", text: $baseURL)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .keyboardType(.URL)
                    
                    TextField("Event Type (optional)", text: $eventType)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                    
                    HStack {
                        Button {
                            testConnection()
                        } label: {
                            if isTesting {
                                ProgressView()
                                    .scaleEffect(0.8)
                            } else {
                                Text("Test Connection")
                            }
                        }
                        .disabled(baseURL.isEmpty || isTesting)
                        
                        Spacer()
                        
                        AgentStatusBadge(status: testStatus)
                    }
                }
                
                Section("Appearance") {
                    // Icon picker
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Icon")
                            .font(.caption)
                            .foregroundColor(.secondary)
                        
                        LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 7), spacing: 12) {
                            ForEach(icons, id: \.self) { iconName in
                                Button {
                                    icon = iconName
                                } label: {
                                    Image(systemName: iconName)
                                        .font(.title2)
                                        .frame(width: 36, height: 36)
                                        .background(
                                            Circle()
                                                .fill(icon == iconName ? Color.accentColor.opacity(0.2) : Color.clear)
                                        )
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                    
                    // Color picker
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Color")
                            .font(.caption)
                            .foregroundColor(.secondary)
                        
                        HStack(spacing: 12) {
                            ForEach(colors, id: \.self) { colorHex in
                                Button {
                                    color = colorHex
                                } label: {
                                    Circle()
                                        .fill(Color(hex: colorHex))
                                        .frame(width: 32, height: 32)
                                        .overlay(
                                            Circle()
                                                .stroke(Color.primary, lineWidth: color == colorHex ? 3 : 0)
                                        )
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                }
                
                Section {
                    Toggle("Set as Default", isOn: $isDefault)
                }
            }
            .navigationTitle(agent == nil ? "Add Agent" : "Edit Agent")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Save") {
                        saveAgent()
                    }
                    .disabled(name.isEmpty || baseURL.isEmpty)
                }
            }
            .onAppear {
                if let agent = agent {
                    name = agent.name
                    description = agent.description
                    baseURL = agent.baseURL
                    icon = agent.icon
                    color = agent.color
                    eventType = agent.eventTypes.first ?? ""
                    isDefault = agent.isDefault
                }
            }
        }
    }
    
    private func testConnection() {
        isTesting = true
        testStatus = .checking
        
        let testAgent = Agent(
            name: name,
            description: description,
            baseURL: baseURL,
            icon: icon,
            color: color
        )
        
        Task {
            do {
                let health = try await AgentService.shared.checkHealth(of: testAgent)
                await MainActor.run {
                    testStatus = health.isHealthy ? .online : .degraded
                    isTesting = false
                }
            } catch {
                await MainActor.run {
                    testStatus = .offline
                    isTesting = false
                }
            }
        }
    }
    
    private func saveAgent() {
        var eventTypes: [String] = []
        if !eventType.isEmpty {
            eventTypes.append(eventType)
        }
        
        let newAgent = Agent(
            id: agent?.id ?? UUID(),
            name: name,
            description: description,
            baseURL: baseURL,
            icon: icon,
            color: color,
            eventTypes: eventTypes,
            isDefault: isDefault
        )
        
        onSave(newAgent)
        dismiss()
    }
}

// MARK: - Preview

#Preview {
    AgentPickerView(agentStatuses: .constant([:]))
        .environmentObject(AppViewModel())
}
