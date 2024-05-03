package assets

import "github.com/emmanuel-colussi-sonarsource/sonar-golc/pkg/goloc/language"

var Languages = language.Languages{
	"ActionScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".as"},
	},
	"Abap": {
		LineComments:      []string{"\""},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".abap", ".ab4", ".flow"},
	},
	"Apex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cls", ".trigger"},
	},
	"C": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".c"},
	},
	"C Header": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".h"},
	},
	"C++": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cpp", ".cc"},
	},
	"C++ Header": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".hh", ".hpp"},
	},
	"COBOL": {
		LineComments:      []string{"*", "/"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".cbl", ".ccp", ".cob", ".cobol", ".cpy"},
	},
	"C#": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".cs"},
	},
	"CSS": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".css"},
	},
	"Golang": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".go"},
	},
	"HTML": {
		LineComments:      []string{},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".html", ".htm", ".cshtml", ".vbhtml", ".aspx", ".ascx", ".rhtml", ".erb", ".shtml", ".shtm", "cmp"},
	},
	"Java": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".java", ".jav"},
	},
	"JavaScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".js", ".jsx", ".jsp", ".jspf"},
	},
	"Kotlin": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".kt", ".kts"},
	},
	"Flex": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".as"},
	},
	"PHP": {
		LineComments:      []string{"//", "#"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".php", ".php3", ".php4", ".php5", ".phtml", ".inc"},
	},
	"Objective-C": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".m"},
	},
	"Oracle PL/SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".pkb"},
	},
	"PL/I": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".pl1"},
	},
	"Python": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"\"\"\"", "\"\"\""}},
		Extensions:        []string{".py"},
	},

	"RPG": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".rpg"},
	},
	"Ruby": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{{"=begin", "=end"}},
		Extensions:        []string{".rb"},
	},
	"Scala": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".scala"},
	},
	"Scss": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".scss"},
	},
	"SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".sql"},
	},
	"Swift": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".swift"},
	},
	"TypeScript": {
		LineComments:      []string{"//"},
		MultiLineComments: [][]string{{"/*", "*/"}},
		Extensions:        []string{".ts", ".tsx"},
	},
	"T-SQL": {
		LineComments:      []string{"--"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".tsql"},
	},
	"Vue": {
		LineComments:      []string{"<!--"},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".vue"},
	},
	"Visual Basic .NET": {
		LineComments:      []string{"'"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".vb"},
	},
	"XML": {
		LineComments:      []string{"<!--"},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".xml", ".XML"},
	},
	"XHTML": {
		LineComments:      []string{"<!--"},
		MultiLineComments: [][]string{{"<!--", "-->"}},
		Extensions:        []string{".xhtml"},
	},
	"YAML": {
		LineComments:      []string{"#"},
		MultiLineComments: [][]string{},
		Extensions:        []string{".yaml", ".yml"},
	},
	"Terraform": {
		LineComments:      []string{},
		MultiLineComments: [][]string{},
		Extensions:        []string{".tf"},
	},
}
