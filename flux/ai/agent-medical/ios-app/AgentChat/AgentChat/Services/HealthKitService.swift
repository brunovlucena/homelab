import Foundation
import HealthKit
import Combine

// MARK: - HealthKit Service

@MainActor
final class HealthKitService: ObservableObject {
    
    static let shared = HealthKitService()
    
    // MARK: - Published Properties
    
    @Published var isAuthorized = false
    @Published var lastHeartRate: HeartRateReading?
    @Published var authorizationStatus: HKAuthorizationStatus = .notDetermined
    @Published var errorMessage: String?
    
    // MARK: - Private Properties
    
    private let healthStore = HKHealthStore()
    private var backgroundDeliveryQuery: HKQuery?
    private let agentService: AgentServiceProtocol
    
    // MARK: - Heart Rate Type
    
    private var heartRateType: HKQuantityType? {
        HKQuantityType.quantityType(forIdentifier: .heartRate)
    }
    
    // MARK: - Initialization
    
    init(agentService: AgentServiceProtocol = AgentService.shared) {
        self.agentService = agentService
        checkAuthorizationStatus()
    }
    
    // MARK: - Authorization
    
    func requestAuthorization() async throws {
        guard let heartRateType = heartRateType else {
            throw HealthKitError.unsupportedDevice
        }
        
        let typesToRead: Set<HKObjectType> = [heartRateType]
        let typesToShare: Set<HKSampleType> = [] // We only read, don't write
        
        do {
            try await healthStore.requestAuthorization(toShare: typesToShare, read: typesToRead)
            await checkAuthorizationStatus()
        } catch {
            throw HealthKitError.authorizationFailed(error.localizedDescription)
        }
    }
    
    func checkAuthorizationStatus() {
        guard let heartRateType = heartRateType else {
            authorizationStatus = .notDetermined
            isAuthorized = false
            return
        }
        
        let status = healthStore.authorizationStatus(for: heartRateType)
        authorizationStatus = status
        isAuthorized = status == .sharingAuthorized
    }
    
    // MARK: - Read Heart Rate Data
    
    /// Reads the most recent heart rate reading
    func readLatestHeartRate() async throws -> HeartRateReading? {
        guard let heartRateType = heartRateType else {
            throw HealthKitError.unsupportedDevice
        }
        
        guard authorizationStatus == .sharingAuthorized else {
            throw HealthKitError.notAuthorized
        }
        
        return try await withCheckedThrowingContinuation { continuation in
            let sortDescriptor = NSSortDescriptor(key: HKSampleSortIdentifierEndDate, ascending: false)
            let query = HKSampleQuery(
                sampleType: heartRateType,
                predicate: nil,
                limit: 1,
                sortDescriptors: [sortDescriptor]
            ) { [weak self] _, results, error in
                guard let self = self else {
                    continuation.resume(returning: nil)
                    return
                }
                
                if let error = error {
                    Task { @MainActor in
                        self.errorMessage = error.localizedDescription
                    }
                    continuation.resume(throwing: HealthKitError.queryFailed(error.localizedDescription))
                    return
                }
                
                guard let sample = results?.first as? HKQuantitySample else {
                    continuation.resume(returning: nil)
                    return
                }
                
                let heartRateUnit = HKUnit(from: "count/min")
                let heartRateValue = sample.quantity.doubleValue(for: heartRateUnit)
                let reading = HeartRateReading(
                    bpm: Int(heartRateValue),
                    timestamp: sample.endDate,
                    device: sample.device?.name ?? "Unknown",
                    metadata: sample.metadata
                )
                
                Task { @MainActor in
                    self.lastHeartRate = reading
                }
                
                continuation.resume(returning: reading)
            }
            
            healthStore.execute(query)
        }
    }
    
    /// Reads heart rate data for a time range
    func readHeartRateData(
        from startDate: Date,
        to endDate: Date,
        limit: Int = 100
    ) async throws -> [HeartRateReading] {
        guard let heartRateType = heartRateType else {
            throw HealthKitError.unsupportedDevice
        }
        
        guard authorizationStatus == .sharingAuthorized else {
            throw HealthKitError.notAuthorized
        }
        
        let predicate = HKQuery.predicateForSamples(
            withStart: startDate,
            end: endDate,
            options: .strictStartDate
        )
        
        let sortDescriptor = NSSortDescriptor(key: HKSampleSortIdentifierEndDate, ascending: false)
        
        return try await withCheckedThrowingContinuation { continuation in
            let query = HKSampleQuery(
                sampleType: heartRateType,
                predicate: predicate,
                limit: limit,
                sortDescriptors: [sortDescriptor]
            ) { _, results, error in
                if let error = error {
                    continuation.resume(throwing: error)
                    return
                }
                
                guard let samples = results as? [HKQuantitySample] else {
                    continuation.resume(returning: [])
                    return
                }
                
                let heartRateUnit = HKUnit(from: "count/min")
                let readings = samples.map { sample in
                    HeartRateReading(
                        bpm: Int(sample.quantity.doubleValue(for: heartRateUnit)),
                        timestamp: sample.endDate,
                        device: sample.device?.name ?? "Unknown",
                        metadata: sample.metadata
                    )
                }
                
                continuation.resume(returning: readings)
            }
            
            healthStore.execute(query)
        }
    }
    
    // MARK: - Background Delivery
    
