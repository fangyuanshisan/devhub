package service

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestUploadPluginPackageZip_DirectRootOKAndNoInstall(t *testing.T) {
	stagingRoot := cleanupUploadStorage(t)
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := zipDirectoryForUpload(t, "examples/plugins/demo_notice", "")
	res, err := svc.UploadPluginPackageZip("demo_notice.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	if res.UploadID == "" || res.PackagePath == "" {
		t.Fatalf("expected upload id and package path: %#v", res)
	}
	if res.Status != "ok" && res.Status != "warning" {
		t.Fatalf("unexpected status: %q", res.Status)
	}
	if !strings.HasPrefix(res.PackagePath, stagingRoot+"/") {
		t.Fatalf("expected staging package path, got %q", res.PackagePath)
	}
	if _, ok := repo.PluginByCode("demo_notice"); ok {
		t.Fatalf("upload must not install plugin")
	}

	detail, err := svc.GetPluginPackageUpload(res.UploadID)
	if err != nil {
		t.Fatalf("GetPluginPackageUpload: %v", err)
	}
	if detail.PackagePath != res.PackagePath {
		t.Fatalf("unexpected detail path: %q vs %q", detail.PackagePath, res.PackagePath)
	}
}

func TestUploadPluginPackageZip_SingleTopLevelRootOK(t *testing.T) {
	_ = cleanupUploadStorage(t)
	svc := New(store.NewMemoryStore())
	raw := zipDirectoryForUpload(t, "examples/plugins/demo_notice", "demo_notice/")
	res, err := svc.UploadPluginPackageZip("demo_notice.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	if !strings.HasSuffix(res.PackagePath, "/demo_notice") {
		t.Fatalf("expected top-level package dir, got %q", res.PackagePath)
	}
}

func TestUploadPluginPackageZip_BlockedManifestRules(t *testing.T) {
	tests := []struct {
		name string
		fill func(t *testing.T, zw *zip.Writer)
		code string
	}{
		{
			name: "missing_manifest",
			fill: func(t *testing.T, zw *zip.Writer) { addUploadZipFile(t, zw, "README.md", []byte("readme")) },
			code: "plugin_package_zip_manifest_missing",
		},
		{
			name: "multiple_manifests",
			fill: func(t *testing.T, zw *zip.Writer) {
				addUploadZipFile(t, zw, "manifest.json", []byte(`{}`))
				addUploadZipFile(t, zw, "other/manifest.json", []byte(`{}`))
			},
			code: "plugin_package_zip_multiple_manifests",
		},
		{
			name: "zip_slip",
			fill: func(t *testing.T, zw *zip.Writer) { addUploadZipFile(t, zw, "../manifest.json", []byte(`{}`)) },
			code: "plugin_package_zip_slip_detected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = cleanupUploadStorage(t)
			svc := New(store.NewMemoryStore())
			raw := makeUploadZip(t, tt.fill)
			_, err := svc.UploadPluginPackageZip("bad.zip", int64(len(raw)), bytes.NewReader(raw))
			if apiCodeForUpload(err) != tt.code {
				t.Fatalf("unexpected code: %q (%v)", apiCodeForUpload(err), err)
			}
		})
	}
}

func TestUploadPluginPackageZip_RejectsInvalidTypeAndUploadSize(t *testing.T) {
	svc := New(store.NewMemoryStore())
	if _, err := svc.UploadPluginPackageZip("demo.tar", 1, strings.NewReader("x")); apiCodeForUpload(err) != "plugin_package_upload_invalid_type" {
		t.Fatalf("expected invalid type, got %v", err)
	}
	if _, err := svc.UploadPluginPackageZip("demo.zip", 21*1024*1024, strings.NewReader("x")); apiCodeForUpload(err) != "plugin_package_upload_too_large" {
		t.Fatalf("expected too large, got %v", err)
	}
}

