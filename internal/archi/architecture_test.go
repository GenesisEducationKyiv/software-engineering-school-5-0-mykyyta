package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchitecture(t *testing.T) {
	// Основні правила Clean Architecture
	t.Run("DomainLayerShouldNotDependOnOuter", func(t *testing.T) {
		domainPackages := []string{
			"../weather",
			"../subscription",
			"../token",
		}

		forbiddenImports := []string{
			"weatherApi/internal/infra",
			"weatherApi/internal/delivery",
			"github.com/gin-gonic/gin",
			"github.com/redis/go-redis",
			"gorm.io/gorm",
			"github.com/sendgrid/sendgrid-go",
		}

		for _, pkg := range domainPackages {
			checkPackageImports(t, pkg, forbiddenImports)
		}
	})

	// HTTP handlers не повинні залежати від інфраструктури напряму
	t.Run("DeliveryShouldNotDependOnInfraDirectly", func(t *testing.T) {
		checkPackageImports(t, "../delivery", []string{
			"weatherApi/internal/infra/database",
			"weatherApi/internal/infra/email",
			"weatherApi/internal/infra/redis",
			"gorm.io/gorm",
			"github.com/redis/go-redis",
			"github.com/sendgrid/sendgrid-go",
		})
	})

	// Інфраструктура не повинна залежати від delivery
	t.Run("InfraShouldNotDependOnDelivery", func(t *testing.T) {
		checkPackageImports(t, "../infra", []string{
			"weatherApi/internal/delivery",
			"github.com/gin-gonic/gin",
		})
	})

	// Перевірка циклічних залежностей
	t.Run("ShouldNotHaveCyclicDependencies", func(t *testing.T) {
		// Domain пакети не повинні залежати від delivery
		domainPackages := []string{
			"../weather",
			"../subscription",
			"../token",
		}

		for _, pkg := range domainPackages {
			checkPackageImports(t, pkg, []string{
				"weatherApi/internal/delivery",
			})
		}
	})

	// Перевірка незалежності доменних пакетів
	t.Run("DomainPackagesShouldBeIndependent", func(t *testing.T) {
		checkPackageImports(t, "../weather", []string{
			"weatherApi/internal/subscription",
			"weatherApi/internal/token",
		})

		checkPackageImports(t, "../subscription", []string{
			"weatherApi/internal/weather",
			"weatherApi/internal/token",
		})

		checkPackageImports(t, "../token", []string{
			"weatherApi/internal/weather",
			"weatherApi/internal/subscription",
		})
	})

	// Конфіг не повинен залежати від бізнес-логіки
	t.Run("ConfigShouldNotDependOnBusiness", func(t *testing.T) {
		checkPackageImports(t, "../config", []string{
			"weatherApi/internal/weather",
			"weatherApi/internal/subscription",
			"weatherApi/internal/token",
			"weatherApi/internal/delivery",
			"weatherApi/internal/infra",
		})
	})

	// Перевірка правильності структури пакетів
	t.Run("PackageStructureValidation", func(t *testing.T) {
		requiredPackages := []string{
			"../weather",
			"../subscription",
			"../token",
			"../delivery",
			"../infra",
			"../config",
		}

		for _, pkg := range requiredPackages {
			checkPackageExists(t, pkg)
		}
	})

	// Перевірка відсутності прямих залежностей на зовнішні сервіси в доменах
	t.Run("NoDirect3rdPartyDependenciesInDomain", func(t *testing.T) {
		domainPackages := []string{
			"../weather",
			"../subscription",
			"../token",
		}

		thirdPartyPackages := []string{
			"github.com/sendgrid/sendgrid-go",
			"github.com/redis/go-redis",
			"gorm.io/gorm",
			"github.com/gin-gonic/gin",
		}

		for _, pkg := range domainPackages {
			checkPackageImports(t, pkg, thirdPartyPackages)
		}
	})

	// Email пакет не повинен залежати від HTTP
	t.Run("EmailShouldNotDependOnHTTP", func(t *testing.T) {
		checkPackageImports(t, "../email", []string{
			"github.com/gin-gonic/gin",
			"net/http",
			"weatherApi/internal/delivery",
		})
	})

	// Job пакет не повинен залежати від HTTP
	t.Run("JobShouldNotDependOnHTTP", func(t *testing.T) {
		checkPackageImports(t, "../job", []string{
			"github.com/gin-gonic/gin",
			"weatherApi/internal/delivery",
		})
	})
}

