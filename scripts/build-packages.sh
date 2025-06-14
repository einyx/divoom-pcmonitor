#!/bin/bash

# Comprehensive build and packaging script for divoom-pcmonitor
# Builds binaries, creates packages (DEB, RPM, Windows installer), and prepares releases

set -e

VERSION="${1:-1.0.0}"
BUILD_DIR="dist"
LDFLAGS="-s -w -X main.version=${VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building divoom-pcmonitor v${VERSION} - Complete Package Build${NC}"
echo "==============================================================================="

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/{binaries,packages/{deb,rpm,windows},archives}

# Build function for binaries
build_binaries() {
    local os=$1
    local arch=$2
    local extension=$3
    local target_name="${os}_${arch}"
    
    echo -e "${BLUE}Building binaries for ${os}/${arch}...${NC}"
    
    export GOOS=${os}
    export GOARCH=${arch}
    export CGO_ENABLED=0
    
    # Build all three applications
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/binaries/divoom-monitor-${target_name}${extension}" main.go
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/binaries/divoom-daemon-${target_name}${extension}" divoom_daemon.go
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/binaries/divoom-test-${target_name}${extension}" test_divoom.go
    
    echo -e "${GREEN}✓ Built binaries for ${os}/${arch}${NC}"
}

# Build all platform binaries
echo -e "${YELLOW}Building cross-platform binaries...${NC}"

# Linux builds
build_binaries "linux" "amd64" ""
build_binaries "linux" "386" ""
build_binaries "linux" "arm64" ""
build_binaries "linux" "arm" ""

# Windows builds
build_binaries "windows" "amd64" ".exe"
build_binaries "windows" "386" ".exe"
build_binaries "windows" "arm64" ".exe"

# macOS builds
build_binaries "darwin" "amd64" ""
build_binaries "darwin" "arm64" ""

# FreeBSD builds
build_binaries "freebsd" "amd64" ""

echo -e "${GREEN}✓ All binaries built successfully${NC}"

