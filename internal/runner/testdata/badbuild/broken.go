package badbuild

func Broken() {
	// syntax error below
	if true {
		// intentional syntax error - missing closing brace
	// }
}