func TestUploadPluginPackageZip_DangerousAndChecksumMismatchBlocked(t *testing.T) {
	tests := []struct {
		name string
		src  string
		code string
	}{
		{name: "dangerous_shell", src: "plugins-local/repository-fixtures/dangerous_shell", code: "plugin_package_dangerous_file"},
		{name: "checksum_mismatch", src: "plugins-local/repository-fixtures/checksum_mismatch", code: "plugin_package_checksum_mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = cleanupUploadStorage(t)
			svc := New(store.NewMemoryStore())
			raw := zipDirectoryForUpload(t, tt.src, "")
			res, err := svc.UploadPluginPackageZip(tt.name+".zip", int64(len(raw)), bytes.NewReader(raw))
			if err != nil {
				t.Fatalf("upload should return blocked result, not transport error: %v", err)
			}
			if res.Status != "blocked" || res.RiskReport.Level != "blocked" {
				t.Fatalf("expected blocked, got %#v", res)
			}
			riskCodes := []string{}
			for _, item := range res.RiskReport.Items {
				riskCodes = append(riskCodes, item.Code)
			}
			if !strings.Contains(strings.Join(res.Errors, " "), tt.code) && !strings.Contains(strings.Join(riskCodes, " "), tt.code) {
				t.Fatalf("expected code %s in result: %#v", tt.code, res)
			}
			quarantineRoot := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_QUARANTINE_ROOT"))
			if quarantineRoot == "" {
				quarantineRoot = "storage/plugins/quarantine"
			}
			if !strings.HasPrefix(res.PackagePath, quarantineRoot+"/") {
				t.Fatalf("blocked extracted package should be quarantined, got %q", res.PackagePath)
			}
		})
	}
}