    /// Enables background delivery of heart rate updates
    func enableBackgroundDelivery() async throws {
        guard let heartRateType = heartRateType else {
            throw HealthKitError.unsupportedDevice
        }
        
        guard authorizationStatus == .sharingAuthorized else {
            throw HealthKitError.notAuthorized
        }
        
        try await healthStore.enableBackgroundDelivery(
            for: heartRateType,
            frequency: .immediate
        )
        
        // Set up observer query for background updates
        setupObserverQuery()
    }
    
    private func setupObserverQuery() {
        guard let heartRateType = heartRateType else { return }
        
        // Remove existing query
        if let existingQuery = backgroundDeliveryQuery {
            healthStore.stop(existingQuery)
        }
        
        let query = HKObserverQuery(sampleType: heartRateType, predicate: nil) { [weak self] query, completionHandler, error in
            guard let self = self else {
                completionHandler()
                return
            }
            
            if let error = error {
                Task { @MainActor in
                    self.errorMessage = "Background delivery error: \(error.localizedDescription)"
                }
                completionHandler()
                return
            }
            
            // Fetch latest reading (don't send automatically - let app decide when to send)
            Task {
                do {
                    _ = try await self.readLatestHeartRate()
                    // Note: Actual sending should be triggered by the app, not automatically
                } catch {
                    // Log error but don't block completion
                }
                completionHandler()
            }
        }
        
        backgroundDeliveryQuery = query
        healthStore.execute(query)
    }
    
    // MARK: - Send Data to Agent
    
    /// Sends heart rate data to the medical agent via CloudEvent
    func sendHeartRateToAgent(_ reading: HeartRateReading, patientId: String?, agentBaseURL: String, userToken: String?) async throws {
        // Use provided patient ID or empty
        let finalPatientId = patientId ?? ""
        
        // Build CloudEvent data
        let eventData: [String: Any] = [
            "patient_id": finalPatientId,
            "heart_rate_bpm": reading.bpm,
            "timestamp": ISO8601DateFormatter().string(from: reading.timestamp),
            "context": "resting", // Can be enhanced to detect context
            "device": reading.device
        ]
        
        // Send CloudEvent to agent
        do {
            let eventDataJson = try JSONSerialization.data(withJSONObject: eventData)
            let eventDataString = String(data: eventDataJson, encoding: .utf8) ?? "{}"
            
        // Build CloudEvent request
        guard let url = URL(string: agentBaseURL) else {
            throw HealthKitError.queryFailed("Invalid agent URL")
        }
            
            var request = URLRequest(url: url)
            request.httpMethod = "POST"
            request.timeoutInterval = 30
            
            // CloudEvents headers
            request.setValue("1.0", forHTTPHeaderField: "ce-specversion")
            request.setValue("io.homelab.medical.heart-rate.report", forHTTPHeaderField: "ce-type")
            request.setValue("/ios-app/agent-chat", forHTTPHeaderField: "ce-source")
            request.setValue(UUID().uuidString, forHTTPHeaderField: "ce-id")
            request.setValue(ISO8601DateFormatter().string(from: Date()), forHTTPHeaderField: "ce-time")
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.setValue("application/cloudevents+json", forHTTPHeaderField: "Accept")
            
            // Auth header
            if let token = userToken {
                request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
            }
            
            // Set body (CloudEvent structured format)
            let cloudEvent: [String: Any] = [
                "specversion": "1.0",
                "type": "io.homelab.medical.heart-rate.report",
                "source": "/ios-app/agent-chat",
                "id": UUID().uuidString,
                "time": ISO8601DateFormatter().string(from: Date()),
                "datacontenttype": "application/json",
                "data": eventData
            ]
            
            request.httpBody = try JSONSerialization.data(withJSONObject: cloudEvent)
            
            // Send request
            let (_, response) = try await URLSession.shared.data(for: request)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                throw HealthKitError.queryFailed("Invalid response")
            }
            
            if !(200...299).contains(httpResponse.statusCode) {
                throw HealthKitError.queryFailed("Server returned status \(httpResponse.statusCode)")
            }
            
            print("✅ Heart rate sent to agent: \(reading.bpm) bpm")
            
        } catch {
            print("❌ Failed to send heart rate to agent: \(error.localizedDescription)")
            throw HealthKitError.queryFailed(error.localizedDescription)
        }
    }
}

// MARK: - Heart Rate Reading Model

struct HeartRateReading: Codable, Identifiable {
    let id = UUID()
    let bpm: Int
    let timestamp: Date
    let device: String
    let metadata: [String: Any]?
    
    enum CodingKeys: String, CodingKey {
        case bpm
        case timestamp
        case device
    }
    
    init(bpm: Int, timestamp: Date, device: String, metadata: [String: Any]? = nil) {
        self.bpm = bpm
        self.timestamp = timestamp
        self.device = device
        self.metadata = metadata
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(bpm, forKey: .bpm)
        try container.encode(timestamp, forKey: .timestamp)
        try container.encode(device, forKey: .device)
    }
}

// MARK: - HealthKit Errors

enum HealthKitError: LocalizedError {
    case unsupportedDevice
    case notAuthorized
    case authorizationFailed(String)
    case queryFailed(String)
    
    var errorDescription: String? {
        switch self {
        case .unsupportedDevice:
            return "HealthKit is not supported on this device"
        case .notAuthorized:
            return "HealthKit authorization is required"
        case .authorizationFailed(let message):
            return "Authorization failed: \(message)"
        case .queryFailed(let message):
            return "Query failed: \(message)"
        }
    }
}

