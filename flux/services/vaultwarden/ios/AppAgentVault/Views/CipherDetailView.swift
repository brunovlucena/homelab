import SwiftUI

struct CipherDetailView: View {
    let cipher: Cipher
    @State private var showPassword = false
    
    var body: some View {
        Form {
            Section("Information") {
                Text(cipher.name ?? "Untitled")
                    .font(.headline)
                
                if let username = cipher.login?.username {
                    LabeledContent("Username") {
                        Text(username)
                            .textSelection(.enabled)
                    }
                }
                
                if let password = cipher.login?.password {
                    LabeledContent("Password") {
                        HStack {
                            if showPassword {
                                Text(password)
                                    .textSelection(.enabled)
                            } else {
                                Text("••••••••")
                            }
                            Button(action: { showPassword.toggle() }) {
                                Image(systemName: showPassword ? "eye.slash" : "eye")
                            }
                            Button(action: {
                                UIPasteboard.general.string = password
                            }) {
                                Image(systemName: "doc.on.doc")
                            }
                        }
                    }
                }
                
                if let notes = cipher.notes {
                    LabeledContent("Notes") {
                        Text(notes)
                            .textSelection(.enabled)
                    }
                }
            }
        }
        .navigationTitle(cipher.name ?? "Password")
        .navigationBarTitleDisplayMode(.inline)
    }
}

#Preview {
    NavigationView {
        CipherDetailView(cipher: Cipher(
            id: "1",
            type: 1,
            name: "Example Site",
            notes: "My notes",
            login: LoginInfo(
                username: "user@example.com",
                password: "password123",
                uris: nil
            ),
            organizationId: nil
        ))
    }
}
