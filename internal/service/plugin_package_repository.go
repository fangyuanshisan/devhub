package service

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	pluginregistry "devhub-gin-backend/internal/plugins"

	"devhub-gin-backend/internal/domain"
)

type PluginPackageRepositoryFilter struct {
	Status         string
	Keyword        string
	RiskLevel      string
	ChecksumStatus string
	ManifestValid  string
	Page           int
	PageSize       int
}

func (s *Service) ListPluginPackages(root string, filter PluginPackageRepositoryFilter) (domain.PluginPackageRepositoryListResponse, error) {
	repoAbs, repoClean, err := pluginregistry.NormalizePluginPackageRepositoryRoot(root)
	if err != nil {
		return domain.PluginPackageRepositoryListResponse{}, err
	}

	items, err := pluginregistry.ScanPluginRepository(repoAbs, repoClean)
	if err != nil {
		return domain.PluginPackageRepositoryListResponse{}, err
	}

	statusFilter := strings.TrimSpace(strings.ToLower(filter.Status))
	if statusFilter == "" {
		statusFilter = "all"
	}
	keyword := strings.TrimSpace(strings.ToLower(filter.Keyword))
	riskFilter := strings.TrimSpace(strings.ToLower(filter.RiskLevel))
	checksumFilter := strings.TrimSpace(strings.ToLower(filter.ChecksumStatus))
	manifestValidFilter := strings.TrimSpace(strings.ToLower(filter.ManifestValid))

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	all := make([]domain.PluginPackageRepositoryListItem, 0, len(items))
	summary := domain.PluginPackageRepositorySummary{}

	for _, it := range items {
		manifestPath := filepath.Join(it.AbsPath, "manifest.json")
		readmePath := filepath.Join(it.AbsPath, "README.md")
		checksumPath := filepath.Join(it.AbsPath, "checksums.json")

		row := domain.PluginPackageRepositoryListItem{
			Path:           it.CleanPath,
			ManifestFound:  fileExists(manifestPath),
			ReadmeFound:    fileExists(readmePath),
			ChecksumFound:  fileExists(checksumPath),
			SignatureFound: fileExists(filepath.Join(it.AbsPath, "signature.json")),
			PublisherFound: fileExists(filepath.Join(it.AbsPath, "publisher.json")),
			UpdatedAt:      it.UpdatedAt.Unix(),
		}

		if !row.ManifestFound {
			row.Status = "invalid"
			row.RiskLevel = "blocked"
			row.RiskSummary = "缺少 manifest.json"
			row.Errors = append(row.Errors, "缺少 manifest.json")
		} else {
			// Reuse existing dry-run logic for detail summary.
			res, derr := s.DryRunPluginPackage(it.CleanPath)
			if derr != nil {
				// Non-fatal: treat as invalid entry with reason.
				row.Status = "invalid"
				row.RiskLevel = "blocked"
				row.RiskSummary = "dry-run 失败"
				row.Errors = append(row.Errors, fmt.Sprintf("dry-run 失败：%s", derr.Error()))
			} else {
				row.Code = res.Package.Code
				row.Name = res.Package.Name
				row.Version = res.Package.Version
				row.TotalFiles = res.FileScan.TotalFiles
				row.TotalSize = res.FileScan.TotalSize
				row.Warnings = res.Warnings
				row.Errors = res.Errors

				row.Status = normalizeRepoStatus(res.Status, res.ManifestValidation.Valid)
				row.RiskLevel = res.RiskReport.Level
				row.RiskSummary = res.RiskReport.Summary
				row.ChecksumStatus = res.Checksum.Status
				row.Signature = &res.Signature
				mv := res.ManifestValidation.Valid
				row.ManifestValid = &mv
			}
		}

		// Summary counts.
		summary.Total++
		switch row.Status {
		case "ok":
			summary.OK++
		case "warning":
			summary.Warning++
		case "blocked":
			summary.Blocked++
		default:
			summary.Invalid++
		}

		// Filters.
		if statusFilter != "all" && row.Status != statusFilter {
			continue
		}
		if riskFilter != "" && strings.ToLower(row.RiskLevel) != riskFilter {
			continue
		}
		if checksumFilter != "" && strings.ToLower(row.ChecksumStatus) != checksumFilter {
			continue
		}
		if manifestValidFilter == "true" {
			if row.ManifestValid == nil || !*row.ManifestValid {
				continue
			}
		}
		if manifestValidFilter == "false" {
			if row.ManifestValid != nil && *row.ManifestValid {
				continue
			}
		}
		if keyword != "" {
			hay := strings.ToLower(strings.Join([]string{row.Path, row.Code, row.Name, row.Version}, " "))
			if !strings.Contains(hay, keyword) {
				continue
			}
		}

		all = append(all, row)
	}

	// Stable ordering: updated desc, then path asc.
	sort.Slice(all, func(i, j int) bool {
		if all[i].UpdatedAt == all[j].UpdatedAt {
			return all[i].Path < all[j].Path
		}
		return all[i].UpdatedAt > all[j].UpdatedAt
	})

	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	paged := all[start:end]

	return domain.PluginPackageRepositoryListResponse{
		Items: paged,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Summary: summary,
	}, nil
}

func (s *Service) GetPluginPackageDetail(path string) (domain.PluginPackageDryRunResult, error) {
	// Reuse dry-run end-to-end validation; it already enforces allowlist and never writes.
	res, err := s.DryRunPluginPackage(path)
	if err == nil {
		return res, nil
	}
	if apiErr, ok := err.(*domain.APIError); ok && apiErr != nil {
		if apiErr.Code == "plugin_package_not_found" {
			return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_detail_not_found", "插件包详情不存在").
				WithStatus(404).
				WithDetail("path", path).
				WithSuggestion("请检查 path 是否正确，或先扫描仓库确认该包存在。")
		}
	}
	return domain.PluginPackageDryRunResult{}, err
}

func normalizeRepoStatus(status string, manifestValid bool) string {
	v := strings.TrimSpace(strings.ToLower(status))
	if !manifestValid {
		return "invalid"
	}
	switch v {
	case "blocked":
		return "blocked"
	case "warning":
		return "warning"
	default:
		return "ok"
	}
}
