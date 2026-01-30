import Foundation
import TinfoilAI
import OpenAI

@main
struct TinfoilExample {
    static func main() async {
        print("Tinfoil Swift SDK Example")
        print("=========================\n")

        do {
            print("Creating TinfoilAI client...")
            print("This will:")
            print("  1. Fetch attestation bundle through proxy")
            print("  2. Verify the enclave (remote attestation)")
            print("  3. Set up EHBP encryption")
            print("  4. Send requests through proxy at http://localhost:8080\n")

            let client = try await TinfoilAI.create(
                baseURL: "http://localhost:8080",
                attestationBundleURL: "http://localhost:8080",
                onVerification: { doc in
                    if let doc = doc {
                        print("Verification completed:")
                        print("  Enclave host: \(doc.enclaveHost)")
                        print("  Code fingerprint: \(doc.codeFingerprint)")
                        print("  Enclave fingerprint: \(doc.enclaveFingerprint)")
                        print("  Security verified: \(doc.securityVerified)\n")
                    }
                }
            )

            print("Client ready! Sending chat request...\n")

            let query = ChatQuery(
                messages: [
                    .user(.init(content: .string("Say hello in exactly 5 words.")))
                ],
                model: "gpt-oss-120b-free"
            )

            print("--- Streaming Response ---")
            for try await chunk in client.chatsStream(query: query) {
                if let content = chunk.choices.first?.delta.content {
                    print(content, terminator: "")
                    fflush(stdout)
                }
            }
            print("\n--- End Response ---\n")

            print("Example completed successfully!")

        } catch TinfoilError.missingAPIKey {
            print("Error: API key not found.")
            print("Set the TINFOIL_API_KEY environment variable:")
            print("  export TINFOIL_API_KEY=tk-your-key-here")

        } catch {
            print("Error: \(error)")
        }
    }
}
