package badbuild

func Broken() {
	// syntax error below
	if true {
		// missing closing brace and statement
	}
}