func TestPromotePluginPackageUpload_OKAndTargetExists(t *testing.T) {
	_ = cleanupUploadStorage(t)
	root := mustProjectRoot(t)
	target := filepath.Join(root, "storage", "plugins", "packages", "demo_notice_upload_promote")
	_ = os.RemoveAll(target)
	t.Cleanup(func() { _ = os.RemoveAll(target) })

	svc := New(store.NewMemoryStore())
	raw := zipMinimalPackageForUpload(t, "demo_notice_upload_promote", "")
	res, err := svc.UploadPluginPackageZip("demo_notice_upload_promote.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	promoted, err := svc.PromotePluginPackageUpload(res.UploadID, false)
	if err != nil {
		t.Fatalf("PromotePluginPackageUpload: %v", err)
	}
	if promoted.PackagePath != "storage/plugins/packages/demo_notice_upload_promote" {
		t.Fatalf("unexpected package path: %q", promoted.PackagePath)
	}
	record, ok := svc.repo.PluginPackageUploadByUploadID(res.UploadID)
	if !ok {
		t.Fatalf("expected upload record")
	}
	if record.Status != domain.PluginPackageUploadStatusPromoted || record.PromotedPath != promoted.PackagePath {
		t.Fatalf("expected promoted upload record, got %#v", record)
	}
	if _, err := os.Stat(filepath.Join(target, "manifest.json")); err != nil {
		t.Fatalf("expected promoted manifest: %v", err)
	}
	installDry, err := svc.DryRunPluginPackage(promoted.PackagePath)
	if err != nil {
		t.Fatalf("install dry-run from local repository: %v", err)
	}
	if installDry.DryRunID == "" || strings.HasPrefix(installDry.Package.Path, "storage/plugins/staging/") {
		t.Fatalf("expected local repository install dry-run id, got %#v", installDry)
	}
	repoList, err := svc.ListPluginPackages("", PluginPackageRepositoryFilter{Keyword: "demo_notice_upload_promote", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListPluginPackages: %v", err)
	}
	if repoList.Pagination.Total == 0 {
		t.Fatalf("expected promoted package in local repository list")
	}
	if got := repoList.Items[0].SourceUploadID; got != res.UploadID {
		t.Fatalf("expected source upload id %q, got %q in %#v", res.UploadID, got, repoList.Items[0])
	}
	if repoList.Items[0].PromotedAt == "" {
		t.Fatalf("expected promoted_at in local repository item: %#v", repoList.Items[0])
	}
	if _, err := svc.PromotePluginPackageUpload(res.UploadID, false); apiCodeForUpload(err) != "plugin_package_promote_target_exists" {
		t.Fatalf("expected target exists, got %v", err)
	}
}

func TestPromotePluginPackageUpload_BlockedRejected(t *testing.T) {
	_ = cleanupUploadStorage(t)
	svc := New(store.NewMemoryStore())
	raw := zipDirectoryForUpload(t, "plugins-local/repository-fixtures/dangerous_shell", "")
	res, err := svc.UploadPluginPackageZip("dangerous_shell.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	if _, err := svc.PromotePluginPackageUpload(res.UploadID, false); apiCodeForUpload(err) != "plugin_package_promote_blocked" {
		t.Fatalf("expected promote blocked, got %v", err)
	}
}

func TestPluginPackageUploadLifecycle_RecordApprovalRescanCancelDeleteCleanup(t *testing.T) {
	_ = cleanupUploadStorage(t)
	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := zipMinimalPackageForUpload(t, "demo_notice_lifecycle", "")
	res, err := svc.UploadPluginPackageZipAs(PluginPackageUploadOperator{ID: 7, Name: "admin"}, "demo_notice_lifecycle.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZipAs: %v", err)
	}
	record, ok := repo.PluginPackageUploadByUploadID(res.UploadID)
	if !ok {
		t.Fatalf("expected upload record")
	}
	if record.Status != domain.PluginPackageUploadStatusStaged || record.UploadedBy != 7 || record.PackageCode != "demo_notice_lifecycle" {
		t.Fatalf("unexpected record: %#v", record)
	}
	list, err := svc.ListPluginPackageUploads(domain.PluginPackageUploadFilter{Status: "staged"})
	if err != nil {
		t.Fatalf("ListPluginPackageUploads: %v", err)
	}
	if list.Pagination.Total == 0 || list.Summary.Staged == 0 {
		t.Fatalf("expected staged summary: %#v", list)
	}
	detail, err := svc.GetPluginPackageUploadDetail(res.UploadID)
	if err != nil {
		t.Fatalf("GetPluginPackageUploadDetail: %v", err)
	}
	if len(detail.Actions) == 0 || !uploadTestActionEnabled(detail.Actions, "promote") {
		t.Fatalf("expected promote action: %#v", detail.Actions)
	}
	if _, err := svc.RescanPluginPackageUpload(res.UploadID); err != nil {
		t.Fatalf("RescanPluginPackageUpload: %v", err)
	}
	approved, err := svc.SubmitPluginPackageUploadApproval(PluginApprovalOperator{ID: 8, Name: "reviewer"}, res.UploadID, "import review")
	if err != nil {
		t.Fatalf("SubmitPluginPackageUploadApproval: %v", err)
	}
	if approved.Record.Status != domain.PluginPackageUploadStatusApprovalPending || approved.Record.ApprovalID == 0 {
		t.Fatalf("unexpected approval state: %#v", approved.Record)
	}
	if _, err := svc.PromotePluginPackageUpload(res.UploadID, false); apiCodeForUpload(err) != "plugin_package_upload_invalid_status" {
		t.Fatalf("approval_pending must not promote, got %v", err)
	}
	reviewed, err := svc.ReviewPluginPackageUploadApproval(PluginApprovalOperator{ID: 9, Name: "approver"}, res.UploadID, true, "ok")
	if err != nil {
		t.Fatalf("ReviewPluginPackageUploadApproval: %v", err)
	}
	if reviewed.Record.Status != domain.PluginPackageUploadStatusApproved {
		t.Fatalf("expected approved, got %#v", reviewed.Record)
	}
}

func TestPluginPackageUploadLifecycle_DeletePromotedKeepsRepositoryAndCleanup(t *testing.T) {
	_ = cleanupUploadStorage(t)
	root := mustProjectRoot(t)
	target := filepath.Join(root, "storage", "plugins", "packages", "demo_notice_lifecycle_promote")
	_ = os.RemoveAll(target)
	t.Cleanup(func() { _ = os.RemoveAll(target) })

	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := zipMinimalPackageForUpload(t, "demo_notice_lifecycle_promote", "")
	res, err := svc.UploadPluginPackageZip("demo_notice_lifecycle_promote.zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	if _, err := svc.PromotePluginPackageUpload(res.UploadID, false); err != nil {
		t.Fatalf("PromotePluginPackageUpload: %v", err)
	}
	if _, err := svc.DeletePluginPackageUpload(res.UploadID); err != nil {
		t.Fatalf("DeletePluginPackageUpload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "manifest.json")); err != nil {
		t.Fatalf("delete upload must keep promoted repository package: %v", err)
	}

	raw2 := zipMinimalPackageForUpload(t, "demo_notice_lifecycle_cleanup", "")
	res2, err := svc.UploadPluginPackageZip("demo_notice_lifecycle_cleanup.zip", int64(len(raw2)), bytes.NewReader(raw2))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip cleanup: %v", err)
	}
	record, _ := repo.PluginPackageUploadByUploadID(res2.UploadID)
	record.Status = domain.PluginPackageUploadStatusExpired
	record.ExpiresAt = "2000-01-01 00:00:00"
	_, _ = repo.SavePluginPackageUpload(record)
	cleaned, err := svc.CleanupPluginPackageUploads()
	if err != nil {
		t.Fatalf("CleanupPluginPackageUploads: %v", err)
	}
	if cleaned.Cleaned == 0 {
		t.Fatalf("expected cleaned upload: %#v", cleaned)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("cleanup must not delete package repository: %v", err)
	}
}

func uploadTestActionEnabled(actions []domain.PluginPackageUploadAction, name string) bool {
	for _, action := range actions {
		if action.Action == name {
			return action.Enabled
		}
	}
	return false
}

func makeUploadZip(t *testing.T, fill func(t *testing.T, zw *zip.Writer)) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fill(t, zw)
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func zipDirectoryForUpload(t *testing.T, relDir string, prefix string) []byte {
	t.Helper()
	root := mustProjectRoot(t)
	src := filepath.Join(root, filepath.FromSlash(relDir))
	return makeUploadZip(t, func(t *testing.T, zw *zip.Writer) {
		err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			h, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			h.Name = filepath.ToSlash(filepath.Join(prefix, rel))
			h.Method = zip.Deflate
			w, err := zw.CreateHeader(h)
			if err != nil {
				return err
			}
			_, err = w.Write(raw)
			return err
		})
		if err != nil {
			t.Fatalf("walk zip dir: %v", err)
		}
	})
}

