package archtest

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

/* ── Core ── */

func Test_Core_No_Direct_Outer_Dependencies(t *testing.T) {
	archtest.Package(t, "weatherApi/internal/weather/...").
		Ignoring("weatherApi/internal/weather/cache/...").
		ShouldNotDependOn(
			"weatherApi/internal/delivery",
			"weatherApi/internal/job",
			"weatherApi/internal/infra",
			"weatherApi/internal/email",
			"weatherApi/internal/config",
		)

	archtest.Package(t, "weatherApi/internal/subscription/...").
		Ignoring("weatherApi/internal/subscription/repo",
			"weatherApi/internal/job").
		ShouldNotDependDirectlyOn(
			"weatherApi/internal/job",
			"weatherApi/internal/infra",
			"gorm.io/gorm",
		)

	archtest.Package(t, "weatherApi/internal/domain").
		ShouldNotDependOn(
			"weatherApi/internal/delivery",
			"weatherApi/internal/email",
			"weatherApi/internal/job",
			"weatherApi/internal/infra",
			"github.com/gin-gonic/gin",
		)

	archtest.Package(t, "weatherApi/internal/token/...").
		ShouldNotDependOn(
			"weatherApi/internal/delivery",
			"weatherApi/internal/job",
			"weatherApi/internal/infra",
			"github.com/gin-gonic/gin",
		)
}

/* ── Inbound ── */

func Test_Inbound_Handlers_No_Infra_DB(t *testing.T) {
	archtest.Package(t, "weatherApi/internal/delivery/handlers/...").
		ShouldNotDependOn(
			"weatherApi/internal/email",
			"weatherApi/internal/infra",
			"gorm.io/gorm",
			"github.com/sendgrid/sendgrid-go",
			"github.com/redis/go-redis/v9",
		)
}

func Test_Job_No_Infra_DB(t *testing.T) {
	archtest.Package(t, "weatherApi/internal/job/...").
		ShouldNotDependOn(
			"weatherApi/internal/infra",
			"gorm.io/gorm",
			"github.com/sendgrid/sendgrid-go",
		)
}

/* ── Outbound ── */

func Test_Outbound_No_Direct_Inbound_Dependencies(t *testing.T) {
	for _, p := range []string{
		"weatherApi/internal/email/...",
		"weatherApi/internal/weather/provider/...",
		"weatherApi/internal/weather/cache/...",
	} {
		archtest.Package(t, p).
			ShouldNotDependOn(
				"weatherApi/internal/delivery",
				"weatherApi/internal/job",
			)
	}

	archtest.Package(t, "weatherApi/internal/subscription/repo").
		ShouldNotDependDirectlyOn(
			"weatherApi/internal/delivery",
			"weatherApi/internal/job",
		)
}

/* ── Infra / Config ── */

func Test_Infra_Isolated(t *testing.T) {
	archtest.Package(t, "weatherApi/internal/infra/...").
		ShouldNotDependOn(
			"weatherApi/internal/delivery",
			"weatherApi/internal/job",
			"weatherApi/internal/email",
			"weatherApi/internal/subscription",
		)
}

func Test_Config_No_Core_Dependencies(t *testing.T) {
	archtest.Package(t, "weatherApi/internal/config").
		ShouldNotDependOn(
			"weatherApi/internal/weather",
			"weatherApi/internal/subscription",
			"weatherApi/internal/token",
			"weatherApi/internal/domain",
		)
}
