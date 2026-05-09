package ps

import (
	"strings"
	"testing"
)

var rawGetHelp = `
NAME
    Get-ChildItem

SYNOPSIS
    Gets the items and child items in one or more specified locations.

SYNTAX
    Get-ChildItem [[-Path] <String[]>] [[-Filter] <String>] [-Attributes <FlagsExpression[FileAttributes]>] [-Depth <UInt32>] [-Directory] [-Exclude <String[]>] [-File] [-Force] [-Hidden] [-Include <String[]>] [-LiteralPath <String[]>] [-Name] [-Recurse] [-UseTransaction] [<CommonParameters>]

DESCRIPTION
    The Get-ChildItem cmdlet gets the items in one or more specified locations. If the item is a container, it gets the items inside the container, known as child items. You can use the Recurse parameter to get items in all child containers and use the Depth parameter to limit the number of levels to recurse.

    Get-ChildItem does not display empty directories. When a Get-ChildItem command includes the Depth or Recurse parameters, empty directories are not included in the output.

    Locations are exposed to Get-ChildItem by PowerShell providers. A location can be a file system directory, registry hive, or a certificate store. For more information, see about_Providers.

PARAMETERS
    -Attributes <FlagsExpression[FileAttributes]>
        Specifies attributes to look for. This parameter supports all attributes and lets you specify complex combinations of attributes.

        For example, to get non-system files (not directories) that are encrypted or compressed, type:
        Get-ChildItem -Attributes !Directory+!System+Encrypted, !Directory+!System+Compressed

    -Depth <UInt32>
        The Depth parameter is added in Windows PowerShell 5.0. It allows you to control the depth of recursion.

    -Directory [<SwitchParameter>]
        To get a list of directories, use the -Directory parameter or the -Attributes parameter with the Directory property. You can use the -Recurse parameter with -Directory.

INPUTS
    System.String
        You can pipe a string that contains a path to Get-ChildItem.

OUTPUTS
    System.IO.FileInfo, System.IO.DirectoryInfo, Microsoft.Win32.RegistryKey
        The type of object returned depends on the objects in the drive path.

NOTES
    Get-ChildItem does not get hidden items by default. To get hidden items, use the -Force parameter.

RELATED LINKS
    Online Version: http://go.microsoft.com/fwlink/?LinkID=113308
    about_Certificate_Provider
    about_Environment_Provider

EXAMPLES
    Example 1: Get child items from the file system directory

    PS C:\> Get-ChildItem -Path C:\Windows\System32

    Example 2: Get all files with specific extension in multiple directories

    PS C:\> Get-ChildItem -Path C:\Windows\*.dll -Recurse -ErrorAction SilentlyContinue | Select-Object FullName

`

func TestExtractSynopsis_basic(t *testing.T) {
	out := ExtractSynopsis(rawGetHelp)
	if !strings.Contains(out, "Gets the items and child items") {
		t.Errorf("synopsis text should appear, got:\n%s", out)
	}
	if strings.Contains(out, "DESCRIPTION") {
		t.Errorf("DESCRIPTION section should not appear, got:\n%s", out)
	}
	if strings.Contains(out, "PARAMETERS") {
		t.Errorf("PARAMETERS section should not appear, got:\n%s", out)
	}
	if strings.Contains(out, "EXAMPLES") {
		t.Errorf("EXAMPLES section should not appear, got:\n%s", out)
	}
	if strings.Contains(out, "RELATED LINKS") {
		t.Errorf("RELATED LINKS should not appear, got:\n%s", out)
	}
}

func TestExtractSynopsis_fallback(t *testing.T) {
	// No recognized sections — fallback to first 3 non-empty lines
	raw := `Get-Process
Gets running processes.
Usage: Get-Process [-Name <string>]
More details here about parameters and usage.
`
	out := ExtractSynopsis(raw)
	if strings.TrimSpace(out) == "" {
		t.Errorf("fallback should return first lines, got empty")
	}
}

func TestExtractSynopsisAndExamples_basic(t *testing.T) {
	out := ExtractSynopsisAndExamples(rawGetHelp, 2)
	if !strings.Contains(out, "Gets the items and child items") {
		t.Errorf("synopsis should appear, got:\n%s", out)
	}
	if !strings.Contains(out, "Example 1") {
		t.Errorf("examples should appear, got:\n%s", out)
	}
	if strings.Contains(out, "NOTES") {
		t.Errorf("NOTES should not appear, got:\n%s", out)
	}
}

func TestExtractSynopsis_token_savings(t *testing.T) {
	out := ExtractSynopsis(rawGetHelp)
	inTok := countTokens(rawGetHelp)
	outTok := countTokens(out)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	if savings < 80 {
		t.Errorf("expected ≥80%% token savings (synopsis only vs full help), got %.1f%%\nInput: %d tokens\nOutput: %d tokens\nOutput:\n%s",
			savings, inTok, outTok, out)
	}
}

func TestExtractSynopsisAndExamples_token_savings(t *testing.T) {
	out := ExtractSynopsisAndExamples(rawGetHelp, 2)
	inTok := countTokens(rawGetHelp)
	outTok := countTokens(out)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	if savings < 60 {
		t.Errorf("expected ≥60%% savings (synopsis+examples vs full help), got %.1f%%\nInput: %d tokens\nOutput: %d tokens",
			savings, inTok, outTok)
	}
}
