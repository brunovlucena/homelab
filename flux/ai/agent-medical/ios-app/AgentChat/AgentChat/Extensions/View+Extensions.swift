import SwiftUI

// MARK: - View Extensions

extension View {
    
    /// Hide keyboard
    func hideKeyboard() {
        UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
    }
    
    /// Apply corner radius to specific corners
    func cornerRadius(_ radius: CGFloat, corners: UIRectCorner) -> some View {
        clipShape(RoundedCorner(radius: radius, corners: corners))
    }
    
    /// Conditional modifier
    @ViewBuilder
    func `if`<Content: View>(_ condition: Bool, transform: (Self) -> Content) -> some View {
        if condition {
            transform(self)
        } else {
            self
        }
    }
    
    /// Apply shadow with color
    func shadow(color: Color = .black.opacity(0.1), radius: CGFloat = 8, x: CGFloat = 0, y: CGFloat = 4) -> some View {
        self.shadow(color: color, radius: radius, x: x, y: y)
    }
}

// MARK: - Rounded Corner Shape

struct RoundedCorner: Shape {
    var radius: CGFloat = .infinity
    var corners: UIRectCorner = .allCorners
    
    func path(in rect: CGRect) -> Path {
        let path = UIBezierPath(
            roundedRect: rect,
            byRoundingCorners: corners,
            cornerRadii: CGSize(width: radius, height: radius)
        )
        return Path(path.cgPath)
    }
}

// MARK: - Color Extensions

extension Color {
    
    /// System background color
    static var systemBackground: Color {
        Color(UIColor.systemBackground)
    }
    
    /// Secondary system background color
    static var secondarySystemBackground: Color {
        Color(UIColor.secondarySystemBackground)
    }
    
    /// Tertiary system background color
    static var tertiarySystemBackground: Color {
        Color(UIColor.tertiarySystemBackground)
    }
}

// MARK: - String Extensions

extension String {
    
    /// Trim whitespace and newlines
    var trimmed: String {
        trimmingCharacters(in: .whitespacesAndNewlines)
    }
    
    /// Check if string is a valid URL
    var isValidURL: Bool {
        guard let url = URL(string: self) else { return false }
        return url.scheme == "http" || url.scheme == "https"
    }
}

// MARK: - Date Extensions

extension Date {
    
    /// Format as relative time (e.g., "2 hours ago")
    var relativeFormatted: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: self, relativeTo: Date())
    }
    
    /// Format as time (e.g., "2:30 PM")
    var timeFormatted: String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: self)
    }
    
    /// Format as date (e.g., "Jan 15, 2024")
    var dateFormatted: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        return formatter.string(from: self)
    }
}

// MARK: - Haptic Feedback

enum HapticFeedback {
    
    static func light() {
        UIImpactFeedbackGenerator(style: .light).impactOccurred()
    }
    
    static func medium() {
        UIImpactFeedbackGenerator(style: .medium).impactOccurred()
    }
    
    static func heavy() {
        UIImpactFeedbackGenerator(style: .heavy).impactOccurred()
    }
    
    static func success() {
        UINotificationFeedbackGenerator().notificationOccurred(.success)
    }
    
    static func warning() {
        UINotificationFeedbackGenerator().notificationOccurred(.warning)
    }
    
    static func error() {
        UINotificationFeedbackGenerator().notificationOccurred(.error)
    }
}
