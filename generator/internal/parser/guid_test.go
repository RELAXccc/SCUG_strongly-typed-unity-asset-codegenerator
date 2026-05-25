package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractClassInfo(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "scug_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name              string
		content           string
		filename          string
		expectedClassName string
		expectedPublic    bool
	}{
		{
			name: "Public class with namespace",
			content: `
namespace MyNamespace {
    public class MyClass : MonoBehaviour {}
}`,
			filename:          "MyClass.cs",
			expectedClassName: "MyNamespace.MyClass",
			expectedPublic:    true,
		},
		{
			name: "Internal class without namespace",
			content: `
class MyInternalClass {
}`,
			filename:          "MyInternalClass.cs",
			expectedClassName: "MyInternalClass",
			expectedPublic:    false,
		},
		{
			name: "Multiple classes, pick filename match",
			content: `
public class OtherClass {}
public class TargetClass {}
`,
			filename:          "TargetClass.cs",
			expectedClassName: "TargetClass",
			expectedPublic:    true,
		},
		{
			name: "Protected class (defaults to internal)",
			content: `
protected class ProtectedClass {}
`,
			filename:          "ProtectedClass.cs",
			expectedClassName: "ProtectedClass",
			expectedPublic:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tc.filename)
			if err := ioutil.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			className, isPublic := extractClassInfo(path)
			if className != tc.expectedClassName {
				t.Errorf("extractClassInfo() className = %q, want %q", className, tc.expectedClassName)
			}
			if isPublic != tc.expectedPublic {
				t.Errorf("extractClassInfo() isPublic = %v, want %v", isPublic, tc.expectedPublic)
			}
		})
	}
}
