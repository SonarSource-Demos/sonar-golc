package utils

// LanguageExcludedFromTotalLOC is the language name whose lines of code are
// excluded from report totals to match SonarQube standard behavior.
const LanguageExcludedFromTotalLOC = "JSON"

// NoteExcludedFromTotal is the note shown in reports when JSON is excluded from total LOC.
const NoteExcludedFromTotal = "JSON is excluded from the total to reproduce standard SonarQube behavior."