func checkPackageImports(t *testing.T, packagePath string, forbiddenImports []string) {
	t.Helper()

	// Перевіряємо, чи існує пакет
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Logf("Package %s does not exist, skipping import check", packagePath)
		return
	}

	// Парсимо всі Go файли в пакеті та підпапках
	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Обробляємо тільки .go файли, крім тестових
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		src, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Logf("Failed to parse file %s: %v", path, err)
			return nil
		}

		// Перевіряємо імпорти
		for _, imp := range src.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			for _, forbidden := range forbiddenImports {
				if importPath == forbidden || strings.HasPrefix(importPath, forbidden) {
					t.Errorf("Package %s should not import %s (found in %s)",
						packagePath, forbidden, path)
				}
			}
		}

		return nil
	})

	require.NoError(t, err)
}

func checkPackageExists(t *testing.T, packagePath string) {
	t.Helper()

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Errorf("Required package %s does not exist", packagePath)
	}
}

// Додаткові перевірки для специфічних правил
func TestSpecificArchitectureRules(t *testing.T) {
	// Перевірка, що в доменних пакетах немає HTTP статус кодів
	t.Run("DomainShouldNotUseHTTPStatusCodes", func(t *testing.T) {
		domainPackages := []string{
			"../weather",
			"../subscription",
			"../token",
		}

		for _, pkg := range domainPackages {
			checkPackageDoesNotContainText(t, pkg, []string{
				"http.StatusOK",
				"http.StatusBadRequest",
				"http.StatusInternalServerError",
				"StatusOK",
				"StatusBadRequest",
			})
		}
	})

	// Перевірка, що handlers використовують тільки дозволені залежності
	t.Run("HandlersShouldOnlyUseAllowedDependencies", func(t *testing.T) {
		allowedImportPrefixes := []string{
			"weatherApi/internal/app",
			"weatherApi/internal/weather",
			"weatherApi/internal/subscription",
			"weatherApi/internal/token",
			"github.com/gin-gonic/gin",
			"github.com/google/uuid",
			"net/http",
			"context",
			"fmt",
			"strconv",
			"time",
			"errors",
			"log",
		}

		checkPackageOnlyUsesAllowedImports(t, "../delivery/handlers", allowedImportPrefixes)
	})

	// Перевірка, що infra пакети не містять бізнес-логіку
	t.Run("InfraShouldNotContainBusinessLogic", func(t *testing.T) {
		checkPackageImports(t, "../infra", []string{
			"weatherApi/internal/weather",
			"weatherApi/internal/subscription",
			"weatherApi/internal/token",
		})
	})

	// Перевірка, що config не має складних залежностей
	t.Run("ConfigShouldHaveMinimalDependencies", func(t *testing.T) {
		checkPackageImports(t, "../config", []string{
			"github.com/gin-gonic/gin",
			"gorm.io/gorm",
			"github.com/redis/go-redis",
		})
	})
}

func checkPackageDoesNotContainText(t *testing.T, packagePath string, forbiddenTexts []string) {
	t.Helper()

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return
	}

	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileContent := string(content)
		for _, forbidden := range forbiddenTexts {
			if strings.Contains(fileContent, forbidden) {
				t.Errorf("Package %s should not contain %s (found in %s)",
					packagePath, forbidden, path)
			}
		}

		return nil
	})

	require.NoError(t, err)
}

func checkPackageOnlyUsesAllowedImports(t *testing.T, packagePath string, allowedImportPrefixes []string) {
	t.Helper()

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Logf("Package %s does not exist, skipping allowed imports check", packagePath)
		return
	}

	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		src, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Logf("Failed to parse file %s: %v", path, err)
			return nil
		}

		for _, imp := range src.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Пропускаємо стандартні бібліотеки Go
			if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
				continue
			}

			// Пропускаємо стандартні пакети Go
			if strings.HasPrefix(importPath, "context") ||
				strings.HasPrefix(importPath, "fmt") ||
				strings.HasPrefix(importPath, "net/") ||
				strings.HasPrefix(importPath, "time") ||
				strings.HasPrefix(importPath, "errors") ||
				strings.HasPrefix(importPath, "log") ||
				strings.HasPrefix(importPath, "strconv") ||
				strings.HasPrefix(importPath, "strings") ||
				strings.HasPrefix(importPath, "encoding/") {
				continue
			}

			isAllowed := false
			for _, allowedPrefix := range allowedImportPrefixes {
				if strings.HasPrefix(importPath, allowedPrefix) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				t.Errorf("Package %s uses disallowed import %s (found in %s)",
					packagePath, importPath, path)
			}
		}

		return nil
	})

	require.NoError(t, err)
}
