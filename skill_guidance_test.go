package dceskills_test

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/lathe-cli/lathe/pkg/config"
	"github.com/lathe-cli/lathe/pkg/lathe"
	"github.com/lathe-cli/lathe/pkg/runtime"

	dceskills "github.com/DaoCloud/daocloud-skills"
	generated "github.com/DaoCloud/daocloud-skills/internal/generated"
)

// buildRoot wires the dce command tree the same way cmd/dce/main.go does.
func buildRoot(t *testing.T) *cobra.Command {
	t.Helper()
	m, err := config.Load(dceskills.CLIConfig)
	if err != nil {
		t.Fatalf("load cli.yaml: %v", err)
	}
	config.Bind(m)
	root := lathe.NewApp(m)
	if err := generated.MountModules(root); err != nil {
		t.Fatalf("mount modules: %v", err)
	}
	return root
}

// The Module availability guidance tells agents to confirm installed modules by
// running `dce global-management about list-g-product-versions`. If that command
// ever stops existing (API/overlay change, module rename), the guidance becomes
// wrong — fail loudly so the prose is kept in step with the generated surface.
func TestModuleAvailabilityGuidanceCommandExists(t *testing.T) {
	root := buildRoot(t)
	path := []string{"global-management", "about", "list-g-product-versions"}
	if _, ok := runtime.FindCatalogCommand(root, path, runtime.CatalogOptions{}); !ok {
		t.Fatalf("skill guidance references `dce %s` but it is not a real command", strings.Join(path, " "))
	}
}

// The Module availability section lives in internal/skill-include/SKILL.md and is
// appended into the generated skill at codegen time via cli.yaml `skill.include`.
// Assert it actually landed in the committed generated SKILL.md, so a regeneration
// that drops the include (or a broken skill.include wiring) is caught.
func TestGeneratedSkillContainsModuleAvailability(t *testing.T) {
	data, err := os.ReadFile("skills/dce/SKILL.md")
	if err != nil {
		t.Fatalf("read skills/dce/SKILL.md: %v", err)
	}
	skill := string(data)
	for _, want := range []string{
		"## Module availability",
		"dce global-management about list-g-product-versions -o json",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("generated SKILL.md missing %q (is cli.yaml skill.include wired and regenerated?)", want)
		}
	}
}
