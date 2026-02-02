import Foundation

// MARK: - Exercise Models

enum ExerciseType: String, Codable, CaseIterable {
    case wordRepetition = "word_repetition"
    case phraseCompletion = "phrase_completion"
    case storyTelling = "story_telling"
    case conversation = "conversation"
    case imitation = "imitation"
    case questionAnswer = "question_answer"
    
    var displayName: String {
        switch self {
        case .wordRepetition: return "Word Repetition"
        case .phraseCompletion: return "Phrase Completion"
        case .storyTelling: return "Story Telling"
        case .conversation: return "Conversation"
        case .imitation: return "Imitation"
        case .questionAnswer: return "Question & Answer"
        }
    }
    
    var icon: String {
        switch self {
        case .wordRepetition: return "repeat"
        case .phraseCompletion: return "text.bubble"
        case .storyTelling: return "book"
        case .conversation: return "bubble.left.and.bubble.right"
        case .imitation: return "person.2"
        case .questionAnswer: return "questionmark.circle"
        }
    }
}

enum DifficultyLevel: String, Codable {
    case beginner = "beginner"
    case intermediate = "intermediate"
    case advanced = "advanced"
    
    var displayName: String {
        rawValue.capitalized
    }
}

struct Exercise: Identifiable, Codable {
    let id: String
    let type: ExerciseType
    let title: String
    let description: String
    let difficulty: DifficultyLevel
    let instructions: String
    let targetWords: [String]
    let expectedDurationMinutes: Int
    let points: Int
}

enum GameStatus: String, Codable {
    case pending = "pending"
    case inProgress = "in_progress"
    case completed = "completed"
    case abandoned = "abandoned"
}

struct GameSession: Identifiable, Codable {
    let id: String
    let userId: String
    let exerciseId: String
    var status: GameStatus
    let startedAt: Date
    var completedAt: Date?
    var score: Int?
    var attempts: Int
    var feedback: String?
}
