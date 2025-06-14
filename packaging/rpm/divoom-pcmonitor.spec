Name:           divoom-pcmonitor
Version:        1.0.0
Release:        1%{?dist}
Summary:        PC monitoring tool for Divoom devices

License:        MIT
URL:            https://github.com/alessio/DivoomPCMonitorTool-Linux
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.19
Requires:       systemd
Requires(pre):  shadow-utils
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
DivoomPCMonitorTool provides real-time PC performance monitoring
that displays CPU usage, temperature, memory usage, and more on
compatible Divoom pixel display devices.

This package includes:
- Interactive monitoring application (divoom-monitor)
- Automatic background daemon (divoom-daemon)
- Device testing utility (divoom-test)
- Systemd service for background monitoring

%prep
%autosetup

%build
go build -ldflags="-s -w -X main.version=%{version}" -o divoom-monitor main.go
go build -ldflags="-s -w -X main.version=%{version}" -o divoom-daemon divoom_daemon.go
go build -ldflags="-s -w -X main.version=%{version}" -o divoom-test test_divoom.go

%install
rm -rf $RPM_BUILD_ROOT

# Install binaries
mkdir -p $RPM_BUILD_ROOT%{_bindir}
install -m 0755 divoom-monitor $RPM_BUILD_ROOT%{_bindir}/
install -m 0755 divoom-daemon $RPM_BUILD_ROOT%{_bindir}/
install -m 0755 divoom-test $RPM_BUILD_ROOT%{_bindir}/

# Install systemd service
mkdir -p $RPM_BUILD_ROOT%{_unitdir}
install -m 0644 packaging/systemd/divoom-monitor.service $RPM_BUILD_ROOT%{_unitdir}/

# Install sysusers configuration
mkdir -p $RPM_BUILD_ROOT%{_sysusersdir}
install -m 0644 packaging/systemd/divoom-user.conf $RPM_BUILD_ROOT%{_sysusersdir}/divoom.conf

%pre
getent group divoom >/dev/null || groupadd -r divoom
getent passwd divoom >/dev/null || \
    useradd -r -g divoom -d /var/lib/divoom -s /sbin/nologin \
    -c "Divoom Monitor Service" divoom
exit 0

%post
%systemd_post divoom-monitor.service
mkdir -p /var/lib/divoom
chown divoom:divoom /var/lib/divoom

echo "DivoomPCMonitorTool installed successfully!"
echo ""
echo "To start the monitoring service:"
echo "  sudo systemctl start divoom-monitor"
echo ""
echo "To check service status:"
echo "  sudo systemctl status divoom-monitor"

%preun
%systemd_preun divoom-monitor.service

%postun
%systemd_postun_with_restart divoom-monitor.service
if [ $1 -eq 0 ]; then
    # Package removal, not upgrade
    userdel divoom 2>/dev/null || :
    rm -rf /var/lib/divoom 2>/dev/null || :
fi

%files
%{_bindir}/divoom-monitor
%{_bindir}/divoom-daemon
%{_bindir}/divoom-test
%{_unitdir}/divoom-monitor.service
%{_sysusersdir}/divoom.conf

%changelog
* Sat Jun 14 2025 DivoomPCMonitorTool Team <noreply@example.com> - 1.0.0-1
- Initial RPM package
- Added systemd daemon support
- Cross-platform monitoring support