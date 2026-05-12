package tui

// helpView renders the keyboard help screen.
func (m Model) helpView(width int, height int) []string {
	lines := []string{
		"keys",
		"----",
		"/            search text",
		"c            clear search",
		"n / N        next / previous search match",
		"f            add guided include filter",
		"x            add guided exclude filter",
		"F            remove last include filter",
		"X            remove last exclude filter",
		"r            reset search and all filters",
		"p            switch profile",
		"space        pause or resume viewport",
		"a            jump to latest and follow",
		"enter        inspect selected entry",
		"up/down      move selection",
		"pgup/pgdown  move selection faster",
		"home/end     jump to top or bottom",
		"?            toggle help",
		"q            quit",
		"",
		"guided filter examples",
		"----------------------",
		"path wildcard /remote.php/dav/*",
		"user_agent wildcard *kube-probe*",
		"status = 200",
		"status >= 500",
		"method = PROPFIND",
		"remote_user = gi8",
		"time after 2026-05-12T13:14:00Z",
		"time before 2026-05-12T13:15:00Z",
	}

	if len(m.loadedConfigs) > 0 {
		lines = append(lines, "", "loaded config", "-------------")
		lines = append(lines, m.loadedConfigs...)
	}

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}
