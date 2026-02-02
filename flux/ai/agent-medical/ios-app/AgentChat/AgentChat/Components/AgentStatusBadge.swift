import SwiftUI

// MARK: - Agent Status Badge (Reusable Component)

struct AgentStatusBadge: View {
    let status: AgentStatus
    let compact: Bool
    
    init(status: AgentStatus, compact: Bool = false) {
        self.status = status
        self.compact = compact
    }
    
    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: status.icon)
                .font(.caption)
                .foregroundColor(status.color)
                .symbolEffect(.pulse, isActive: status == .checking)
            
            if !compact {
                Text(status.text)
                    .font(.caption)
                    .foregroundColor(status.color)
            }
        }
        .padding(.horizontal, compact ? 6 : 10)
        .padding(.vertical, 4)
        .background(
            Capsule()
                .fill(status.color.opacity(0.15))
        )
    }
}

// MARK: - Agent Card (For Selection)

struct AgentCard: View {
    let agent: Agent
    let isSelected: Bool
    let status: AgentStatus
    let onTap: () -> Void
    
    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 16) {
                // Icon
                ZStack {
                    Circle()
                        .fill(Color(hex: agent.color).opacity(0.15))
                        .frame(width: 50, height: 50)
                    
                    Image(systemName: agent.icon)
                        .font(.title2)
                        .foregroundColor(Color(hex: agent.color))
                }
                
                // Info
                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Text(agent.name)
                            .font(.headline)
                            .foregroundColor(.primary)
                        
                        if agent.isDefault {
                            Text("Default")
                                .font(.caption2)
                                .foregroundColor(.secondary)
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(
                                    Capsule()
                                        .fill(Color(.systemGray5))
                                )
                        }
                    }
                    
                    Text(agent.description)
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }
                
                Spacer()
                
                // Status & Selection
                VStack(alignment: .trailing, spacing: 8) {
                    AgentStatusBadge(status: status, compact: true)
                    
                    if isSelected {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(.accentColor)
                    }
                }
            }
            .padding(16)
            .background(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .fill(Color(.systemBackground))
                    .shadow(color: isSelected ? .accentColor.opacity(0.3) : .black.opacity(0.05), radius: isSelected ? 8 : 4)
            )
            .overlay(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .stroke(isSelected ? Color.accentColor : .clear, lineWidth: 2)
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Color Extension

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        
        let a, r, g, b: UInt64
        switch hex.count {
        case 3: // RGB (12-bit)
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6: // RGB (24-bit)
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8: // ARGB (32-bit)
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 0, 0, 0)
        }
        
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

// MARK: - Preview

#Preview {
    VStack(spacing: 16) {
        AgentStatusBadge(status: .online)
        AgentStatusBadge(status: .offline, compact: true)
        
        AgentCard(
            agent: .medical,
            isSelected: true,
            status: .online,
            onTap: {}
        )
        .padding()
    }
}
