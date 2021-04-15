package cmd

// func SetupDefaultCommand(defaultCommand string) {
// 	commands := nonRootSubCmds()
// 	fmt.Printf("Commands: %#v\n", commands)

// 	givenCommand := ""
// 	if len(os.Args) > 1 {
// 		givenCommand = os.Args[1]
// 	}

// 	for _, v := range commands {
// 		if givenCommand == v {
// 			fmt.Printf("DC %s is %s\n", defaultCommand, v)
// 			return
// 		}
// 	}

// 	fmt.Println("Adding default")
// 	os.Args = append([]string{os.Args[0], defaultCommand}, os.Args[1:]...)
// }

// func nonRootSubCmds() (l []string) {
// l = append(l, "help")
// 	for _, c := range rootCmd.Commands() {
// 		// isAlias := false
// 		// append(c.Aliases, c.Name())
// 		// for _, a := range append(c.Aliases, c.Name()) {
// 		// 	if a == rootCmd.Aliases[0] {
// 		// 		isAlias = true
// 		// 		break
// 		// 	}
// 		// }
// 		// if !isAlias {
// 		l = append(l, c.Name())
// 		l = append(l, c.Aliases...)
// 		// }
// 	}
// 	return
// }