# Function to create DEB package
build_deb_package() {
    local arch=$1
    local deb_arch=$2
    
    echo -e "${BLUE}Building DEB package for ${arch}...${NC}"
    
    local deb_dir="${BUILD_DIR}/packages/deb/divoom-pcmonitor-${VERSION}-${deb_arch}"
    
    # Copy DEB structure
    cp -r packaging/deb ${deb_dir}
    
    # Update architecture in control file
    sed -i "s/Architecture: amd64/Architecture: ${deb_arch}/" ${deb_dir}/DEBIAN/control
    sed -i "s/Version: 1.0.0/Version: ${VERSION}/" ${deb_dir}/DEBIAN/control
    
    # Copy binaries
    cp ${BUILD_DIR}/binaries/divoom-monitor-linux_${arch} ${deb_dir}/usr/bin/divoom-monitor
    cp ${BUILD_DIR}/binaries/divoom-daemon-linux_${arch} ${deb_dir}/usr/bin/divoom-daemon
    cp ${BUILD_DIR}/binaries/divoom-test-linux_${arch} ${deb_dir}/usr/bin/divoom-test
    
    # Copy systemd files
    cp packaging/systemd/divoom-monitor.service ${deb_dir}/etc/systemd/system/
    cp packaging/systemd/divoom-user.conf ${deb_dir}/usr/lib/sysusers.d/divoom.conf
    
    # Set permissions
    chmod 755 ${deb_dir}/usr/bin/*
    chmod 644 ${deb_dir}/etc/systemd/system/divoom-monitor.service
    chmod 644 ${deb_dir}/usr/lib/sysusers.d/divoom.conf
    
    # Build DEB package
    dpkg-deb --build ${deb_dir} ${BUILD_DIR}/packages/divoom-pcmonitor-${VERSION}-${deb_arch}.deb
    
    echo -e "${GREEN}✓ DEB package created: divoom-pcmonitor-${VERSION}-${deb_arch}.deb${NC}"
}

# Function to create RPM package
build_rpm_package() {
    echo -e "${BLUE}Building RPM package...${NC}"
    
    # Create RPM build environment
    local rpm_root="${BUILD_DIR}/packages/rpm"
    mkdir -p ${rpm_root}/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
    
    # Create source tarball
    local src_dir="${rpm_root}/SOURCES/divoom-pcmonitor-${VERSION}"
    mkdir -p ${src_dir}
    
    # Copy source files
    cp *.go ${src_dir}/
    cp go.mod go.sum ${src_dir}/
    cp -r packaging ${src_dir}/
    
    cd ${rpm_root}/SOURCES
    tar -czf divoom-pcmonitor-${VERSION}.tar.gz divoom-pcmonitor-${VERSION}/
    cd - > /dev/null
    
    # Copy spec file
    cp packaging/rpm/divoom-pcmonitor.spec ${rpm_root}/SPECS/
    
    # Update version in spec file
    sed -i "s/Version:        1.0.0/Version:        ${VERSION}/" ${rpm_root}/SPECS/divoom-pcmonitor.spec
    
    # Build RPM
    rpmbuild --define "_topdir ${PWD}/${rpm_root}" -ba ${rpm_root}/SPECS/divoom-pcmonitor.spec
    
    # Copy built RPM
    cp ${rpm_root}/RPMS/x86_64/divoom-pcmonitor-${VERSION}-1.*.rpm ${BUILD_DIR}/packages/
    
    echo -e "${GREEN}✓ RPM package created${NC}"
}

# Function to create Windows installer
build_windows_installer() {
    echo -e "${BLUE}Building Windows installer...${NC}"
    
    local win_dir="${BUILD_DIR}/packages/windows"
    mkdir -p ${win_dir}
    
    # Copy Windows binaries
    cp ${BUILD_DIR}/binaries/divoom-monitor-windows_amd64.exe ${win_dir}/divoom-monitor-windows.exe
    cp ${BUILD_DIR}/binaries/divoom-daemon-windows_amd64.exe ${win_dir}/divoom-daemon-windows.exe
    cp ${BUILD_DIR}/binaries/divoom-test-windows_amd64.exe ${win_dir}/divoom-test-windows.exe
    
    # Copy installer files
    cp packaging/windows/installer.nsi ${win_dir}/
    cp packaging/windows/README.txt ${win_dir}/
    
    # Create a simple icon file (you'd normally have a real .ico file)
    echo "Placeholder for icon.ico" > ${win_dir}/icon.ico
    echo "Placeholder for LICENSE" > ${win_dir}/LICENSE
    
    echo -e "${YELLOW}Windows installer files prepared in ${win_dir}${NC}"
    echo -e "${YELLOW}Run 'makensis installer.nsi' in that directory to create the installer${NC}"
}

# Function to create source archives
create_archives() {
    echo -e "${BLUE}Creating binary archives...${NC}"
    
    cd ${BUILD_DIR}/binaries
    
    # Create archives for each platform
    for os in linux windows darwin freebsd; do
        for arch in amd64 386 arm64 arm; do
            local ext=""
            if [ "$os" = "windows" ]; then
                ext=".exe"
            fi
            
            if ls divoom-*-${os}_${arch}${ext} 1> /dev/null 2>&1; then
                local archive_name="divoom-pcmonitor-${VERSION}-${os}-${arch}"
                
                if [ "$os" = "windows" ]; then
                    zip -j "../archives/${archive_name}.zip" divoom-*-${os}_${arch}${ext}
                else
                    tar -czf "../archives/${archive_name}.tar.gz" divoom-*-${os}_${arch}${ext}
                fi
                
                echo -e "${GREEN}✓ Created archive: ${archive_name}${NC}"
            fi
        done
    done
    
    cd - > /dev/null
}

# Create checksums
create_checksums() {
    echo -e "${BLUE}Creating checksums...${NC}"
    
    cd ${BUILD_DIR}
    find packages archives -type f \( -name "*.deb" -o -name "*.rpm" -o -name "*.tar.gz" -o -name "*.zip" \) -exec sha256sum {} \; > checksums.txt
    cd - > /dev/null
    
    echo -e "${GREEN}✓ Checksums created${NC}"
}

# Check if required tools are available
check_dependencies() {
    echo -e "${YELLOW}Checking build dependencies...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi
    
    if ! command -v dpkg-deb &> /dev/null; then
        echo -e "${YELLOW}Warning: dpkg-deb not available, DEB packages will be skipped${NC}"
        SKIP_DEB=1
    fi
    
    if ! command -v rpmbuild &> /dev/null; then
        echo -e "${YELLOW}Warning: rpmbuild not available, RPM packages will be skipped${NC}"
        SKIP_RPM=1
    fi
    
    echo -e "${GREEN}✓ Dependencies checked${NC}"
}

# Main build process
check_dependencies

# Build packages
if [ -z "$SKIP_DEB" ]; then
    echo -e "${YELLOW}Building DEB packages...${NC}"
    build_deb_package "amd64" "amd64"
    build_deb_package "386" "i386"
    build_deb_package "arm64" "arm64"
    build_deb_package "arm" "armhf"
fi

if [ -z "$SKIP_RPM" ]; then
    echo -e "${YELLOW}Building RPM packages...${NC}"
    build_rpm_package
fi

echo -e "${YELLOW}Building Windows installer files...${NC}"
build_windows_installer

echo -e "${YELLOW}Creating binary archives...${NC}"
create_archives

create_checksums

echo ""
echo -e "${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${GREEN}=======================================\n${NC}"

echo "Built artifacts:"
echo "📦 Packages:"
find ${BUILD_DIR}/packages -name "*.deb" -o -name "*.rpm" | sed 's/^/  /'
echo ""
echo "📁 Binary archives:"
find ${BUILD_DIR}/archives -name "*.tar.gz" -o -name "*.zip" | sed 's/^/  /'
echo ""
echo "🔍 Checksums: ${BUILD_DIR}/checksums.txt"
echo ""

echo "Installation examples:"
echo "  Ubuntu/Debian: sudo dpkg -i ${BUILD_DIR}/packages/divoom-pcmonitor-${VERSION}-amd64.deb"
echo "  RHEL/CentOS:   sudo rpm -i ${BUILD_DIR}/packages/divoom-pcmonitor-${VERSION}-1.*.rpm"
echo "  Manual:        Extract archive and copy binaries to /usr/local/bin/"

echo ""
echo "Service management:"
echo "  Start service: sudo systemctl start divoom-monitor"
echo "  Enable on boot: sudo systemctl enable divoom-monitor"
echo "  Check status: sudo systemctl status divoom-monitor"