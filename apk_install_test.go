package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInstallAPKRequiresValidFileAndOnlineDevice(t *testing.T) {
	apk := filepath.Join(t.TempDir(), "sample.apk")
	if err := os.WriteFile(apk, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, serial, path, devices, output, wantError string
		wantInstall                                    bool
	}{
		{"online", "ABC123", apk, "List of devices attached\nABC123 device\n", "Success", "", true},
		{"offline", "ABC123", apk, "List of devices attached\nABC123 offline\n", "", "not online", false},
		{"wrong device", "OTHER", apk, "List of devices attached\nABC123 device\n", "", "not online", false},
		{"invalid serial", "ABC 123", apk, "", "", "invalid device serial", false},
		{"missing APK", "ABC123", filepath.Join(t.TempDir(), "missing.apk"), "", "", "access APK", false},
		{"wrong extension", "ABC123", filepath.Join(t.TempDir(), "sample.zip"), "", "", ".apk", false},
		{"install failure", "ABC123", apk, "List of devices attached\nABC123 device\n", "Failure [INSTALL_FAILED]", "INSTALL_FAILED", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			installed := false
			client := &ADBClient{
				run: func(_ context.Context, _ string, args ...string) ([]byte, error) {
					if !reflect.DeepEqual(args, []string{"devices", "-l"}) {
						t.Fatalf("unexpected ADB command: %v", args)
					}
					return []byte(tt.devices), nil
				},
				installRun: func(_ context.Context, _ string, args ...string) ([]byte, error) {
					installed = true
					if !reflect.DeepEqual(args, []string{"-s", tt.serial, "install", tt.path}) {
						t.Fatalf("unexpected install command: %v", args)
					}
					return []byte(tt.output), nil
				},
			}
			err := client.InstallAPK(context.Background(), tt.serial, tt.path)
			if (err == nil) != (tt.wantError == "") || (err != nil && !strings.Contains(err.Error(), tt.wantError)) {
				t.Fatalf("InstallAPK() = %v; want error containing %q", err, tt.wantError)
			}
			if installed != tt.wantInstall {
				t.Fatalf("installed = %v; want %v", installed, tt.wantInstall)
			}
		})
	}
}
