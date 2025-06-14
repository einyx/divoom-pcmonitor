
echo "Building Divoom Monitor for Linux..."

cd DivoomMonitor

# Restore dependencies
echo "Restoring dependencies..."
dotnet restore

# Build release version
echo "Building release version..."
dotnet build -c Release

# Create self-contained executable
echo "Creating self-contained executable..."
dotnet publish -c Release -r linux-x64 --self-contained true -p:PublishSingleFile=true -p:PublishTrimmed=true

echo "Build complete!"
echo "Executable location: DivoomMonitor/bin/Release/net8.0/linux-x64/publish/DivoomMonitor"
echo ""
echo "To run: sudo ./DivoomMonitor/bin/Release/net8.0/linux-x64/publish/DivoomMonitor"