func zipMinimalPackageForUpload(t *testing.T, code string, prefix string) []byte {
	t.Helper()
	manifest := strings.ReplaceAll(`{"code":"__CODE__","name":"Upload Promote","version":"1.0.0","compatible_core_version":">=1.4.0","is_system":false,"content_types":["upload_item"],"content_type_definitions":[{"type":"upload_item","name":"Upload Item","plugin_code":"__CODE__","create_permission":"__CODE__.item.create","edit_permission":"__CODE__.item.edit","delete_permission":"__CODE__.item.delete","audit_permission":"__CODE__.item.audit","default_status":"draft","allow_comment":true,"allow_like":true,"allow_favorite":true,"seo_type":"Article"}],"permissions":[{"code":"__CODE__.item.create","name":"create","scope":"community"},{"code":"__CODE__.item.edit","name":"edit","scope":"own"},{"code":"__CODE__.item.delete","name":"delete","scope":"own"},{"code":"__CODE__.item.audit","name":"audit","scope":"community"}],"menus":[{"code":"__CODE__.admin","title":"Upload","path":"/admin-next/__CODE__","location":"admin","area":"admin","permission":"__CODE__.item.audit"}],"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/__CODE__","handler":"reserved","auth":"admin","permission":"__CODE__.item.audit"}]}`, "__CODE__", code)
	return makeUploadZip(t, func(t *testing.T, zw *zip.Writer) {
		addUploadZipFile(t, zw, filepath.ToSlash(filepath.Join(prefix, "manifest.json")), []byte(manifest))
		addUploadZipFile(t, zw, filepath.ToSlash(filepath.Join(prefix, "README.md")), []byte("# Upload Promote\n"))
		addUploadZipFile(t, zw, filepath.ToSlash(filepath.Join(prefix, "config.example.json")), []byte("{}"))
	})
}

func addUploadZipFile(t *testing.T, zw *zip.Writer, name string, raw []byte) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	if _, err := w.Write(raw); err != nil {
		t.Fatalf("write zip file: %v", err)
	}
}

func cleanupUploadStorage(t *testing.T) string {
	t.Helper()
	root := mustProjectRoot(t)

	// Keep test storage under allowlisted "storage/plugins/packages" to reuse the same
	// plugin package path whitelist as production code, and avoid root-owned dirs in dev env.
	uploads := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-uploads")
	staging := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-staging")
	quarantine := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-quarantine")

	t.Setenv("DEVHUB_PLUGIN_PACKAGE_UPLOADS_ROOT", uploads)
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_STAGING_ROOT", staging)
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_QUARANTINE_ROOT", quarantine)

	for _, rel := range []string{uploads, staging, quarantine} {
		_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel)))
	}
	t.Cleanup(func() {
		for _, rel := range []string{uploads, staging, quarantine} {
			_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel)))
		}
	})
	return staging
}

func apiCodeForUpload(err error) string {
	if api, ok := err.(*domain.APIError); ok && api != nil {
		return api.Code
	}
	return ""
}
