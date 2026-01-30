// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "TinfoilExample",
    platforms: [
        .macOS(.v14)
    ],
    dependencies: [
        .package(url: "https://github.com/tinfoilsh/tinfoil-swift", from: "0.4.0")
    ],
    targets: [
        .executableTarget(
            name: "TinfoilExample",
            dependencies: [
                .product(name: "TinfoilAI", package: "tinfoil-swift")
            ],
            path: "Sources"
        )
    ]
)
