package gopkg

import (
	"fmt"
	"testing"
)

func TestGoModFile(t *testing.T) {
	fmt.Println(GoModFile.Module.Mod)
	fmt.Println(GoModName())
}
